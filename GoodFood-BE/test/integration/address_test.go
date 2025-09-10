package integration

import (
	"GoodFood-BE/internal/server/handlers"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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
			name:       "Missing accountID",
			url:        "/address/fetch?page=1",
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Did not receive accountID",
			checkData:  false,
		},
		{
			name:       "Missing page",
			url:        "/address/fetch?accountID=1",
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Did not receive pageNum",
			checkData:  false,
		},
		{
			name: "Account has no address",
			url:  "/address/fetch?accountID=1&page=1",
			seedData: func() {
				seedData(t, &AccountSeed{seedAccount: true, numberOfRecords: 1}, false, false, false, &AddressSeed{seedAddress: false, numberOfRecords: 0})
			},
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully fetched addresses",
			checkData:  true,
		},
		{
			name: "Account has address",
			url:  "/address/fetch?accountID=1&page=1",
			seedData: func() {
				seedData(t, &AccountSeed{seedAccount: true, numberOfRecords: 1}, true, true, true, &AddressSeed{seedAddress: true, numberOfRecords: 1})
			},
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully fetched addresses",
			checkData:  true,
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

func TestFetchAddress_Pagination(t *testing.T) {
	app := setupApp()

	//Seed 12 records into table Address for pagination test
	seedData(t, &AccountSeed{seedAccount: true, numberOfRecords: 1}, true, true, true, &AddressSeed{seedAddress: true, numberOfRecords: 12})

	tests := []struct {
		name      string
		page      int
		wantLen   int
		wantTotal float64
	}{
		{
			name:      "Page 1 returns 6 addresses",
			page:      1,
			wantLen:   6,
			wantTotal: 2,
		},
		{
			name:      "Page 2 returns next 6 addresses",
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
			url := fmt.Sprintf("/address/fetch?accountID=1&page=%d", tt.page)
			req := httptest.NewRequest("GET", url, nil)
			resp, _ := app.Test(req, -1)

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var body map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&body)

			// Check length of data
			data := body["data"].([]interface{})
			assert.Len(t, data, tt.wantLen)

			// Check totalPage
			assert.Equal(t, tt.wantTotal, body["totalPage"])
		})
	}
}

func TestAddressInsert(t *testing.T) {
	app := setupApp()

	tests := []struct {
		name       string
		body       string
		seedData   func()
		wantStatus int
		wantMsg    string
		checkData  bool
	}{
		{
			name:       "Invalid JSON body",
			body:       `{"phoneNumber:"000"`, //missing closing curly brace
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Invalid request body",
			checkData:  false,
		},
		{
			name: "Insert failed due to missing foreign keys",
			body: `{
				"phoneNumber": "000",
				"fullName": "Test User",
				"address": "Addr 1",
				"specificAddress": "Addr detail 1",
				"status": true,
				"provinceID": 1,
				"districtID": 1,
				"wardID": 1,
				"accountID": 1,
				"wardCode": "550300",
				"deleteStatus": false
			}`,
			seedData:   func() {},
			wantStatus: http.StatusInternalServerError,
			wantMsg:    "Couldn't insert new address",
			checkData:  false,
		},
		{
			name: "Insert succeeds",
			body: `{
				"phoneNumber": "000",
				"fullName": "Test User",
				"address": "Addr 1",
				"specificAddress": "Addr detail 1",
				"status": true,
				"provinceID": 1,
				"districtID": 1,
				"wardID": 1,
				"accountID": 1,
				"wardCode": "26734",
				"deleteStatus": false
			}`,
			seedData: func() {
				seedData(t, &AccountSeed{seedAccount: true, numberOfRecords: 1}, true, true, true, &AddressSeed{seedAddress: false, numberOfRecords: 0})
			},
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully inserted new address",
			checkData:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.seedData()

			//Send request
			req := httptest.NewRequest("POST", "/address/insert", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			resp, _ := app.Test(req, -1)
			//Check status
			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			var body map[string]interface{}
			//Parse response
			_ = json.NewDecoder(resp.Body).Decode(&body)

			//Check message
			assert.Equal(t, tt.wantMsg, body["message"])

			if tt.checkData {
				assert.Contains(t, body, "data")
				// assert.Equal(t,"000",body["data"].(map[string]interface{})["phoneNumber"])
			}

			//Reset tables data
			_, err := testdb.Exec(`TRUNCATE TABLE address, ward, district, province, account RESTART IDENTITY CASCADE`)
			assert.NoError(t, err)
		})
	}
}

