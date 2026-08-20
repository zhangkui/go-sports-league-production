package response

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Standard API envelope. Code 0 = success, non-zero = business error.
type Envelope struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	Timestamp string      `json:"timestamp"`
	RequestID string      `json:"request_id"`
}

// Standard business error codes.
const (
	CodeSuccess           = 0
	CodeBadRequest        = 40000
	CodeUnauthorized      = 40100
	CodeForbidden         = 40300
	CodeNotFound          = 40400
	CodeConflict          = 40900
	CodeTooManyRequests   = 42900
	CodeInternal          = 50000
)

// FromCtx pulls the request id from context if present.
func FromCtx(r *http.Request) string {
	if v := r.Header.Get("X-Request-Id"); v != "" {
		return v
	}
	return uuid.NewString()
}

// Write sends a success envelope with data.
func Write(w http.ResponseWriter, r *http.Request, data interface{}) {
	writeJSON(w, http.StatusOK, &Envelope{
		Code:      CodeSuccess,
		Message:   "success",
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RequestID: FromCtx(r),
	})
}

// WriteCreated sends 201 with data.
func WriteCreated(w http.ResponseWriter, r *http.Request, data interface{}) {
	writeJSON(w, http.StatusCreated, &Envelope{
		Code:      CodeSuccess,
		Message:   "created",
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RequestID: FromCtx(r),
	})
}

// WritePage sends a paginated list.
func WritePage[T any](w http.ResponseWriter, r *http.Request, list []T, total int64, page, size int) {
	if list == nil {
		list = []T{}
	}
	Write(w, r, map[string]interface{}{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": size,
	})
}

// Error writes an error envelope with the given HTTP status and business code.
func Error(w http.ResponseWriter, r *http.Request, httpStatus, code int, msg string) {
	writeJSON(w, httpStatus, &Envelope{
		Code:      code,
		Message:   msg,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RequestID: FromCtx(r),
	})
}

func writeJSON(w http.ResponseWriter, status int, env *Envelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(env)
}
