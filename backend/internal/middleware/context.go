package middleware

import (
	"context"
	"net/http"

	"github.com/goxm2/sports-league/internal/models"
)

type ctxKey int

const (
	ctxKeyRequestID ctxKey = iota
	ctxKeyUser
	ctxKeyClientIP
)

// RequestIDMiddleware injects an X-Request-Id (generated if absent).
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = newUUID()
		}
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRequestID, id)))
	})
}

// RequestIDFromCtx retrieves the request id, or empty string.
func RequestIDFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyRequestID).(string)
	return v
}

// WithUser stores the authenticated user in the request context.
func WithUser(r *http.Request, u *models.User) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), ctxKeyUser, u))
}

// UserFromCtx retrieves the authenticated user, if any.
func UserFromCtx(r *http.Request) *models.User {
	v, _ := r.Context().Value(ctxKeyUser).(*models.User)
	return v
}

// ClientIPFromCtx retrieves the client IP, if set.
func ClientIPFromCtx(r *http.Request) string {
	if v, ok := r.Context().Value(ctxKeyClientIP).(string); ok {
		return v
	}
	return realIP(r)
}

// SetClientIP stores the resolved client IP in context.
func SetClientIP(r *http.Request) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), ctxKeyClientIP, realIP(r)))
}

func realIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	host := r.RemoteAddr
	for i := len(host) - 1; i >= 0; i-- {
		if host[i] == ':' {
			return host[:i]
		}
	}
	return host
}