func TestAddressDetail(t *testing.T) {
	app := setupApp()

	tests := []struct {
		name       string
		url        string
		seedData   func()
		wantStatus int
		wantMsg    string
		checkData  bool
	}{
		{
			name:       "Missing accountID",
			url:        "/address/detail?addressID=1",
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Did not receive accountID",
			checkData:  false,
		},
		{
			name:       "Missing addressID",
			url:        "/address/detail?accountID=1",
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Did not receive addressID",
			checkData:  false,
		},
		{
			name:       "Address does not exist",
			url:        "/address/detail?accountID=1&addressID=1",
			seedData:   func() {},
			wantStatus: http.StatusNotFound,
			wantMsg:    "Address not found",
			checkData:  false,
		},
		{
			name: "Unauthorized access (address belongs to another account)",
			url:  "/address/detail?addressID=1&accountID=2",
			seedData: func() {
				seedData(t, &AccountSeed{seedAccount: true, numberOfRecords: 2}, true, true, true, &AddressSeed{seedAddress: true, numberOfRecords: 2})
			},
			wantStatus: http.StatusForbidden,
			wantMsg:    "Address belongs to another account!",
			checkData:  false,
		},
		{
			name: "Fetch address details successfully",
			url:  "/address/detail?addressID=1&accountID=1",
			seedData: func() {
				seedData(t, &AccountSeed{seedAccount: true, numberOfRecords: 1}, true, true, true, &AddressSeed{seedAddress: true, numberOfRecords: 1})
			},
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully fetched address details",
			checkData:  true,
		},
	}

	//Run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.seedData()
			req := httptest.NewRequest("GET", tt.url, nil)
			resp, _ := app.Test(req, -1)

			var body map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&body)

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			assert.Equal(t, tt.wantMsg, body["message"])

			if tt.checkData {
				assert.Contains(t, body, "data")
			}

			_, err := testdb.Exec(`TRUNCATE TABLE address, ward, district, province, account RESTART IDENTITY CASCADE`)
			assert.NoError(t, err)
		})
	}
}

func TestAddressUpdate(t *testing.T) {
	app := setupApp()

	tests := []struct {
		name       string
		url        string
		body       string
		seedData   func()
		wantStatus int
		wantMsg    string
		checkData  bool
	}{
		{
			name:       "Missing accountID",
			url:        "/address/update?addressID=1",
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Did not receive accountID",
			checkData:  false,
		},
		{
			name:       "Missing addressID",
			url:        "/address/update?accountID=1",
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Did not receive addressID",
			checkData:  false,
		},
		{
			name: "Invalid body request",
			url:  "/address/update?accountID=1&addressID=1",
			body: `{
				phoneNumber: "000"
			}`,
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Invalid body request",
			checkData:  false,
		},
		{
			name: "Address does not exist",
			url:  "/address/update?accountID=1&addressID=1",
			body: `{
				"addressID": 1,
				"phoneNumber": "000",
				"fullName": "Test User",
				"address": "Addr 1",
				"specificAddress": "Addr detail 1",
				"status": true,
				"provinceID": 1,
				"districtID": 1,
				"wardID": 1,
				"accountID": 1,
				"wardCode": "550300",
				"deleteStatus": false
			}`,
			seedData:   func() {},
			wantStatus: http.StatusNotFound,
			wantMsg:    "Address not found",
			checkData:  false,
		},
		{
			name: "Unauthorized access (address belongs to another account)",
			url:  "/address/update?addressID=1&accountID=2",
			body: `{}`,
			seedData: func() {
				seedData(t, &AccountSeed{seedAccount: true, numberOfRecords: 2}, true, true, true, &AddressSeed{seedAddress: true, numberOfRecords: 2})
			},
			wantStatus: http.StatusForbidden,
			wantMsg:    "Address belongs to another account!",
			checkData:  false,
		},
		{
			name: "Updated address successfully",
			url:  "/address/update?addressID=1&accountID=1",
			body: `{
				"addressID": 1,
				"phoneNumber": "000",
				"fullName": "User23444",
				"address": "Addr 23444",
				"specificAddress": "Addr detail 23444",
				"status":true,
				"provinceID":1,
				"districtID":1,
				"wardID":1,
				"wardCode":"26734",
				"deleteStatus": false
			}`,
			seedData: func() {
				seedData(t, &AccountSeed{seedAccount: true, numberOfRecords: 1}, true, true, true, &AddressSeed{seedAddress: true, numberOfRecords: 1})
			},
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully updated the address",
			checkData:  true,
		},
	}

	//Run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.seedData()
			req := httptest.NewRequest("PUT", tt.url, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			resp, _ := app.Test(req, -1)

			var body map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&body)

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			assert.Equal(t, tt.wantMsg, body["message"])

			if tt.checkData {
				assert.Contains(t, body, "data")
				t.Log(body["data"])
			}

			_, err := testdb.Exec(`TRUNCATE TABLE address, ward, district, province, account RESTART IDENTITY CASCADE`)
			assert.NoError(t, err)
		})
	}

}

