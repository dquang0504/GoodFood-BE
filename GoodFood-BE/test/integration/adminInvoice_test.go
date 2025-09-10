package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)


func TestAdminInvoiceFetch(t *testing.T) {
	app := SetupApp()

	//Table tests setup
	tests := []struct {
		name       string
		url        string
		seedData   func()
		wantStatus int
		wantMsg    string
		checkData  bool
	}{
		{
			name:       "Missing page param",
			url:        "/admin/order",
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Did not receive page",
			checkData:  false,
		},
		{
			name:       "Invalid dateFrom/dateTo format",
			url:        `/admin/order?page=1&dateFrom="hihihoho"&dateTo="hohohihi"`,
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Invalid format for dateFrom/dateTo (expect yyyy-mm-dd)",
			checkData:  false,
		},
		{
			name: "Valid date range but dateFrom > dateTo",
			url:  `/admin/order?page=1&dateFrom=2025-09-10&dateTo=2025-01-01`,
			seedData: func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Date to can't be before date from",
			checkData:  false,
		},
		{
			name:       "Search with Invoice ID",
			url:        "/admin/order?page=1&sort=Invoice ID&search=1",
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Did not receive page",
			checkData:  false,
		},
		{
			name:       "Missing page param",
			url:        "/admin/order",
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Did not receive page",
			checkData:  false,
		},
		{
			name:       "Missing page param",
			url:        "/admin/order",
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Did not receive page",
			checkData:  false,
		},
		{
			name:       "Missing page param",
			url:        "/admin/order",
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Did not receive page",
			checkData:  false,
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
			if tt.checkData {
				assert.Contains(t, body, "data")
			}

			//Reset tables data
			_, err := testdb.Exec(`TRUNCATE TABLE address, ward, district, province, account RESTART IDENTITY CASCADE`)
			assert.NoError(t, err)

		})
	}
}