package handlers

import (
	"GoodFood-BE/internal/utils"
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
)

func TestAdminInvoice_Paginate(t *testing.T) {
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

func TestBuildCancelEmailBody(t *testing.T) {
	html := utils.BuildCancelEmailBody("Out of stock", false)
	assert.Contains(t, html, "Out of stock")
	assert.Contains(t, html, "Cash on Delivery (COD)")

	htmlPaid := utils.BuildCancelEmailBody("Payment issue", true)
	assert.Contains(t, htmlPaid, "Payment issue")
	assert.Contains(t, htmlPaid, "refund")
}

// utils.ParseDateRange → test input invalid date, empty string, dateFrom > dateTo.
func TestParseDateRange(t *testing.T){
	//Table tests setup
	tests := []struct {
		name         string
		dateFromStr     string
		dateToStr    	 string
		wantDateFrom   time.Time
		wantDateTo     time.Time
		wantMsg      string
	}{
		{
			name: "Invalid date format",
			dateFromStr: "00-00-0000",
			dateToStr: "99-99-9999",
			wantDateFrom: time.Time{}, 
			wantDateTo: time.Time{},
			wantMsg: "Invalid format for dateFrom/dateTo (expect yyyy-mm-dd)",
		},
		{
			name: "Empty dates",
			dateFromStr: "",
			dateToStr: "",
			wantDateFrom: time.Time{},
			wantDateTo: time.Time{},
			wantMsg: "Date from or date to is empty!",
		},
		{
			name: "Date from cannot be after date to",
			dateFromStr: "2025-12-12",
			dateToStr: "2025-05-05",
			wantDateFrom: time.Time{},
			wantDateTo: time.Time{},
			wantMsg: "Date to can't be before date from",
		},
		{
			name: "Successful date parse!",
			dateFromStr: "2025-04-04",
			dateToStr: "2025-05-05",
			wantDateFrom: mustParseDate("2025-04-04"),
			wantDateTo: mustParseDate("2025-05-05"),
			wantMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dateFrom, dateTo, err := utils.ParseDateRange(tt.dateFromStr, tt.dateToStr)
			assert.Equal(t, tt.wantDateFrom, dateFrom)
			assert.Equal(t, tt.wantDateTo, dateTo)
			if tt.wantMsg != ""{
				assert.Equal(t,tt.wantMsg, err.Error())
			}else{
				assert.NoError(t, err)
			}
		})
	}
}

func mustParseDate(date string) time.Time {
    t, err := time.Parse("2006-01-02", date)
    if err != nil {
        panic(err) // chỉ chạy khi test, nên panic cũng được
    }
    return t
}