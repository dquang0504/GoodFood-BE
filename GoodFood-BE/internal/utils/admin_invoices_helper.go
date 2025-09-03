package utils

import (
	"GoodFood-BE/internal/dto"
	"GoodFood-BE/models"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"gopkg.in/gomail.v2"
)

//These constants define allowed invoice statuses
const(
	StatusOrderPlaced = "Order Placed"
	StatusCancelled = "Cancelled"
	StatusOrderConfirmed = "Order Confirmed"
	StatusProcessing = "Order Processing"
	StatusShipping = "Shipping"
	StatusDelivered = "Delivered"
)

// FetchInvoiceCards gets summary stats (total, canceled).
func FetchInvoiceCards(c *fiber.Ctx) (dto.InvoiceCards, error){
	cards := dto.InvoiceCards{}
	err := queries.Raw(`
		SELECT COALESCE(COUNT("invoiceID"),0) AS total,
		COUNT(CASE WHEN "invoiceStatusID" = 6 THEN 1 END) AS canceled
		FROM invoice
	`).Bind(c.Context(),boil.GetContextDB(),&cards);

	return cards, err
}

// ParseDateRange validates and parses date strings (returns zero time if empty).
func ParseDateRange(dateFromStr string, dateToStr string) (time.Time, time.Time, error){
	var dateFrom, dateTo time.Time
	var err error
	if dateFromStr != ""{
		dateFrom, err = time.Parse("2006-01-02",dateFromStr);
		if err != nil{
			return time.Time{},time.Time{},errors.New("Invalid format for dateFrom (expect yyyy-mm-dd)")
		}
	}
	if dateToStr != ""{
		dateTo, err = time.Parse("2006-01-02",dateToStr);
		if err != nil{
			return time.Time{},time.Time{},errors.New("Invalid format for dateFrom (expect yyyy-mm-dd)")
		}
	}
	return dateFrom,dateTo,nil
}

//BuildInvoiceFilters builds filtering logic for search/sort/date
func BuildInvoiceFilters(c *fiber.Ctx, search, sort string, dateFrom, dateTo time.Time)([]qm.QueryMod,[]qm.QueryMod,error){
	queryMods := []qm.QueryMod{
		qm.Load(models.InvoiceRels.InvoiceStatusIDInvoiceStatus),
		qm.OrderBy("\"invoiceID\" DESC"),
	}
	queryModsTotal := []qm.QueryMod{}

	if search != "" {
		switch sort {
		case "Invoice ID":
			queryMods = append(queryMods, qm.Where("CAST(\"invoiceID\" AS TEXT) ILIKE ?", "%"+search+"%"))
			queryModsTotal = append(queryModsTotal, qm.Where("CAST(\"invoiceID\" AS TEXT) ILIKE ?", "%"+search+"%"))
		case "Customer name":
			queryMods = append(queryMods, qm.Where("\"receiveName\" ILIKE ?", "%"+search+"%"))
			queryModsTotal = append(queryModsTotal, qm.Where("\"receiveName\" ILIKE ?", "%"+search+"%"))
		case "Invoice status":
			status, err := models.InvoiceStatuses(qm.Where("\"statusName\" ILIKE ?", "%"+search+"%")).One(c.Context(), boil.GetContextDB())
			if err != nil {
				return nil,nil,err
			}
			queryMods = append(queryMods, qm.Where("\"invoiceStatusID\" = ?", status.InvoiceStatusID))
			queryModsTotal = append(queryModsTotal, qm.Where("\"invoiceStatusID\" = ?", status.InvoiceStatusID))
		default:
			// fallback
		}
	} else if sort == "Created at" && !dateFrom.IsZero() && !dateTo.IsZero() && dateFrom.Before(dateTo) {
		queryMods = append(queryMods, qm.Where("DATE(\"createdAt\") BETWEEN ? AND ?", dateFrom, dateTo))
		queryModsTotal = append(queryModsTotal, qm.Where("DATE(\"createdAt\") BETWEEN ? AND ?", dateFrom, dateTo))
	}
	
	return queryMods,queryModsTotal, nil
}

func MapInvoices(invoices []*models.Invoice) []dto.InvoiceResponse{
	res := make([]dto.InvoiceResponse, len(invoices));
	for i, invoice := range invoices{
		res[i] = dto.InvoiceResponse{
			Invoice: *invoice,
			InvoiceStatus: invoice.R.InvoiceStatusIDInvoiceStatus,
		}
	}
	return res;
}

// FetchInvoiceAndStatus loads an invoice and its current status.
func FetchInvoiceAndStatus(c *fiber.Ctx, invoiceID int)(*models.Invoice, *models.InvoiceStatus, error){
	invoice, err := models.Invoices(qm.Where("\"invoiceID\" = ?",invoiceID)).One(c.Context(),boil.GetContextDB());
	if err != nil{
		return nil,nil,err
	}
	status, err := models.InvoiceStatuses(qm.Where("\"invoiceStatusID\" = ?",invoice.InvoiceStatusID)).One(c.Context(),boil.GetContextDB());
	if err != nil{
		return nil,nil,err
	}

	return invoice, status, nil
}

