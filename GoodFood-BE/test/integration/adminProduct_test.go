package integration

import (
	"GoodFood-BE/internal/dto"
	"GoodFood-BE/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdminProductFetch(t *testing.T) {
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
			url:          "/admin/product",
			seedData:     func() {},
			wantStatus:   http.StatusBadRequest,
			wantMsg:      "Did not receive page",
			validateData: nil,
		},
		{
			name:       "Search product by name - no data",
			url:        "/admin/product?page=1&sort=Product+name&search=Banana",
			seedData:   func() {},
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully fetched products values",
			validateData: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].([]interface{})
				assert.Empty(t, data)
			},
		},
		{
			name:       "Search product by name - has data",
			url:        "/admin/product?page=1&sort=Product+name&search=Product+12",
			seedData:   func() { SeedData(t, SeedProductsBasic) },
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully fetched products values",
			validateData: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].([]interface{})
				assert.NotEmpty(t, data)
				first := data[0].(map[string]interface{})
				assert.Equal(t, "Product 12", first["productName"])
			},
		},
		{
			name:       "Search product by type - has data",
			url:        "/admin/product?page=1&sort=Product+type&search=Type+1",
			seedData:   func() { SeedData(t, SeedProductsBasic) },
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully fetched products values",
			validateData: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].([]interface{})
				assert.NotEmpty(t, data)
				first := data[0].(map[string]interface{})
				assert.Equal(t, "Type 1", first["productType"].(map[string]interface{})["typeName"])
			},
		},
		{
			name:       "Happy path",
			url:        "/admin/product?page=1&sort=Product+name&search=Product+12",
			seedData:   func() { SeedData(t, SeedProductsBasic) },
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully fetched products values",
			validateData: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].([]interface{})
				assert.NotEmpty(t, data)
			},
		},
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
			_, err := testdb.Exec(`TRUNCATE TABLE product, product_type, product_images RESTART IDENTITY CASCADE`)
			assert.NoError(t, err)
		})
	}
}

func TestAdminProduct_Pagination(t *testing.T) {
	app := SetupApp()

	//Seed data into table product for pagination
	SeedData(t, SeedProductsBasic)

	tests := []struct {
		name      string
		page      int
		wantLen   int
		wantTotal float64
	}{
		{
			name:      "Page 1 returns 6 products",
			page:      1,
			wantLen:   6,
			wantTotal: 2,
		},
		{
			name:      "Page 2 returns next 6 products",
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
			url := fmt.Sprintf("/admin/product?page=%d", tt.page)
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

func TestGetAdminProductDetail(t *testing.T) {
	app := SetupApp()

	tests := []struct {
		name         string
		url          string
		seed         func()
		wantStatus   int
		wantMsg      string
		validateData func(t *testing.T, body map[string]interface{})
	}{
		{
			name:         "Missing productID",
			url:          "/admin/product/detail",
			seed:         func() {},
			wantStatus:   http.StatusBadRequest,
			wantMsg:      "Did not receive productID!",
			validateData: nil,
		},
		{
			name:         "Product not found",
			url:          "/admin/product/detail?productID=1",
			seed:         func() {},
			wantStatus:   http.StatusInternalServerError,
			wantMsg:      "Product not found!",
			validateData: nil,
		},
		{
			name:       "Happy path",
			url:        "/admin/product/detail?productID=1",
			seed:       func() { SeedData(t, SeedProductsBasic) },
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully fetched products detail",
			validateData: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].([]interface{})
				assert.NotEmpty(t, data)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.seed()

			req := httptest.NewRequest("GET", tt.url, nil)
			resp, _ := app.Test(req, -1)

			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			var body map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&body)
			assert.Equal(t, tt.wantMsg, body["message"])

			// Reset DB
			_, err := testdb.Exec(`TRUNCATE TABLE product, product_type, product_images RESTART IDENTITY CASCADE`)
			assert.NoError(t, err)
		})
	}
}

func TestAdminProductCreate(t *testing.T) {
	app := SetupApp()

	tests := []struct {
		name       string
		payload    interface{}
		seedData func()
		wantStatus int
		wantMsg    string
		wantErr    dto.ProductError
		validateData func(t *testing.T, body map[string]interface{})
	}{
		{
			name:       "Invalid body",
			payload:    `invalid-json`,
			seedData: func(){},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Invalid request body",
			wantErr: dto.ProductError{},
			validateData: nil,
		},
		{
			name: "Validation failed - missing productName",
			payload: dto.ProductResponse{
				Product:     models.Product{Price: 10, Weight: 1},
				ProductType: models.ProductType{TypeName: "Fruit"},
				ProductImages: []models.ProductImage{
					{Image: "url1.jpg"},
				},
			},
			seedData: func(){},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "",
			wantErr: dto.ProductError{
				ErrProductName: "Please input product name!",
			},
			validateData: nil,
		},
		{
			name: "Validation failed - negative price",
			payload: dto.ProductResponse{
				Product: models.Product{ProductName: "Banana", Price: -5, Weight: 1},
				ProductType: models.ProductType{TypeName: "Fruit"},
				ProductImages: []models.ProductImage{
					{Image: "url1.jpg"},
				},
			},
			seedData: func(){},
			wantStatus: http.StatusBadRequest,
			wantErr: dto.ProductError{
				ErrPrice: "Price can't be lower than 0!",
			},
			validateData: nil,
		},
		{
			name: "Happy path",
			payload: dto.ProductResponse{
				Product: models.Product{ProductName: "Banana", Price: 10, Weight: 1, ProductTypeID: 1},
				ProductType: models.ProductType{TypeName: "Type 1"},
				ProductImages: []models.ProductImage{
					{Image: "banana1.jpg"},
				},
			},
			wantStatus: http.StatusOK,
			seedData: func() {SeedData(t,SeedConfig{ProductTypes: &ProductTypeSeed{seedProductType: true,numberOfRecords: 1}})},
			wantMsg:    "Successfully created new product",
			wantErr: dto.ProductError{},
			validateData: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].(map[string]interface{})
				assert.NotEmpty(t, data)
				assert.Equal(t, "Banana", data["productName"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.seedData();

			var req *http.Request
			if str, ok := tt.payload.(string); ok {
				// invalid JSON
				req = httptest.NewRequest("POST", "/admin/product/create", strings.NewReader(str))
			} else {
				b, _ := json.Marshal(tt.payload)
				req = httptest.NewRequest("POST", "/admin/product/create", bytes.NewReader(b))
			}
			req.Header.Set("Content-Type", "application/json")

			resp, _ := app.Test(req, -1)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			if tt.wantMsg == "" {
				var got errorResponse
				if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
					t.Fatalf("failed to decode body: %v", err)
				}

				// Compare
				assert.Equal(t, tt.wantErr, got.Err)
			} else {
				var body map[string]interface{}
				_ = json.NewDecoder(resp.Body).Decode(&body)
				assert.Equal(t, tt.wantMsg, body["message"])
				if tt.validateData != nil{
					tt.validateData(t,body);
				}
			}

			// Reset DB
			_, err := testdb.Exec(`TRUNCATE TABLE product, product_type, product_images RESTART IDENTITY CASCADE`)
			assert.NoError(t, err)
		})
	}
}

