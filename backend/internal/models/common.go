package models

import "time"

// JSONTime wraps time.Time with RFC3339 JSON marshalling.
type JSONTime time.Time

// Pagination params parsed from query string.
type Pagination struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
	OrderBy  string `json:"order_by" form:"order_by"`
	Order    string `json:"order" form:"order"` // asc|desc
}

func (p *Pagination) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 200 {
		p.PageSize = 20
	}
	if p.Order != "asc" && p.Order != "desc" {
		p.Order = "desc"
	}
}

func (p Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}

func (p Pagination) Limit() int {
	return p.PageSize
}

// PageResult wraps a list with pagination metadata.
type PageResult[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}