func TestAddressDelete(t *testing.T) {
	app := setupApp()

	tests := []struct {
		name       string
		url        string
		seedData   func()
		wantStatus int
		wantMsg    string
		checkData  bool
	}{
		{
			name:       "Missing accountID",
			url:        "/address/delete?addressID=1",
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Did not receive accountID",
			checkData:  false,
		},
		{
			name:       "Missing addressID",
			url:        "/address/delete?accountID=1",
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Did not receive addressID",
			checkData:  false,
		},
		{
			name: "Address does not exist",
			url:  "/address/delete?accountID=1&addressID=1",
			seedData:   func() {},
			wantStatus: http.StatusNotFound,
			wantMsg:    "Address not found",
			checkData:  false,
		},
		{
			name: "Unauthorized access (address belongs to another account)",
			url:  "/address/delete?addressID=1&accountID=2",
			seedData: func() {
				seedData(t, &AccountSeed{seedAccount: true, numberOfRecords: 2}, true, true, true, &AddressSeed{seedAddress: true, numberOfRecords: 2})
			},
			wantStatus: http.StatusForbidden,
			wantMsg:    "Address belongs to another account!",
			checkData:  false,
		},
		{
			name: "Deleted address successfully",
			url:  "/address/delete?addressID=1&accountID=1",
			seedData: func() {
				seedData(t, &AccountSeed{seedAccount: true, numberOfRecords: 1}, true, true, true, &AddressSeed{seedAddress: true, numberOfRecords: 1})
			},
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully deleted the address",
			checkData:  true,
		},
	}

	//Run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.seedData()
			req := httptest.NewRequest("DELETE", tt.url, nil)
			req.Header.Set("Content-Type", "application/json")
			resp, _ := app.Test(req, -1)

			var body map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&body)

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			assert.Equal(t, tt.wantMsg, body["message"])

			if tt.checkData {
				assert.Contains(t, body, "data")
				t.Log(body["data"])
			}

			_, err := testdb.Exec(`TRUNCATE TABLE address, ward, district, province, account RESTART IDENTITY CASCADE`)
			assert.NoError(t, err)
		})
	}

}

func TestAddressFill(t *testing.T) {
	app := setupApp()

	tests := []struct {
		name       string
		url        string
		seedData   func()
		wantStatus int
		wantMsg    string
		checkData  bool
	}{
		{
			name:       "Missing accountID",
			url:        "/address/fill?addressID=1",
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Did not receive accountID",
			checkData:  false,
		},
		{
			name: "No default delivery address",
			url:  "/address/fill?accountID=1",
			seedData:   func() {},
			wantStatus: http.StatusInternalServerError,
			wantMsg:    "Please go set your default delivery address!",
			checkData:  false,
		},
		{
			name: "Fetched fill address successfully",
			url:  "/address/fill?accountID=1",
			seedData: func() {
				seedData(t, &AccountSeed{seedAccount: true, numberOfRecords: 1}, true, true, true, &AddressSeed{seedAddress: true, numberOfRecords: 1})
			},
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully fetched the fill address",
			checkData:  true,
		},
	}

	//Run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.seedData()
			req := httptest.NewRequest("GET", tt.url, nil)
			req.Header.Set("Content-Type", "application/json")
			resp, _ := app.Test(req, -1)

			var body map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&body)

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			assert.Equal(t, tt.wantMsg, body["message"])

			if tt.checkData {
				assert.Contains(t, body, "data")
				t.Log(body["data"])
			}

			_, err := testdb.Exec(`TRUNCATE TABLE address, ward, district, province, account RESTART IDENTITY CASCADE`)
			assert.NoError(t, err)
		})
	}

}

