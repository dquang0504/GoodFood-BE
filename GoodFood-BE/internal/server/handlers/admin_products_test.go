package handlers

import (
	"GoodFood-BE/internal/dto"
	"GoodFood-BE/internal/utils"
	"GoodFood-BE/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationProduct(t *testing.T) {
    tests := []struct {
        name    string
        input   dto.ProductResponse
        diff    bool
        wantOk  bool
        wantErr dto.ProductError
    }{
        {
            name: "Valid product",
            input: dto.ProductResponse{
                Product: models.Product{
					ProductID: 1,
					ProductName: "Apple",
					Price:       100,
					Weight:      50,
				},
                ProductType: models.ProductType{TypeName: "Fruit"},
                ProductImages: []models.ProductImage{
                    {Image: "apple.jpg"},
                },
            },
            diff:   true,
            wantOk: true,
            wantErr: dto.ProductError{},
        },
        {
            name: "Missing name",
            input: dto.ProductResponse{},
            diff:   true,
            wantOk: false,
            wantErr: dto.ProductError{ErrProductName: "Please input product name!"},
        },
        {
            name: "Negative price",
            input: dto.ProductResponse{
				Product: models.Product{
					ProductID: 1,
					ProductName: "Orange",
                	Price:       -10,
					Weight:      50,
				},
                ProductType: models.ProductType{TypeName: "Fruit"},
                ProductImages: []models.ProductImage{{Image: "orange.jpg"}},
            },
            diff:   true,
            wantOk: false,
            wantErr: dto.ProductError{ErrPrice: "Price can't be lower than 0!"},
        },
        {
            name: "Missing images when diff = true",
            input: dto.ProductResponse{
				Product: models.Product{
					ProductID: 1,
					ProductName: "Banana",
                	Price:       20,
					Weight:      50,
				},   
                ProductType: models.ProductType{TypeName: "Fruit"},
            },
            diff:   true,
            wantOk: false,
            wantErr: dto.ProductError{ErrImages: "Please upload the product's image!"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ok, errResp := validationProduct(&tt.input, tt.diff)
            assert.Equal(t, tt.wantOk, ok)
            if !tt.wantOk {
                assert.NotEqual(t, dto.ProductError{}, errResp)
            }
        })
    }
}

func TestAdminProductPaginate(t *testing.T) {
	tests := []struct {
		name        string
		total       int
		page        int
		pageSize    int
		wantOffset  int
		wantTotalPg int
	}{
		{
			name:        "No records",
			total:       0,
			page:        1,
			pageSize:    6,
			wantOffset:  0,
			wantTotalPg: 0,
		},
		{
			name:        "Single page",
			total:       5,
			page:        1,
			pageSize:    6,
			wantOffset:  0,
			wantTotalPg: 1,
		},
		{
			name:        "Multiple pages first page",
			total:       12,
			page:        1,
			pageSize:    6,
			wantOffset:  0,
			wantTotalPg: 2,
		},
		{
			name:        "Multiple pages second page",
			total:       12,
			page:        2,
			pageSize:    6,
			wantOffset:  6,
			wantTotalPg: 2,
		},
		{
			name:        "Page exceeds total",
			total:       12,
			page:        5,
			pageSize:    6,
			wantOffset:  24,
			wantTotalPg: 2,
		},
		{
			name:        "Invalid page input (0)",
			total:       10,
			page:        0,
			pageSize:    6,
			wantOffset:  0,
			wantTotalPg: 2,
		},
		{
			name:        "Invalid pageSize input (0)",
			total:       10,
			page:        1,
			pageSize:    0, //default fallback is 6
			wantOffset:  0,
			wantTotalPg: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset, totalPage := utils.Paginate(tt.page, tt.pageSize, tt.total)
			assert.Equal(t, tt.wantOffset, offset)
			assert.Equal(t, tt.wantTotalPg, totalPage)
		})
	}
}