package integration

import (
	"GoodFood-BE/internal/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/gomail.v2"
)

func TestAdminInvoiceFetch(t *testing.T) {
	app := SetupApp()

	//Table tests setup
	tests := []struct {
		name         string
		url          string
		seedData     func()
		wantStatus   int
		wantMsg      string
		validateData func(t *testing.T, body map[string]interface{})
	}{
		{
			name:         "Missing page param",
			url:          "/admin/order",
			seedData:     func() {},
			wantStatus:   http.StatusBadRequest,
			wantMsg:      "Did not receive page",
			validateData: nil,
		},
		{
			name:         "Invalid dateFrom/dateTo format",
			url:          `/admin/order?page=1&dateFrom="hihihoho"&dateTo="hohohihi"`,
			seedData:     func() {},
			wantStatus:   http.StatusBadRequest,
			wantMsg:      "Invalid format for dateFrom/dateTo (expect yyyy-mm-dd)",
			validateData: nil,
		},
		{
			name:         "Valid date range but dateFrom > dateTo",
			url:          `/admin/order?page=1&dateFrom=2025-09-10&dateTo=2025-01-01`,
			seedData:     func() {},
			wantStatus:   http.StatusBadRequest,
			wantMsg:      "Date to can't be before date from",
			validateData: nil,
		},
		{
			name:       "Search with Invoice ID - no data",
			url:        "/admin/order?page=1&sort=Invoice+ID&search=1",
			seedData:   func() {},
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully fetched invoice values",
			validateData: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].([]interface{})
				assert.Empty(t, data)
			},
		},
		{
			name:       "Search with Invoice ID - has data",
			url:        "/admin/order?page=1&sort=Invoice+ID&search=1",
			seedData:   func() { SeedData(t, SeedAccountWithInvoicesNoDetail) },
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully fetched invoice values",
			validateData: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].([]interface{})
				assert.NotEmpty(t, data)
				first := data[0].(map[string]interface{})
				assert.Equal(t, float64(1), first["invoiceID"])
			},
		},
		{
			name:       "Search with Customer name - no data",
			url:        "/admin/order?page=1&sort=Customer+name&search=Usertest",
			seedData:   func() {},
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully fetched invoice values",
			validateData: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].([]interface{})
				assert.Empty(t, data)
			},
		},
		{
			name:       "Search with Customer name - has data",
			url:        "/admin/order?page=1&sort=Customer+name&search=Usertest",
			seedData:   func() { SeedData(t, SeedAccountWithInvoicesNoDetail) },
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully fetched invoice values",
			validateData: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].([]interface{})
				assert.NotEmpty(t, data)
				first := data[0].(map[string]interface{})
				assert.Equal(t, "Usertest", first["receiveName"])
			},
		},
		{
			name:       "Search with Invoice status - no data",
			url:        "/admin/order?page=1&sort=Invoice+status&search=Order+Placed",
			seedData:   func() {},
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully fetched invoice values",
			validateData: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].([]interface{})
				assert.Empty(t, data)
			},
		},
		{
			name:       "Search with Invoice status - has data",
			url:        "/admin/order?page=1&sort=Invoice+status&search=Order+Placed",
			seedData:   func() { SeedData(t, SeedAccountWithInvoicesNoDetail) },
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully fetched invoice values",
			validateData: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].([]interface{})
				assert.NotEmpty(t, data)
				first := data[0].(map[string]interface{})
				assert.Equal(t, float64(1), first["invoiceStatusID"])
			},
		},
		{
			name:       "Search with Date range - no data",
			url:        "/admin/order?page=1&sort=Created+at&dateFrom=2025-09-09&dateTo=2025-09-10",
			seedData:   func() {},
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully fetched invoice values",
			validateData: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].([]interface{})
				assert.Empty(t, data)
			},
		},
		{
			name:       "Search with Date range - has data",
			url:        "/admin/order?page=1&sort=Created+at&dateFrom=2025-09-01&dateTo=2025-09-05",
			seedData:   func() { SeedData(t, SeedAccountWithInvoicesNoDetail) },
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully fetched invoice values",
			validateData: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].([]interface{})
				assert.NotEmpty(t, data)
				first := data[0].(map[string]interface{})
				t1, _ := time.Parse(time.RFC3339, first["createdAt"].(string))
				assert.Equal(t, "2025-09-05", t1.Format("2006-01-02"))
			},
		},
		{
			name:       "Happy path",
			url:        "/admin/order?page=1&sort=Customer+name&search=Usertest",
			seedData:   func() { SeedData(t, SeedAccountWithInvoicesNoDetail) },
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully fetched invoice values",
			validateData: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].([]interface{})
				assert.NotEmpty(t, data)
				first := data[0].(map[string]interface{})
				assert.Equal(t, first["receiveName"], "Usertest")
			},
		},
	}

	//Run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//Seed data
			tt.seedData()

			//Send requests
			req := httptest.NewRequest("GET", tt.url, nil)
			resp, _ := app.Test(req, -1)

			//Check status
			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			var body map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&body)

			//Check message
			assert.Equal(t, tt.wantMsg, body["message"])

			//Check data
			if tt.validateData != nil {
				tt.validateData(t, body)
			}

			//Reset tables data
			_, err := testdb.Exec(`TRUNCATE TABLE invoice, invoice_detail, account, product, product_type RESTART IDENTITY CASCADE`)
			assert.NoError(t, err)
			
		})
	}
}

