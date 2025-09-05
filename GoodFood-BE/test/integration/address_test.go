package integration

import (
	"GoodFood-BE/internal/server/handlers"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupApp() *fiber.App{
	app := fiber.New()

	boil.SetDB(db);

	app.Get("/address/fetch",handlers.FetchAddress)
	app.Post("/address/insert",handlers.AddressInsert)
	app.Get("/address/detail",handlers.AddressDetail)
	app.Put("/address/update",handlers.AddressUpdate)
	app.Delete("/address/delete",handlers.AddressDelete)
	app.Get("/address/fill",handlers.AddressFill)
	app.Put("/address/quickChange",handlers.AddressQuickChange)

	return app;
}

func TestFetchAddress(t *testing.T) {
	app := setupApp()

	// Seed data test
	_, err := db.Exec(`
		INSERT INTO public.address ("phoneNumber","fullName","accountID", "specificAddress", status) 
		VALUES ('0799607411','Đặng Duy Quang',123, 'Test Street', true)
	`)
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/address/fetch?accountID=123&page=1", nil)
	resp, _ := app.Test(req, -1)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, err)

	assert.Equal(t, "Success", body["status"])
	assert.Contains(t, body, "data")
	assert.Equal(t, "Successfully fetched addresses", body["message"])
}