func TestAddressQuickChange(t *testing.T) {
	app := setupApp()

	tests := []struct {
		name       string
		url        string
		seedData   func()
		wantStatus int
		wantMsg    string
		checkData  bool
	}{
		{
			name:       "Missing accountID",
			url:        "/address/quickChange?addressID=1&toBeDisabled=2",
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Did not receive accountID",
			checkData:  false,
		},
		{
			name:       "Missing addressID",
			url:        "/address/quickChange?accountID=1&toBeDisabled=2",
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Did not receive addressID",
			checkData:  false,
		},
		{
			name:       "Missing toBeDisabled",
			url:        "/address/quickChange?accountID=1&addressID=1",
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Did not receive toBeDisabled address",
			checkData:  false,
		},
		{
			name:       "To-be-updated address and to-be-disabled address cannot be the same",
			url:        "/address/quickChange?accountID=1&addressID=1&toBeDisabled=1",
			seedData:   func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "addressID and toBeDisabled cannot be the same",
			checkData:  false,
		},
		{
			name: "To-be-updated address does not exist",
			url:  "/address/quickChange?accountID=1&addressID=2&toBeDisabled=1",
			seedData:   func() {seedData(t,&AccountSeed{seedAccount: true,numberOfRecords: 1},true,true,true,&AddressSeed{seedAddress: true,numberOfRecords: 1})},
			wantStatus: http.StatusNotFound,
			wantMsg:    "Cannot find the specified to-be-updated address",
			checkData:  false,
		},
		{
			name: "To-be-updated address does not exist",
			url:  "/address/quickChange?accountID=1&addressID=1&toBeDisabled=2",
			seedData:   func() {seedData(t,&AccountSeed{seedAccount: true,numberOfRecords: 1},true,true,true,&AddressSeed{seedAddress: true,numberOfRecords: 1})},
			wantStatus: http.StatusNotFound,
			wantMsg:    "Cannot find the specified to-be-disabled address",
			checkData:  false,
		},
		{
			name: "Unauthorized access (address belongs to another account)",
			url:  "/address/quickChange?addressID=1&accountID=2&toBeDisabled=2",
			seedData: func() {
				seedData(t, &AccountSeed{seedAccount: true, numberOfRecords: 2}, true, true, true, &AddressSeed{seedAddress: true, numberOfRecords: 2})
			},
			wantStatus: http.StatusForbidden,
			wantMsg:    "Address belongs to another account!",
			checkData:  false,
		},
		{
			name: "Updated address successfully",
			url:  "/address/quickChange?addressID=1&accountID=1&toBeDisabled=2",
			seedData: func() {
				seedData(t, &AccountSeed{seedAccount: true, numberOfRecords: 1}, true, true, true, &AddressSeed{seedAddress: true, numberOfRecords: 2})
			},
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully updated the address",
			checkData:  true,
		},
	}

	//Run test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.seedData()
			req := httptest.NewRequest("PUT", tt.url, nil)
			req.Header.Set("Content-Type", "application/json")
			resp, _ := app.Test(req, -1)

			var body map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&body)

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			assert.Equal(t, tt.wantMsg, body["message"])

			if tt.checkData {
				assert.Contains(t, body, "data")
				t.Log(body["data"])
			}

			_, err := testdb.Exec(`TRUNCATE TABLE address, ward, district, province, account RESTART IDENTITY CASCADE`)
			assert.NoError(t, err)
		})
	}

}

func seedData(t *testing.T, accountSeed *AccountSeed, withWard bool, withDistrict bool, withProvince bool, addressSeed *AddressSeed) {
	//Reset tables data
	_, err := testdb.Exec(`TRUNCATE TABLE address, ward, district, province, account RESTART IDENTITY CASCADE`)
	assert.NoError(t, err)

	//Seed data for table account
	if accountSeed.seedAccount {
		for i := 0; i < accountSeed.numberOfRecords; i++ {
			_, err := testdb.Exec(`INSERT INTO account (username,password,"phoneNumber",email,"fullName",gender,avatar,status,role,"emailVerified") 
			VALUES($1, 'pwd', $2, $3, 'Test User', true, '', true, true, true)`,
				fmt.Sprintf("user%d", i), fmt.Sprintf("00%d", i), fmt.Sprintf("u%d@gmail.com", i),
			)
			assert.NoError(t, err)
		}
	}
	//Seed data for table province
	if withProvince {
		_, err = testdb.Exec(`INSERT INTO province ("provinceCode", "provinceName") VALUES (79, 'HCM')`)
		assert.NoError(t, err)
	}
	//Seed data for table district
	if withDistrict {
		_, err = testdb.Exec(`INSERT INTO district ("districtCode", "districtName", "provinceID") VALUES (760, 'Q1', 1)`)
		assert.NoError(t, err)
	}
	//Seed data for table ward
	if withWard {
		_, err = testdb.Exec(`INSERT INTO ward ("wardCode", "wardName", "districtID") VALUES ('26734', 'BN', 1)`)
		assert.NoError(t, err)
	}
	//Seed data for table address
	if addressSeed.seedAddress {
		// Ensure prerequisites are seeded
		if !(withProvince && withDistrict && withWard && accountSeed.seedAccount) {
			t.Fatal("Cannot seed address without province, district, ward, and account")
		}
		for i := 0; i < addressSeed.numberOfRecords; i++ {
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

}

type AccountSeed struct {
	seedAccount     bool
	numberOfRecords int
}

type AddressSeed struct {
	seedAddress     bool
	numberOfRecords int
}
