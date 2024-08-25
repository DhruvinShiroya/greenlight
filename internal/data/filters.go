package data

import (
	"math"
	"strings"

	"github.com/DhruvinShiroya/greenlight/internal/validator"
)

type Filter struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

type Metadata struct {
	CurrentPage int `json:"current_page,omitempty`
	PageSize    int `json:"page_size,omitempty`
	FirstPage   int `json:"first_page,omitempty`
	LastPage    int `json:"last_page,omitempty`
	TotalRecord int `json:"total_record,omitempty`
}

func ValidateFilter(v *validator.Validator, f Filter) {
	// check that page and pagesize has valid type and range of number
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 1_000_000, "page", "must be less than 1 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 20, "page_size", "must be less than twenty")

	// now check if the sort pramter contains the values match in the SortSafelist
	v.Check(validator.In(f.Sort, f.SortSafelist...), "sort", "invalid sort field")
}

// return sortColumn value and trim prefix if its '-'
func (f Filter) sortColumn() string {
	for _, safeVal := range f.SortSafelist {
		if f.Sort == safeVal {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	// it is fail safe against sql injection but i will provide default value
	panic("unsafe sort paramter: " + f.Sort)
	// return "id"
}

// get the sort dircetion based on Sort field prefix
func (f Filter) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

func (f Filter) limit() int {
	return f.PageSize
}

func (f Filter) offset() int {
	return f.PageSize * (f.Page - 1)
}

func calculateMetaData(TotalRecord, page, pageSize int) Metadata {
	if TotalRecord == 0 {
		return Metadata{}
	}
	return Metadata{
		CurrentPage: page,
		PageSize:    pageSize,
		FirstPage:   1,
		LastPage:    int(math.Ceil(float64(TotalRecord) / float64(pageSize))),
		TotalRecord: TotalRecord,
	}
}
