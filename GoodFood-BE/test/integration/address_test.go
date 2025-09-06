package integration

import (
	"GoodFood-BE/internal/server/handlers"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)


func setupApp() *fiber.App {
	app := fiber.New()

	app.Get("/address/fetch", handlers.FetchAddress)
	app.Post("/address/insert", handlers.AddressInsert)
	app.Get("/address/detail", handlers.AddressDetail)
	app.Put("/address/update", handlers.AddressUpdate)
	app.Delete("/address/delete", handlers.AddressDelete)
	app.Get("/address/fill", handlers.AddressFill)
	app.Put("/address/quickChange", handlers.AddressQuickChange)

	return app
}

func TestFetchAddress(t *testing.T) {
	app := setupApp()

	// Cleanup before seeding
	_, err := testdb.Exec(`
		TRUNCATE TABLE public.address,
		                 public.ward,
		                 public.district,
		                 public.province,
		                 public.account
		RESTART IDENTITY CASCADE
	`)
	assert.NoError(t, err)

	// Seed data test
	//mock data for account table
	_, err = testdb.Exec(`
		INSERT INTO public.account (username,password,"phoneNumber",email,"fullName",gender,avatar,status,role,"emailVerified") 
		VALUES ('quang','e10adc3949ba59abbe56e057f20f883e','0799607411','quang@gmail.com','Đặng Duy Quang', true, 'quang.jpg', true, true, true)
	`)
	assert.NoError(t, err)
	//Mock data for ward, district and province tables
	_, err = testdb.Exec(`
		INSERT INTO public.province ("provinceCode", "provinceName")
		VALUES (79, 'Hồ Chí Minh')
	`)
	assert.NoError(t, err)
	_, err = testdb.Exec(`
		INSERT INTO public.district ("districtCode", "districtName", "provinceID")
		VALUES (760, 'Quận 1', 1)
	`)
	assert.NoError(t, err)
	_, err = testdb.Exec(`
		INSERT INTO public.ward ("wardCode", "wardName", "districtID")
		VALUES (26734, 'Phường Bến Nghé', 1)
	`)
	assert.NoError(t, err)

	//Mock data for address table
	_, err = testdb.Exec(`
		INSERT INTO public.address ("phoneNumber","fullName",address,"specificAddress",status,"provinceID","districtID","wardID","deleteStatus","accountID","wardCode") 
		VALUES ('0799607411','Đặng Duy Quang','444/38/25D CMT9','444/38/25D CMT9, Bui Huu Nghia, Binh Thuy, Can Tho',true, 1, 1, 1, true,1, true)
	`)
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/address/fetch?accountID=1317231&page=1", nil)
	resp, _ := app.Test(req, -1)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, err)

	assert.Equal(t, "Success", body["status"])
	assert.Contains(t, body, "data")
	assert.Equal(t, "Successfully fetched addresses", body["message"])
	assert.NotEmpty(t, body["data"])
}
