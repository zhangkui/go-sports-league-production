package models

import (
	"strings"
	"time"
)

// User is a registered platform account.
type User struct {
	ID           int64      `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	FullName     string     `json:"full_name" db:"full_name"`
	Phone        string     `json:"phone" db:"phone"`
	Status       int        `json:"status" db:"status"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	Roles        []Role     `json:"roles,omitempty" db:"-"`
	Permissions  []string   `json:"permissions,omitempty" db:"-"`
}

func (u User) IsActive() bool { return u.Status == 1 }

// SanitisedRoles returns role codes for the user.
func (u User) RoleCodes() []string {
	out := make([]string, 0, len(u.Roles))
	for _, r := range u.Roles {
		out = append(out, r.Code)
	}
	return out
}

// HasPermission checks whether the user holds a resource:action permission.
func (u User) HasPermission(code string) bool {
	for _, p := range u.Permissions {
		if p == code {
			return true
		}
	}
	return false
}

func (u User) HasAnyPermission(codes ...string) bool {
	for _, c := range codes {
		if u.HasPermission(c) {
			return true
		}
	}
	return false
}

// IsAdmin reports whether the user has the system admin role.
func (u User) IsAdmin() bool {
	for _, r := range u.Roles {
		if r.Code == "admin" {
			return true
		}
	}
	return false
}

// Role is an RBAC role.
type Role struct {
	ID          int64     `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	IsBuiltin   bool      `json:"is_builtin" db:"is_builtin"`
	PermissionCodes []string `json:"permission_codes,omitempty" db:"-"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Permission is a resource:action grant.
type Permission struct {
	ID       int64  `json:"id" db:"id"`
	Resource string `json:"resource" db:"resource"`
	Action   string `json:"action" db:"action"`
	Name     string `json:"name" db:"name"`
}

// Code returns the canonical "resource:action" string.
func (p Permission) Code() string {
	return p.Resource + ":" + p.Action
}

// RefreshToken is the persistent record of an issued refresh token.
type RefreshToken struct {
	ID        int64      `json:"-" db:"id"`
	TokenID   string     `json:"-" db:"token_id"`
	UserID    int64      `json:"user_id" db:"user_id"`
	TokenHash string     `json:"-" db:"token_hash"`
	IssuedAt  time.Time  `json:"issued_at" db:"issued_at"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
	UserAgent string     `json:"user_agent" db:"user_agent"`
	IP        string     `json:"ip" db:"ip"`
}

// IsRevoked reports whether the token has been revoked.
func (r RefreshToken) IsRevoked() bool { return r.RevokedAt != nil }

// IsExpired reports whether the token is past its expiry.
func (r RefreshToken) IsExpired() bool { return time.Now().After(r.ExpiresAt) }

// AuditLog records a security-relevant state change.
type AuditLog struct {
	ID         int64     `json:"id" db:"id"`
	UserID    *int64    `json:"user_id,omitempty" db:"user_id"`
	Username   string    `json:"username" db:"username"`
	Action     string    `json:"action" db:"action"`
	Resource   string    `json:"resource" db:"resource"`
	ResourceID string    `json:"resource_id,omitempty" db:"resource_id"`
	Method     string    `json:"method,omitempty" db:"method"`
	Path       string    `json:"path,omitempty" db:"path"`
	StatusCode *int      `json:"status_code,omitempty" db:"status_code"`
	IP         string    `json:"ip,omitempty" db:"ip"`
	RequestID  string    `json:"request_id,omitempty" db:"request_id"`
	Detail     string    `json:"detail,omitempty" db:"detail"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// PermCode joins resource and action with ':'.
func PermCode(resource, action string) string {
	return strings.TrimSpace(resource) + ":" + strings.TrimSpace(action)
}
