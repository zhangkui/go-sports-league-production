package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/goxm2/sports-league/internal/pkg/logx"
)

// statusWriter captures the response status for logging.
type statusWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (s *statusWriter) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusWriter) Write(b []byte) (int, error) {
	n, err := s.ResponseWriter.Write(b)
	s.size += n
	return n, err
}

// LoggingMiddleware records structured access logs for each request.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, SetClientIP(r))
		latency := time.Since(start)
		logx.Info("http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"size", sw.size,
			"latency_ms", latency.Milliseconds(),
			"ip", realIP(r),
			"ua", r.UserAgent(),
		)
	})
}

// RecoveryMiddleware catches panics and returns a 500 envelope.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logx.Error("panic recovered", "err", rec, "path", r.URL.Path)
				http.Error(w, `{"code":50000,"message":"internal server error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware handles preflight and sets permissive CORS headers.
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	allowAll := false
	allowed := map[string]bool{}
	for _, o := range allowedOrigins {
		o = strings.TrimSpace(o)
		if o == "*" {
			allowAll = true
		}
		allowed[o] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if allowAll || (origin != "" && allowed[origin]) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,Idempotency-Key,X-Request-Id")
				w.Header().Set("Access-Control-Expose-Headers", "X-Request-Id,X-RateLimit-Remaining")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
