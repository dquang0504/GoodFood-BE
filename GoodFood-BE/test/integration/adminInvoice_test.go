package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
		validateData func(t *testing.T, body map[string]interface{})
	}{
		{
			name:       "Missing page param",
			url:        "/admin/order",
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Did not receive page",
			validateData: nil,
		},
		{
			name:       "Invalid dateFrom/dateTo format",
			url:        `/admin/order?page=1&dateFrom="hihihoho"&dateTo="hohohihi"`,
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Invalid format for dateFrom/dateTo (expect yyyy-mm-dd)",
			validateData: nil,
		},
		{
			name: "Valid date range but dateFrom > dateTo",
			url:  `/admin/order?page=1&dateFrom=2025-09-10&dateTo=2025-01-01`,
			seedData: func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Date to can't be before date from",
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
				assert.Empty(t,data);
			},
		},
		{
			name:       "Search with Invoice ID - has data",
			url:        "/admin/order?page=1&sort=Invoice+ID&search=1",
			seedData:   func() {SeedData(t,SeedAccountWithInvoices)},
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully fetched invoice values",
			validateData: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].([]interface{})
				assert.NotEmpty(t,data);
				first := data[0].(map[string]interface{})
				assert.Equal(t,float64(1),first["invoiceID"])
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
				assert.Empty(t,data)
			},
		},
		{
			name:       "Search with Customer name - has data",
			url:        "/admin/order?page=1&sort=Customer+name&search=Usertest",
			seedData:   func() {SeedData(t,SeedAccountWithInvoices)},
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully fetched invoice values",
			validateData: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].([]interface{})
				assert.NotEmpty(t,data);
				first := data[0].(map[string]interface{});
				assert.Equal(t,"Usertest",first["receiveName"]);
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
				assert.Empty(t,data)
			},
		},
		{
			name:       "Search with Invoice status - has data",
			url:        "/admin/order?page=1&sort=Invoice+status&search=Order+Placed",
			seedData:   func() {SeedData(t,SeedAccountWithInvoices)},
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully fetched invoice values",
			validateData: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].([]interface{})
				assert.NotEmpty(t,data);
				first := data[0].(map[string]interface{});
				assert.Equal(t,float64(1),first["invoiceStatusID"]);
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
				assert.Empty(t,data)
			},
		},
		{
			name:       "Search with Date range - has data",
			url:        "/admin/order?page=1&sort=Created+at&dateFrom=2025-09-01&dateTo=2025-09-05",
			seedData:   func() {SeedData(t,SeedAccountWithInvoices)},
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully fetched invoice values",
			validateData: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].([]interface{})
				assert.NotEmpty(t,data);
				first := data[0].(map[string]interface{});
				t1, _ := time.Parse(time.RFC3339, first["createdAt"].(string))
				assert.Equal(t,"2025-09-05",t1.Format("2006-01-02"));
			},
		},
		// {
		// 	name:       "Missing page param",
		// 	url:        "/admin/order",
		// 	seedData:   func() {},
		// 	wantStatus: http.StatusBadRequest,
		// 	wantMsg:    "Did not receive page",
		// 	validateData: nil,
		// },

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
				tt.validateData(t,body);
			}

			//Reset tables data
			_, err := testdb.Exec(`TRUNCATE TABLE address, ward, district, province, account RESTART IDENTITY CASCADE`)
			assert.NoError(t, err)

		})
	}
}