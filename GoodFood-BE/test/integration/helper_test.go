package integration

import (
	"GoodFood-BE/internal/server/handlers"
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func SetupApp() *fiber.App {
	app := fiber.New()

	// User
	app.Post("/user/register", handlers.HandleRegister)
	app.Get("/user/login", handlers.HandleLogin)
	app.Post("/user/login/google", handlers.HandleLoginGoogle)
	app.Post("/user/login/facebook", handlers.HandleLoginFacebook)
	app.Post("/user/refresh-token", handlers.RefreshToken)
	app.Put("/user/update", handlers.HandleUpdateAccount)
	app.Post("/user/forgot-password/sendOTP", handlers.HandleForgotPassword)
	app.Get("/user/forgot-password/validate", handlers.ValidateResetToken)
	app.Post("/user/forgot-password/reset", handlers.HandleResetPassword)
	app.Post("/user/contact", handlers.HandleContact)

	// Products
	app.Get("/products/getFeaturings", handlers.GetFour)
	app.Get("/products/getTypes", handlers.GetTypes)
	app.Get("/products", handlers.GetProductsByPage)
	app.Post("/products/classify-image", handlers.ClassifyImage)
	app.Get("/products/detail", handlers.GetDetail)
	app.Get("/products/similar", handlers.GetSimilar)

	// Cart
	app.Get("/cart/fetch", handlers.FetchCart)
	app.Get("/cart", handlers.GetCartDetail)
	app.Post("/cart/modify", handlers.Cart_ModifyQuantity)
	app.Delete("/cart/delete", handlers.DeleteCartItem)
	app.Post("/cart/add", handlers.AddToCart)
	app.Delete("/cart/deleteAll", handlers.DeleteAllItems)

	// Address
	app.Get("/address/fetch", handlers.FetchAddress)
	app.Post("/address/insert", handlers.AddressInsert)
	app.Get("/address/detail", handlers.AddressDetail)
	app.Put("/address/update", handlers.AddressUpdate)
	app.Delete("/address/delete", handlers.AddressDelete)
	app.Get("/address/fill", handlers.AddressFill)
	app.Put("/address/quickChange", handlers.AddressQuickChange)

	// Invoice
	app.Post("/invoice/pay", handlers.InvoicePay)
	app.Post("/invoice/pay/vnpay", handlers.InvoicePayVNPAY)

	// Order history
	app.Get("/order-history", handlers.GetOrderHistory)
	app.Put("/order-history/update", handlers.CancelOrder)
	app.Get("/order-history/details", handlers.GetOrderHistoryDetail)

	// Review
	app.Get("/review", handlers.GetReviewData)
	app.Post("/review/create", handlers.HandleSubmitReview)
	app.Get("/review/detail", handlers.GetReviewDetail)
	app.Put("/review/update", handlers.HandleUpdateReview)

	// Change password
	app.Post("/change-password/submit", handlers.ChangePasswordSubmit)

	// Admin dashboard
	app.Get("/admin/dashboard", handlers.GetDashboard)
	app.Get("/admin/linechart", handlers.GetLineChart)
	app.Get("/admin/piechart", handlers.GetPieChart)
	app.Get("/admin/barchart", handlers.GetBarChart)

	// Admin order
	app.Get("/admin/order", handlers.GetAdminInvoice)
	app.Get("/admin/order/detail", handlers.GetAdminInvoiceDetail)
	app.Put("/admin/order/update", handlers.UpdateInvoice)

	// Admin user
	app.Get("/admin/user", handlers.GetAdminUsers)
	app.Get("/admin/user/detail", handlers.GetAdminUserDetail)
	app.Post("/admin/user/create", handlers.AdminUserCreate)
	app.Put("/admin/user/update", handlers.AdminUserUpdate)

	// Admin product type
	app.Get("/admin/product-type", handlers.GetAdminProductTypes)
	app.Get("/admin/product-type/detail", handlers.GetAdminProductTypeDetail)
	app.Post("/admin/product-type/create", handlers.AdminProductTypeCreate)
	app.Put("/admin/product-type/update", handlers.AdminProductTypeUpdate)

	// Admin product
	app.Get("/admin/product", handlers.GetAdminProducts)
	app.Get("/admin/product/detail", handlers.GetAdminProductDetail)
	app.Post("/admin/product/create", handlers.AdminProductCreate)
	app.Put("/admin/product/update", handlers.AdminProductUpdate)

	// Admin statistics
	app.Get("/admin/statistic", handlers.GetAdminStatistics)

	// Admin review
	app.Get("/admin/review", handlers.GetAdminReview)
	app.Get("/admin/review/review-analysis", handlers.GetAdminReviewAnalysis)
	app.Get("/admin/review/detail", handlers.GetAdminReviewDetail)
	app.Post("/admin/review/reply", handlers.InsertReviewReply)
	app.Put("/admin/review/update", handlers.UpdateReviewReply)

	return app
}

func SeedData(t *testing.T, cfg SeedConfig) {
	//Reset tables data
	_, err := testdb.Exec(`TRUNCATE TABLE address, ward, district, province, account, invoice, invoice_detail, account, product, product_type RESTART IDENTITY CASCADE`)
	assert.NoError(t, err)

	//Seed data for table account
	if cfg.Accounts != nil && cfg.Accounts.seedAccount {
		for i := 0; i < cfg.Accounts.numberOfRecords; i++ {
			_, err := testdb.Exec(`INSERT INTO account (username,password,"phoneNumber",email,"fullName",gender,avatar,status,role,"emailVerified") 
			VALUES($1, 'pwd', $2, $3, 'Test User', true, '', true, true, true)`,
				fmt.Sprintf("user%d", i), fmt.Sprintf("00%d", i), fmt.Sprintf("u%d@gmail.com", i),
			)
			assert.NoError(t, err)
		}
	}
	//Seed data for table province
	if cfg.Provinces {
		_, err = testdb.Exec(`INSERT INTO province ("provinceCode", "provinceName") VALUES (79, 'HCM')`)
		assert.NoError(t, err)
	}
	//Seed data for table district
	if cfg.Districts {
		_, err = testdb.Exec(`INSERT INTO district ("districtCode", "districtName", "provinceID") VALUES (760, 'Q1', 1)`)
		assert.NoError(t, err)
	}
	//Seed data for table ward
	if cfg.Wards {
		_, err = testdb.Exec(`INSERT INTO ward ("wardCode", "wardName", "districtID") VALUES ('26734', 'BN', 1)`)
		assert.NoError(t, err)
	}
	//Seed data for table address
	if cfg.Addresses != nil && cfg.Addresses.seedAddress {
		// Ensure prerequisites are seeded
		if !(cfg.Provinces && cfg.Districts && cfg.Wards && cfg.Accounts.seedAccount) {
			t.Fatal("Cannot seed address without province, district, ward, and account")
		}
		for i := 0; i < cfg.Addresses.numberOfRecords; i++ {
			_, err := testdb.Exec(`
			INSERT INTO address 
			("phoneNumber","fullName",address,"specificAddress",status,"provinceID","districtID","wardID","deleteStatus","accountID","wardCode") 
			VALUES ($1,$2,$3,$4,true,1,1,1,true,1,'26734')`,
				"000",
				fmt.Sprintf("User %d", i+1),
				fmt.Sprintf("Addr %d", i+1),
				fmt.Sprintf("Addr detail %d", i+1),
			)
			assert.NoError(t, err)
		}
	}

	//Seed data for table invoice status
	if cfg.InvoiceStatuses{
		statuses := []string{"Order Placed","Order Confirmed", "Order Processing", "Shipping", "Delivered", "Cancelled"}
        for _, status := range statuses{
            _, err := testdb.Exec(`
                INSERT INTO invoice_status
                ("statusName")
            VALUES($1)`,status)
            assert.NoError(t, err)
        }
	}

	//Seed data for table invoice
	if cfg.Invoices != nil && cfg.Invoices.seedInvoice {
		for i := 0; i < cfg.Invoices.numberOfRecords; i++ {
			_, err := testdb.Exec(`
                INSERT INTO invoice
                ("shippingFee","totalPrice","createdAt","paymentMethod",status,note,"cancelReason","receiveAddress","receiveName","receivePhone","accountID","invoiceStatusID")
            VALUES(12000,150000,$1,true,true,'','','Addr detail','Usertest','0799607411',1,1)`,fmt.Sprintf("2025-09-0%02d",i+1))
            assert.NoError(t, err)
		}
	}

	//Seed data for table product type
	if cfg.ProductTypes != nil && cfg.ProductTypes.seedProductType{
		for i := 0; i < cfg.ProductTypes.numberOfRecords; i++ {
			_, err := testdb.Exec(`
                INSERT INTO product_type
                ("typeName",status)
            VALUES($1,true)`,fmt.Sprintf("Type %d",i+1))
            assert.NoError(t, err)
		}
	}

	//Seed data for table product
	if cfg.Products != nil && cfg.Products.seedProduct{
		for i := 0; i < cfg.Products.numberOfRecords; i++ {
			_, err := testdb.Exec(`
                INSERT INTO product
                ("productName",price,"coverImage",description,status,"insertDate","productTypeID",weight)
            VALUES($1,75000,'test.png','Delicious food',true,'2025-09-12',1,1200)`,fmt.Sprintf("Product %d",i+1))
            assert.NoError(t, err)
		}
	}

	//Seed data for table invoice detail
	if cfg.InvoiceDetails != nil && cfg.InvoiceDetails.seedInvoiceDetail{
		if cfg.Invoices != nil{
			for i := 0; i < cfg.InvoiceDetails.numberOfRecords; i++ {
				_, err := testdb.Exec(`
					INSERT INTO invoice_detail
					(quantity,price,"productID","invoiceID")
				VALUES(3,25000,$1,$2)`,i+1,i+1)
				assert.NoError(t, err)
			}
		}
	}

}

type AccountSeed struct {
	seedAccount     bool
	numberOfRecords int
}

type AddressSeed struct {
	seedAddress     bool
	numberOfRecords int
}

type InvoiceSeed struct {
	seedInvoice     bool
	numberOfRecords int
}

type ProductSeed struct{
	seedProduct bool
	numberOfRecords int
}

type ProductTypeSeed struct{
	seedProductType bool
	numberOfRecords int
}

type InvoiceDetailSeed struct{
	seedInvoiceDetail bool
	numberOfRecords int
}

type SeedConfig struct {
	Accounts  *AccountSeed
	Addresses *AddressSeed
	Invoices  *InvoiceSeed
	InvoiceDetails  *InvoiceDetailSeed
	Products *ProductSeed
	ProductTypes *ProductTypeSeed
	Provinces bool
	Districts bool
	Wards     bool
	InvoiceStatuses bool
}
