package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdminReviewFetch(t *testing.T) {
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
			name:         "Missing page param",
			url:          "/admin/review",
			seedData:     func() {},
			wantStatus:   http.StatusBadRequest,
			wantMsg:      "Did not receive page",
			validateData: nil,
		},
		{
			name:         "Invalid dateFrom/dateTo format",
			url:          `/admin/review?page=1&dateFrom="hihihoho"&dateTo="hohohihi"`,
			seedData:     func() {},
			wantStatus:   http.StatusBadRequest,
			wantMsg:      "Invalid format for dateFrom/dateTo (expect yyyy-mm-dd)",
			validateData: nil,
		},
		{
			name:         "Valid date range but dateFrom > dateTo",
			url:          `/admin/review?page=1&dateFrom=2025-09-10&dateTo=2025-01-01`,
			seedData:     func() {},
			wantStatus:   http.StatusBadRequest,
			wantMsg:      "Date to can't be before date from",
			validateData: nil,
		},
		{
			name:       "Search product name - no data",
			url:        "/admin/review?page=1&sort=Product+Name&search=Banana&dateFrom=2025-09-23&dateTo=2025-09-24",
			seedData:   func() {},
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully fetched review values",
			validateData: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].([]interface{})
				assert.Empty(t, data)
			},
		},
		// {
		// 	name:       "Happy path",
		// 	url:        "/admin/product-type?page=1&search=Type+12",
		// 	seedData:   func() { SeedData(t, SeedProductsBasic) },
		// 	wantStatus: http.StatusOK,
		// 	wantMsg:    "Successfully fetched product types values",
		// 	validateData: func(t *testing.T, body map[string]interface{}) {
		// 		data := body["data"].([]interface{})
		// 		assert.NotEmpty(t, data)
		// 		first := data[0].(map[string]interface{})
		// 		assert.Equal(t, "Type 12", first["typeName"])
		// 	},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Seed
			tt.seedData()

			// Request
			req := httptest.NewRequest("GET", tt.url, nil)
			resp, _ := app.Test(req, -1)

			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			var body map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&body)

			assert.Equal(t, tt.wantMsg, body["message"])

			if tt.validateData != nil {
				tt.validateData(t, body)
			}

			// Reset DB
			_, err := testdb.Exec(`TRUNCATE TABLE product, review, review_images, invoice, invoice_detail RESTART IDENTITY CASCADE`)
			assert.NoError(t, err)
		})
	}
}

// func TestAdminProductTypePagination(t *testing.T) {
// 	app := SetupApp()

// 	//Seed data into table product types for pagination
// 	SeedData(t, SeedProductsBasic)

// 	tests := []struct {
// 		name      string
// 		page      int
// 		wantLen   int
// 		wantTotal float64
// 	}{
// 		{
// 			name:      "Page 1 returns 6 products",
// 			page:      1,
// 			wantLen:   6,
// 			wantTotal: 2,
// 		},
// 		{
// 			name:      "Page 2 returns next 6 products",
// 			page:      2,
// 			wantLen:   6,
// 			wantTotal: 2,
// 		},
// 		{
// 			name:      "Page 3 returns empty list",
// 			page:      3,
// 			wantLen:   0,
// 			wantTotal: 2,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			url := fmt.Sprintf("/admin/product-type?page=%d", tt.page)
// 			req := httptest.NewRequest("GET", url, nil)
// 			resp, _ := app.Test(req, -1)

// 			assert.Equal(t, http.StatusOK, resp.StatusCode)

// 			var body map[string]interface{}
// 			_ = json.NewDecoder(resp.Body).Decode(&body)

// 			// Check length of data
// 			data := body["data"].([]interface{})
// 			assert.Len(t, data, tt.wantLen)

// 			// Check totalPage
// 			fmt.Println(body["totalPage"])
// 			assert.Equal(t, tt.wantTotal, body["totalPage"])
// 		})
// 	}
// }

// func TestGetAdminProductTypeDetail(t *testing.T) {
// 	app := SetupApp()

// 	tests := []struct {
// 		name         string
// 		url          string
// 		seed         func()
// 		wantStatus   int
// 		wantMsg      string
// 		validateData func(t *testing.T, body map[string]interface{})
// 	}{
// 		{
// 			name:         "Missing productTypeID",
// 			url:          "/admin/product-type/detail",
// 			seed:         func() {},
// 			wantStatus:   http.StatusBadRequest,
// 			wantMsg:      "Did not receive typeID",
// 			validateData: nil,
// 		},
// 		{
// 			name:         "Product type not found",
// 			url:          "/admin/product-type/detail?typeID=1",
// 			seed:         func() {},
// 			wantStatus:   http.StatusInternalServerError,
// 			wantMsg:      "Product type not found!",
// 			validateData: nil,
// 		},
// 		{
// 			name:       "Happy path",
// 			url:        "/admin/product-type/detail?typeID=1",
// 			seed:       func() { SeedData(t, SeedProductsBasic) },
// 			wantStatus: http.StatusOK,
// 			wantMsg:    "Successfully fetched product types detail",
// 			validateData: func(t *testing.T, body map[string]interface{}) {
// 				data := body["data"].(map[string]interface{})
// 				assert.NotEmpty(t, data)
// 				assert.Equal(t, "Type 1", data["typeName"])
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.seed()