func TestAdminInvoice_Pagination(t *testing.T) {
	app := SetupApp()

	//Seed data into table Invoice for pagination
	SeedData(t, SeedAccountWithInvoicesNoDetail)

	tests := []struct {
		name      string
		page      int
		wantLen   int
		wantTotal float64
	}{
		{
			name:      "Page 1 returns 6 invoices",
			page:      1,
			wantLen:   6,
			wantTotal: 2,
		},
		{
			name:      "Page 2 returns next 6 invoices",
			page:      2,
			wantLen:   6,
			wantTotal: 2,
		},
		{
			name:      "Page 3 returns empty list",
			page:      3,
			wantLen:   0,
			wantTotal: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/admin/order?page=%d", tt.page)
			req := httptest.NewRequest("GET", url, nil)
			resp, _ := app.Test(req, -1)

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var body map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&body)

			// Check length of data
			data := body["data"].([]interface{})
			assert.Len(t, data, tt.wantLen)

			// Check totalPage
			fmt.Println(body["totalPage"])
			assert.Equal(t, tt.wantTotal, body["totalPage"])
		})
	}
}

func TestAdminInvoice_Detail(t *testing.T) {
	app := SetupApp()

	tests := []struct {
		name         string
		url          string
		seedData     func()
		wantStatus   int
		wantMsg      string
		validateData func(t *testing.T, body map[string]interface{})
	}{
		{
			name:       "Missing invoiceID",
			url:        "/admin/order/detail",
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Did not receive invoiceID",
		},
		{
			name:       "Invoice not found",
			url:        "/admin/order/detail?invoiceID=10120",
			seedData:   func() {},
			wantStatus: http.StatusInternalServerError,
			wantMsg:    "Invoice not found!",
		},
		{
			name: "Invoice detail no items",
			url:  "/admin/order/detail?invoiceID=1",
			seedData: func() {
				// Seed invoice nhưng không có invoice_detail
				SeedData(t, SeedAccountWithInvoicesNoDetail)
			},
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully fetched invoice detail values",
			validateData: func(t *testing.T, body map[string]interface{}) {
				details := body["listInvoiceDetails"].([]interface{})
				assert.Empty(t, details)
			},
		},
		{
			name: "Happy path",
			url:  "/admin/order/detail?invoiceID=1",
			seedData: func() {
				// Seed đủ invoice, status, và invoice_detail
				SeedData(t, SeedHappyPathInvoice)
			},
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully fetched invoice detail values",
			validateData: func(t *testing.T, body map[string]interface{}) {
				details := body["listInvoiceDetails"].([]interface{})
				assert.NotEmpty(t, details)
				statusList := body["listStatus"].([]interface{})
				assert.NotEmpty(t, statusList)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.seedData()

			req := httptest.NewRequest("GET", tt.url, nil)
			resp, _ := app.Test(req, -1)

			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			var body map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&body)

			assert.Equal(t, tt.wantMsg, body["message"])

			if tt.validateData != nil {
				tt.validateData(t, body)
			}

			// reset DB
			_, err := testdb.Exec(`TRUNCATE TABLE invoice, invoice_detail, invoice_status ,address, ward, district, province, account RESTART IDENTITY CASCADE`)
			assert.NoError(t, err)
		})
	}
}

