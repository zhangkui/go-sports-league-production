package pagination

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/goxm2/sports-league/internal/models"
)

// Parse reads page/page_size/order_by/order from query.
func Parse(r *http.Request) models.Pagination {
	q := r.URL.Query()
	p := models.Pagination{
		Page:     atoi(q.Get("page"), 1),
		PageSize: atoi(q.Get("page_size"), 20),
		OrderBy:  q.Get("order_by"),
		Order:     q.Get("order"),
	}
	p.Normalize()
	return p
}

// WhitelistOrderBy sanitises the order_by column against an allow-list.
// Returns the validated "column" (safe to interpolate, NOT a user value
// echoed blindly) or empty string if disallowed.
func WhitelistOrderBy(p models.Pagination, allowed map[string]string) string {
	if p.OrderBy == "" {
		return ""
	}
	col, ok := allowed[strings.ToLower(p.OrderBy)]
	if !ok {
		return ""
	}
	dir := "DESC"
	if strings.EqualFold(p.Order, "asc") {
		dir = "ASC"
	}
	return col + " " + dir
}

func atoi(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return def
	}
	return n
}

// PathInt64 reads an id path variable as int64.
func PathInt64(r *http.Request, key string) (int64, error) {
	v := r.PathValue(key)
	if v == "" {
		// gorilla/mux fallback
		v = muxVar(r, key)
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}