// 			req := httptest.NewRequest("GET", tt.url, nil)
// 			resp, _ := app.Test(req, -1)

// 			assert.Equal(t, tt.wantStatus, resp.StatusCode)

// 			var body map[string]interface{}
// 			_ = json.NewDecoder(resp.Body).Decode(&body)
// 			assert.Equal(t, tt.wantMsg, body["message"])

// 			// Reset DB
// 			_, err := testdb.Exec(`TRUNCATE TABLE product, product_type, product_images RESTART IDENTITY CASCADE`)
// 			assert.NoError(t, err)
// 		})
// 	}
// }

// func TestAdminProductTypeCreate(t *testing.T) {
// 	app := SetupApp()

// 	tests := []struct {
// 		name         string
// 		payload      interface{}
// 		seedData     func()
// 		wantStatus   int
// 		wantMsg      string
// 		wantErr      dto.ProductTypeError
// 		validateData func(t *testing.T, body map[string]interface{})
// 	}{
// 		{
// 			name:         "Invalid body",
// 			payload:      `invalid-json`,
// 			seedData:     func() {},
// 			wantStatus:   http.StatusBadRequest,
// 			wantMsg:      "Invalid request body",
// 			wantErr:      dto.ProductTypeError{},
// 			validateData: nil,
// 		},
// 		{
// 			name: "Validation failed - missing typeName",
// 			payload: models.ProductType{
// 				TypeName: "",
// 				Status:   true,
// 			},
// 			seedData:   func() {},
// 			wantStatus: http.StatusBadRequest,
// 			wantMsg:    "",
// 			wantErr: dto.ProductTypeError{
// 				ErrTypeName: "Please input product type!",
// 			},
// 			validateData: nil,
// 		},
// 		{
// 			name: "Validation failed - product type already exists",
// 			payload: models.ProductType{
// 				TypeName: "Type 1",
// 				Status:   true,
// 			},
// 			seedData:   func() { SeedData(t, SeedProductsBasic) },
// 			wantStatus: http.StatusBadRequest,
// 			wantErr: dto.ProductTypeError{
// 				ErrTypeName: "Product type already exists",
// 			},
// 			validateData: nil,
// 		},
// 		{
// 			name: "Happy path",
// 			payload: models.ProductType{
// 				TypeName: "Type 1",
// 				Status:   true,
// 			},
// 			wantStatus: http.StatusOK,
// 			seedData:   func() {},
// 			wantMsg:    "Successfully created new product type",
// 			wantErr:    dto.ProductTypeError{},
// 			validateData: func(t *testing.T, body map[string]interface{}) {
// 				data := body["data"].(map[string]interface{})
// 				assert.NotEmpty(t, data)
// 				assert.Equal(t, "Type 1", data["typeName"])
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {

// 			tt.seedData()

// 			var req *http.Request
// 			if str, ok := tt.payload.(string); ok {
// 				// invalid JSON
// 				req = httptest.NewRequest("POST", "/admin/product-type/create", strings.NewReader(str))
// 			} else {
// 				b, _ := json.Marshal(tt.payload)
// 				req = httptest.NewRequest("POST", "/admin/product-type/create", bytes.NewReader(b))
// 			}
// 			req.Header.Set("Content-Type", "application/json")

// 			resp, _ := app.Test(req, -1)
// 			assert.Equal(t, tt.wantStatus, resp.StatusCode)

// 			if tt.wantMsg == "" {
// 				var got errorResponseProductType
// 				if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
// 					t.Fatalf("failed to decode body: %v", err)
// 				}

// 				// Compare
// 				assert.Equal(t, tt.wantErr, got.Err)
// 			} else {
// 				var body map[string]interface{}
// 				_ = json.NewDecoder(resp.Body).Decode(&body)
// 				assert.Equal(t, tt.wantMsg, body["message"])
// 				if tt.validateData != nil {
// 					tt.validateData(t, body)
// 				}
// 			}

// 			// Reset DB
// 			_, err := testdb.Exec(`TRUNCATE TABLE product, product_type, product_images RESTART IDENTITY CASCADE`)
// 			assert.NoError(t, err)
// 		})
// 	}
// }
