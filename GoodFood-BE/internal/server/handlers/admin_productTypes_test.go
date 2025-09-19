package handlers

import (
	"GoodFood-BE/internal/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdminProductTypePaginate(t *testing.T) {
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