// FetchStatusProgression builds the next possible statuses based on current status.
func FetchStatusProgression(c *fiber.Ctx, status *models.InvoiceStatus)([]*models.InvoiceStatus,error){
	var queryMods []qm.QueryMod

	switch(status.StatusName){
		case StatusOrderPlaced:
			queryMods = append(queryMods, qm.Where("\"statusName\" LIKE ? OR \"statusName\" LIKE ? OR \"statusName\" LIKE ?",status.StatusName,"Cancelled","Order Confirmed"))
		case StatusOrderConfirmed:
			queryMods = append(queryMods, qm.Where("\"statusName\" LIKE ? OR \"statusName\" LIKE ?",status.StatusName,"Order Processing"))
		case StatusProcessing:
			queryMods = append(queryMods, qm.Where("\"statusName\" LIKE ? OR \"statusName\" LIKE ?",status.StatusName,"Shipping"))
		case StatusShipping:
			queryMods = append(queryMods, qm.Where("\"statusName\" LIKE ? OR \"statusName\" LIKE ?",status.StatusName,"Delivered"))
		case StatusDelivered:
			queryMods = append(queryMods, qm.Where("\"statusName\" LIKE ?",status.StatusName))
		default:
			// fallback: return only current status
			queryMods = append(queryMods, qm.Where("\"statusName\" LIKE ?",status.StatusName))
	}

	return models.InvoiceStatuses(queryMods...).All(c.Context(),boil.GetContextDB());
}

// FetchInvoiceDetails loads invoice details with joined product and customer info.
func FetchInvoiceDetails(c *fiber.Ctx, invoiceID int)([]*dto.InvoiceDetailResponse, error){
	details := []*dto.InvoiceDetailResponse{}
	err := queries.Raw(`
		SELECT invoice_detail.*,
		product."productName" as food,
		invoice."receiveName" AS name, 
		invoice."receivePhone" AS phone, 
		invoice."receiveAddress" AS address
		FROM invoice 
		INNER JOIN invoice_detail ON invoice."invoiceID" = invoice_detail."invoiceID"
		INNER JOIN product ON invoice_detail."productID" = product."productID"
		WHERE invoice_detail."invoiceID" = $1
	`,invoiceID).Bind(c.Context(),boil.GetContextDB(),&details);
	if err != nil{
		return nil, err
	}

	return details, nil
}

// UpdateInvoiceStatus updates an invoice's status with business rules applied.
func UpdateInvoiceStatus(c *fiber.Ctx, invoiceID int, status dto.UpdateInvoiceStruct)(*models.Invoice,error){
	//Find status record
	invoiceStatus, err := models.InvoiceStatuses(qm.Where("\"statusName\" LIKE ?",status.StatusName)).One(c.Context(),boil.GetContextDB())
	if err != nil{
		return nil, err
	}

	//Find invoice
	invoice,err := models.Invoices(
		qm.Where("\"invoiceID\" = ?",invoiceID),
		qm.Load(models.InvoiceRels.AccountIDAccount),
	).One(c.Context(),boil.GetContextDB())
	if err != nil{
		return nil,err
	}

	//Apply business logicc
	if invoice.InvoiceStatusID != invoiceStatus.InvoiceStatusID && invoiceStatus.InvoiceStatusID != 6{
		invoice.InvoiceStatusID += 1
	}
	//invoiceStatusID reaches 5 meaning the order has been delivered, meaning invoice status is paid.
	if invoiceStatus.InvoiceStatusID == 5{
		invoice.InvoiceStatusID = 5
		invoice.Status = true;
	}
	//invoiceStatusID reaches 6 meaning the order has been cancelled, also sends an email clarifying the
	//cancelation
	if invoiceStatus.InvoiceStatusID == 6{
		invoice.InvoiceStatusID = 6
		invoice.CancelReason = status.CancelReason
		err := SendOrderCancelEmail(invoice.R.AccountIDAccount.Email,invoice.CancelReason.String,invoice.Status);
		if err != nil{
			return nil, err
		}
	}
	
	//Update DB
	_,err = invoice.Update(c.Context(),boil.GetContextDB(),boil.Infer())
	if err != nil{
		return nil, err
	}

	return invoice, nil
}

// SendOrderCancelEmail sends an order cancellation email from Admin to customer.
// If isPaid = true, the message will include refund instructions.
func SendOrderCancelEmail(toEmail string, reason string, isPaid bool) error {
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", "williamdang0404@gmail.com")
	mailer.SetHeader("To", toEmail)
	mailer.SetHeader("Subject", "❌ Order Cancellation Notice from GoodFood24h")

	// Common intro
	intro := `
		<p>Hello,</p>
		<p>We regret to inform you that your order has been cancelled by <strong>GoodFood24h</strong>.</p>
	`

	// Common reason section
	reasonSection := fmt.Sprintf(`
		<p><strong>Reason for cancellation:</strong></p>
		<p style="background: #f9f9f9; padding: 10px; border-left: 4px solid #F44336; white-space: pre-line;">
			%s
		</p>
	`, reason)

	// Message based on payment status
	var extra string
	if isPaid {
		extra = `
			<p>Since your order was already paid online, please contact our support hotline 
			at <strong>0799607411</strong> for detailed instructions regarding your refund.</p>
		`
	} else {
		extra = `
			<p>As your order was set to Cash on Delivery (COD), no payment has been made. 
			We sincerely apologize for the inconvenience caused.</p>
		`
	}

	// Final email body
	emailBody := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; color: #333; padding: 20px; max-width: 600px; margin: auto; border: 1px solid #ddd; border-radius: 8px;">
			<h2 style="color: #F44336;">❌ Your order has been cancelled</h2>
			%s
			%s
			%s
			<hr style="margin: 30px 0; border: none; border-top: 1px solid #eee;">
			<p style="font-size: 14px;">If you have any other questions, please feel free to contact us at <a href="mailto:williamdang0404@gmail.com">williamdang0404@gmail.com</a>.</p>
			<p style="font-size: 13px; color: #888;">This email was sent automatically from the GoodFood24h system.</p>
		</div>
	`, intro, reasonSection, extra)

	mailer.SetBody("text/html", emailBody)

	dialer := gomail.NewDialer("smtp.gmail.com", 587, "williamdang0404@gmail.com", "yhjd uzhk hhvp zfiq")

	err := dialer.DialAndSend(mailer)
	return err
}