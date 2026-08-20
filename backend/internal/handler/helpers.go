package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/goxm2/sports-league/internal/middleware"
	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/errorsx"
	"github.com/goxm2/sports-league/internal/pkg/pagination"
	"github.com/goxm2/sports-league/internal/pkg/response"
)

// ctx helpers ---------------------------------------------------------

func currentUser(r *http.Request) (int64, string) {
	u := middleware.UserFromCtx(r)
	if u == nil {
		return 0, "anonymous"
	}
	return u.ID, u.Username
}

func pathID(r *http.Request, key string) (int64, error) {
	return pagination.PathInt64(r, key)
}

func parsePage(r *http.Request) models.Pagination { return pagination.Parse(r) }

// decodeJSON decodes the request body into v, validating non-empty body.
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid request body: "+err.Error())
		return false
	}
	return true
}

// writeErr maps an app error to its HTTP response.
func writeErr(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}
	ae, ok := err.(*errorsx.AppError)
	if !ok {
		response.Error(w, r, http.StatusInternalServerError, response.CodeInternal, err.Error())
		return
	}
	response.Error(w, r, ae.HTTPStatus, ae.Code, ae.Message)
}

// queryInt reads an integer query param with a default.
func queryInt(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

// queryStr reads a string query param.
func queryStr(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}
