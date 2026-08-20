package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/logx"
	"github.com/goxm2/sports-league/internal/service"
)

// AuditMiddleware records audit log entries for mutating requests on
// audited resources. It runs after auth so the user is known.
type AuditMiddleware struct {
	Svc *service.ReportService
}

func NewAuditMiddleware(rep *service.ReportService) *AuditMiddleware {
	return &AuditMiddleware{Svc: rep}
}

// auditedRoutes maps a path prefix + method to an audit action/resource.
func auditAction(r *http.Request) (action, resource string, ok bool) {
	if r.Method == http.MethodGet || r.Method == http.MethodOptions {
		return "", "", false
	}
	path := r.URL.Path
	prefix := "/api/v1/"
	if len(path) <= len(prefix) {
		return "", "", false
	}
	rest := path[len(prefix):]
	// take first segment as resource
	seg := rest
	for i := 0; i < len(rest); i++ {
		if rest[i] == '/' {
			seg = rest[:i]
			break
		}
	}
	seg = pluralToSingular(seg)
	switch seg {
	case "users", "user":
		resource = "users"
	case "roles", "role":
		resource = "roles"
	case "seasons", "season":
		resource = "seasons"
	case "teams", "team":
		resource = "teams"
	case "players", "player":
		resource = "players"
	case "venues", "venue":
		resource = "venues"
	case "schedules", "schedule":
		resource = "schedules"
	case "matches", "match":
		resource = "matches"
	case "disciplines", "discipline":
		resource = "disciplines"
	case "appeals", "appeal":
		resource = "appeals"
	case "transfers", "transfer":
		resource = "transfers"
	default:
		return "", "", false
	}
	action = r.Method + ":" + resource
	return action, resource, true
}

// Wrap records an audit entry after the handler completes.
func (am *AuditMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		action, resource, ok := auditAction(r)
		if !ok {
			return
		}
		uid, username, _ := AuditSubject(r)
		statusCode := sw.status
		entry := &models.AuditLog{
			UserID:    nilIfZero(uid),
			Username:  username,
			Action:    action,
			Resource:  resource,
			Method:    r.Method,
			Path:      r.URL.Path,
			StatusCode: &statusCode,
			IP:        realIP(r),
			RequestID: RequestIDFromCtx(r.Context()),
		}
		// best-effort: never block on audit failure
		ctx, cancel := context.WithTimeout(context.Background(), auditTimeout)
		defer cancel()
		_ = am.Svc.Audit.Create(ctx, nil, entry)
		if entry.ID == 0 {
			// fallback: at least log it
			logx.Info("audit", "action", action, "resource", resource, "user", username, "status", statusCode)
		}
	})
}

func pluralToSingular(s string) string {
	// crude: strip trailing 's'
	if len(s) > 1 && s[len(s)-1] == 's' {
		return s[:len(s)-1]
	}
	return s
}

func nilIfZero(n int64) *int64 {
	if n == 0 {
		return nil
	}
	return &n
}

const auditTimeout = 2 * time.Second

var _ = strconv.Itoa
