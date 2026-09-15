// Package pagination provides the shared {items,total,page,page_size}
// request/response envelope used identically by every list endpoint
// across api-docs/openapi/*.yaml (GET /events, GET /users/me/bookings, ...).
package pagination

import "strconv"

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type Params struct {
	Page     int
	PageSize int
}

// ParseParams parses raw query values (empty string = use default),
// clamping page_size to [1, MaxPageSize] and page to >= 1.
func ParseParams(pageStr, pageSizeStr string) Params {
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = DefaultPage
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return Params{Page: page, PageSize: pageSize}
}

func (p Params) Offset() int { return (p.Page - 1) * p.PageSize }
func (p Params) Limit() int  { return p.PageSize }

type Envelope[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
}

func New[T any](items []T, total int64) Envelope[T] {
	if items == nil {
		items = []T{}
	}
	return Envelope[T]{Items: items, Total: total}
}
