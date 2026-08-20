package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/response"
	"github.com/goxm2/sports-league/internal/service"
)

// AuthMiddleware validates the Bearer JWT and loads the user into context.
type AuthMiddleware struct {
	Auth *service.AuthService
}

func NewAuthMiddleware(auth *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{Auth: auth}
}

// RequireAuth wraps a handler to require a valid access token.
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := bearerToken(r)
		if tokenStr == "" {
			response.Error(w, r, http.StatusUnauthorized, response.CodeUnauthorized, "missing or malformed Authorization header")
			return
		}
		claims, err := m.Auth.JWT.ParseAccessToken(tokenStr)
		if err != nil {
			response.Error(w, r, http.StatusUnauthorized, response.CodeUnauthorized, "invalid or expired token")
			return
		}
		u, err := m.Auth.Users.GetByID(r.Context(), claims.UserID)
		if err != nil || !u.IsActive() {
			response.Error(w, r, http.StatusUnauthorized, response.CodeUnauthorized, "user not found or disabled")
			return
		}
		_ = m.Auth.Users.LoadRolesAndPerms(r.Context(), u)
		next.ServeHTTP(w, WithUser(r, u))
	})
}

// OptionalAuth loads the user if a token is present, but does not reject.
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := bearerToken(r)
		if tokenStr == "" {
			next.ServeHTTP(w, r)
			return
		}
		claims, err := m.Auth.JWT.ParseAccessToken(tokenStr)
		if err == nil {
			if u, e := m.Auth.Users.GetByID(r.Context(), claims.UserID); e == nil && u.IsActive() {
				_ = m.Auth.Users.LoadRolesAndPerms(r.Context(), u)
				next.ServeHTTP(w, WithUser(r, u))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// RequirePermission checks the context user holds the given permission code.
func RequirePermission(code string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := UserFromCtx(r)
			if u == nil {
				response.Error(w, r, http.StatusUnauthorized, response.CodeUnauthorized, "authentication required")
				return
			}
			if u.IsAdmin() || u.HasPermission(code) {
				next.ServeHTTP(w, r)
				return
			}
			response.Error(w, r, http.StatusForbidden, response.CodeForbidden, "insufficient permissions: "+code)
		})
	}
}

// RequireAnyPermission checks the user holds at least one of the codes.
func RequireAnyPermission(codes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := UserFromCtx(r)
			if u == nil {
				response.Error(w, r, http.StatusUnauthorized, response.CodeUnauthorized, "authentication required")
				return
			}
			if u.IsAdmin() || u.HasAnyPermission(codes...) {
				next.ServeHTTP(w, r)
				return
			}
			response.Error(w, r, http.StatusForbidden, response.CodeForbidden, "insufficient permissions")
		})
	}
}

// RequireRole checks the user holds one of the given role codes.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := UserFromCtx(r)
			if u == nil {
				response.Error(w, r, http.StatusUnauthorized, response.CodeUnauthorized, "authentication required")
				return
			}
			if u.IsAdmin() {
				next.ServeHTTP(w, r)
				return
			}
			for _, want := range roles {
				for _, have := range u.RoleCodes() {
					if have == want {
						next.ServeHTTP(w, r)
						return
					}
				}
			}
			response.Error(w, r, http.StatusForbidden, response.CodeForbidden, "insufficient role")
		})
	}
}

// AuditSubject returns the acting user id (0 if anonymous) for audit logging.
func AuditSubject(r *http.Request) (int64, string, []string) {
	u := UserFromCtx(r)
	if u == nil {
		return 0, "anonymous", nil
	}
	return u.ID, u.Username, u.RoleCodes()
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// ensure models.User import is used for godoc reference.
var _ = models.User{}

// ctx import marker
var _ context.Context = nil
