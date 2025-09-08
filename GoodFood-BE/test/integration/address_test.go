package integration

import (
	"GoodFood-BE/internal/server/handlers"
	"encoding/json"
	"fmt"
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

		})
	}
}

func TestFetchAddress_Pagination(t *testing.T){
	app := setupApp()

	//Seed 12 records into table Address for pagination test
	seedData(t,&AccountSeed{seedAccount: true,numberOfRecords: 1},true,true,true,&AddressSeed{seedAddress: true,numberOfRecords: 12})

	tests := []struct{
		name string
		page int
		wantLen int
		wantTotal float64
	}{
		{
			name: "Page 1 returns 6 addresses",
			page: 1,
			wantLen: 6,
			wantTotal: 2,
		},
		{
			name: "Page 2 returns next 6 addresses",
			page: 2,
			wantLen: 6,
			wantTotal: 2,
		},
		{
			name: "Page 3 returns empty list",
			page: 3,
			wantLen: 0,
			wantTotal: 2,
		},
	}

	for _,tt := range tests{
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

func seedData(t *testing.T, accountSeed *AccountSeed, withWard bool, withDistrict bool, withProvince bool, addressSeed *AddressSeed) {
	//Reset tables data
	_, err := testdb.Exec(`TRUNCATE TABLE address, ward, district, province, account RESTART IDENTITY CASCADE`)
	assert.NoError(t, err)

	//Seed data for table account
	if accountSeed.seedAccount {
		_, err := testdb.Exec(`INSERT INTO account (username,password,"phoneNumber",email,"fullName",gender,avatar,status,role,"emailVerified") 
		VALUES('user', 'pwd', '000', 'u@mail.com', 'Test User', true, '', true, true, true)`)
		assert.NoError(t, err)
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
		_, err = testdb.Exec(`INSERT INTO ward ("wardCode", "wardName", "districtID") VALUES (26734, 'BN', 1)`)
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
			VALUES ($1,$2,$3,$4,true,1,1,1,true,1,true)`,
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