func TestAdminInvoice_Update(t *testing.T) {
	app := SetupApp()

	tests := []struct {
		name         string
		url          string
		body         string
		seedData     func()
		wantStatus   int
		wantMsg      string
		validateData func(t *testing.T, body map[string]interface{})
	}{
		{
			name:       "Missing invoiceID",
			url:        "/admin/order/update",
			body:       `{}`,
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Did not receive invoiceID",
		},
		{
			name:       "Invalid body",
			url:        "/admin/order/update?invoiceID=1",
			body:       `invalid-json`,
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Invalid body!",
		},
		{
			name:       "Invoice status not found",
			url:        "/admin/order/update?invoiceID=1",
			body:       `{"statusName":"NotExists"}`,
			seedData:   func() {},
			wantStatus: http.StatusInternalServerError,
			wantMsg:    "Invoice status not found!",
		},
		{
			name:       "Invoice not found",
			url:        "/admin/order/update?invoiceID=999",
			body:       `{"statusName":"Order Placed"}`,
			seedData:   func() { SeedData(t, SeedConfig{InvoiceStatuses: true}) },
			wantStatus: http.StatusInternalServerError,
			wantMsg:    "Invoice not found!",
		},
		{
			name:       "Normal update (status increments)",
			url:        "/admin/order/update?invoiceID=1",
			body:       `{"statusName":"Order Confirmed"}`,
			seedData:   func() { SeedData(t, SeedHappyPathInvoice) },
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully updated invoice status!",
			validateData: func(t *testing.T, body map[string]interface{}) {
				invoice := body["invoice"].(map[string]interface{})
				assert.Equal(t, float64(2), invoice["invoiceStatusID"])
			},
		},
		{
			name:       "Delivered → Paid (statusID = 5)",
			url:        "/admin/order/update?invoiceID=5",
			body:       `{"statusName":"Delivered"}`,
			seedData:   func() { SeedData(t, SeedHappyPathInvoice) },
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully updated invoice status!",
			validateData: func(t *testing.T, body map[string]interface{}) {
				invoice := body["invoice"].(map[string]interface{})
				assert.Equal(t, float64(5), invoice["invoiceStatusID"])
				assert.Equal(t, true, invoice["status"])
			},
		},
		{
			name:       "Cancelled order (statusID = 6)",
			url:        "/admin/order/update?invoiceID=6",
			body:       `{"statusName":"Cancelled","cancelReason":"Out of stock"}`,
			seedData:   func() { SeedData(t, SeedHappyPathInvoice) },
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully updated invoice status!",
			validateData: func(t *testing.T, body map[string]interface{}) {
				invoice := body["invoice"].(map[string]interface{})
				assert.Equal(t, float64(6), invoice["invoiceStatusID"])
				assert.Equal(t, "Out of stock", invoice["cancelReason"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// seed data
			tt.seedData()

			req := httptest.NewRequest("PUT", tt.url, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			resp, _ := app.Test(req, -1)

			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			var body map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&body)

			assert.Equal(t, tt.wantMsg, body["message"])

			if tt.validateData != nil {
				tt.validateData(t, body)
			}

			// reset DB
			_, err := testdb.Exec(`TRUNCATE TABLE invoice, invoice_status, account, product, product_type, invoice_detail RESTART IDENTITY CASCADE`)
			assert.NoError(t, err)
		})
	}
}

func TestSendOrderCancelEmail_Integration(t *testing.T) {
	send := gomail.SendFunc(func(from string, to []string, msg io.WriterTo) error {
		var buf bytes.Buffer
		_, err := msg.WriteTo(&buf)
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Integration test") // kiểm tra nội dung email
		return nil
	})

	_ = gomail.NewDialer("localhost", 1025, "", "")
	msg := gomail.NewMessage()
	msg.SetHeader("From", "test@example.com")
	msg.SetHeader("To", "user@example.com")
	msg.SetHeader("Subject", "❌ Order Cancellation Notice")
	msg.SetBody("text/html", utils.BuildCancelEmailBody("Integration test", false))

	err := send("test@example.com", []string{"user@example.com"}, msg)
	assert.NoError(t, err)
}