type errorResponse struct {
	Status string           `json:"status"`
	Err    dto.ProductError `json:"err"`
}

func TestAdminProductUpdate(t *testing.T) {
	app := SetupApp()

	tests := []struct {
		name       string
		productID  string
		payload    interface{}
		seed       func()
		wantStatus int
		wantMsg    string
		wantErr    dto.ProductError
		validateData func(t *testing.T, body map[string]interface{})
	}{
		{
			name:       "Missing productID",
			productID:  "",
			payload:    dto.ProductResponse{},
			seed:       func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Did not receive productID",
			wantErr: dto.ProductError{},
		},
		{
			name:       "Invalid body",
			productID:  "1",
			payload:    `invalid-json`,
			seed:       func() {},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "Invalid request body",
			wantErr: dto.ProductError{},
		},
		{
			name: "Validation failed - missing productName",
			productID:  "1",
			payload: dto.ProductResponse{
				Product:     models.Product{Price: 10, Weight: 1},
				ProductType: models.ProductType{TypeName: "Fruit"},
				ProductImages: []models.ProductImage{
					{Image: "url1.jpg"},
				},
			},
			seed: func(){},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "",
			wantErr: dto.ProductError{
				ErrProductName: "Please input product name!",
			},
		},
		{
			name: "Validation failed - negative price",
			productID:  "1",
			payload: dto.ProductResponse{
				Product: models.Product{ProductName: "Banana", Price: -5, Weight: 1},
				ProductType: models.ProductType{TypeName: "Fruit"},
				ProductImages: []models.ProductImage{
					{Image: "url1.jpg"},
				},
			},
			seed: func(){},
			wantStatus: http.StatusBadRequest,
			wantErr: dto.ProductError{
				ErrPrice: "Price can't be lower than 0!",
			},
		},
		{
			name:      "Product not found",
			productID: "999",
			payload: dto.ProductResponse{
				Product: models.Product{ProductName: "Updated Banana", Price: 20, Weight: 2},
				ProductType: models.ProductType{TypeName: "Fruit"},
				ProductImages: []models.ProductImage{
					{Image: "banana2.jpg"},
				},
			},
			seed:       func() {},
			wantStatus: http.StatusInternalServerError,
			wantMsg:    "Product not found!",
			wantErr: dto.ProductError{},
		},
		{
			name:      "Happy path",
			productID: "1",
			payload: dto.ProductResponse{
				Product: models.Product{ProductID: 1,ProductName: "Product update", Price: 20, Weight: 2, ProductTypeID: 1},
				ProductType: models.ProductType{TypeName: "Type 1"},
				ProductImages: []models.ProductImage{
					{Image: "banana2.jpg"},
				},
			},
			seed:       func() {SeedData(t,SeedProductsBasic)},
			wantStatus: http.StatusOK,
			wantMsg:    "Successfully updated the product",
			wantErr: dto.ProductError{},
			validateData: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].(map[string]interface{})
				assert.NotEmpty(t,data);
				assert.Equal(t,"Product update",data["productName"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.seed();

			var req *http.Request
			if str, ok := tt.payload.(string); ok {
				req = httptest.NewRequest("PUT", "/admin/product/update?productID="+tt.productID, strings.NewReader(str))
			} else {
				b, _ := json.Marshal(tt.payload)
				req = httptest.NewRequest("PUT", "/admin/product/update?productID="+tt.productID, bytes.NewReader(b))
			}
			req.Header.Set("Content-Type", "application/json")

			resp, _ := app.Test(req, -1)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			if tt.wantMsg == "" {
				var got errorResponse
				if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
					t.Fatalf("failed to decode body: %v", err)
				}
				// Compare
				assert.Equal(t, tt.wantErr, got.Err)
			} else {
				var body map[string]interface{}
				_ = json.NewDecoder(resp.Body).Decode(&body)
				assert.Equal(t, tt.wantMsg, body["message"])
				if tt.validateData != nil{
					tt.validateData(t,body);
				}
			}


			// Reset DB
			_, err := testdb.Exec(`TRUNCATE TABLE product, product_type, product_images RESTART IDENTITY CASCADE`)
			assert.NoError(t, err)
		})
	}
}
