package service

import (
	"context"
	"strings"

	"github.com/goxm2/sports-league/internal/auth"
	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/errorsx"
	"github.com/goxm2/sports-league/internal/pkg/hash"
	"github.com/goxm2/sports-league/internal/repository"
)

// AuthService handles authentication, token issuance & rotation.
type AuthService struct {
	*Deps
}

func NewAuthService(d *Deps) *AuthService { return &AuthService{Deps: d} }

// Register creates a new self-registered user (spectator by default).
func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest) (*models.User, error) {
	if err := requireFields(map[string]string{
		"username": req.Username, "email": req.Email, "password": req.Password,
	}); err != nil {
		return nil, err
	}
	if err := usernameValid(req.Username); err != nil {
		return nil, err
	}
	if err := emailValid(req.Email); err != nil {
		return nil, err
	}
	if err := passwordValid(req.Password); err != nil {
		return nil, err
	}
	h, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, errorsx.Internal("password hashing failed")
	}
	u := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: h,
		FullName:     req.FullName,
		Status:       1,
	}
	if err := s.Users.Create(ctx, u); err != nil {
		return nil, err
	}
	// assign default spectator role
	spec, err := s.findRoleByCode(ctx, "spectator")
	if err == nil && spec != nil {
		_ = s.Users.AssignRoles(ctx, u.ID, []int64{spec.ID})
	}
	_ = s.Users.LoadRolesAndPerms(ctx, u)
	return u, nil
}

// Login authenticates a user and returns an access/refresh token pair.
func (s *AuthService) Login(ctx context.Context, req models.LoginRequest, ip, ua string) (*models.TokenPair, error) {
	if err := requireFields(map[string]string{"username": req.Username, "password": req.Password}); err != nil {
		return nil, err
	}
	u, err := s.Users.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, errorsx.Unauthorized("invalid credentials")
	}
	if !u.IsActive() {
		return nil, errorsx.Forbidden("account is disabled")
	}
	if err := hash.ComparePassword(u.PasswordHash, req.Password); err != nil {
		return nil, errorsx.Unauthorized("invalid credentials")
	}
	_ = s.Users.LoadRolesAndPerms(ctx, u)
	_ = s.Users.TouchLogin(ctx, u.ID)

	access, err := s.JWT.IssueAccessToken(u.ID, u.Username, u.RoleCodes())
	if err != nil {
		return nil, errorsx.Internal("token issue failed")
	}
	refreshID := auth.NewRefreshTokenID()
	rt := &models.RefreshToken{
		TokenID: refreshID, UserID: u.ID, ExpiresAt: s.JWT.RefreshExpiry(),
		UserAgent: ua, IP: ip,
	}
	rt.TokenHash, _ = hash.TokenHash(refreshID)
	if err := s.Tokens.Create(ctx, rt); err != nil {
		return nil, errorsx.Internal("refresh token persist failed")
	}
	return &models.TokenPair{
		AccessToken:  access,
		RefreshToken: refreshID,
		ExpiresIn:    int64(s.JWT.AccessTTL().Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// Refresh rotates a refresh token: revokes the old, issues a new pair.
func (s *AuthService) Refresh(ctx context.Context, refreshID, ip, ua string) (*models.TokenPair, error) {
	if refreshID == "" {
		return nil, errorsx.BadRequest("refresh_token is required")
	}
	rt, err := s.Tokens.GetByTokenID(ctx, refreshID)
	if err != nil {
		return nil, errorsx.Unauthorized("invalid refresh token")
	}
	if rt.IsRevoked() || rt.IsExpired() {
		_ = s.Tokens.Revoke(ctx, refreshID)
		return nil, errorsx.Unauthorized("refresh token expired or revoked")
	}
	u, err := s.Users.GetByID(ctx, rt.UserID)
	if err != nil || !u.IsActive() {
		return nil, errorsx.Unauthorized("user not found or disabled")
	}
	_ = s.Users.LoadRolesAndPerms(ctx, u)
	// rotate: revoke old, issue new
	_ = s.Tokens.Revoke(ctx, refreshID)
	access, _ := s.JWT.IssueAccessToken(u.ID, u.Username, u.RoleCodes())
	newID := auth.NewRefreshTokenID()
	nrt := &models.RefreshToken{
		TokenID: newID, UserID: u.ID, ExpiresAt: s.JWT.RefreshExpiry(),
		UserAgent: ua, IP: ip,
	}
	nrt.TokenHash, _ = hash.TokenHash(newID)
	if err := s.Tokens.Create(ctx, nrt); err != nil {
		return nil, errorsx.Internal("refresh token persist failed")
	}
	return &models.TokenPair{
		AccessToken:  access,
		RefreshToken: newID,
		ExpiresIn:    int64(s.JWT.AccessTTL().Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// Logout revokes the supplied refresh token.
func (s *AuthService) Logout(ctx context.Context, refreshID string) error {
	if refreshID == "" {
		return nil
	}
	if err := s.Tokens.Revoke(ctx, refreshID); err != nil {
		return errorsx.Internal("logout failed")
	}
	return nil
}

// ChangePassword validates the old password and sets a new one.
func (s *AuthService) ChangePassword(ctx context.Context, userID int64, req models.ChangePasswordRequest) error {
	if err := requireFields(map[string]string{"old_password": req.OldPassword, "new_password": req.NewPassword}); err != nil {
		return err
	}
	if err := passwordValid(req.NewPassword); err != nil {
		return err
	}
	u, err := s.Users.GetByID(ctx, userID)
	if err != nil {
		return errorsx.NotFoundID("user", userID)
	}
	if err := hash.ComparePassword(u.PasswordHash, req.OldPassword); err != nil {
		return errorsx.BadRequest("old password incorrect")
	}
	h, err := hash.HashPassword(req.NewPassword)
	if err != nil {
		return errorsx.Internal("password hashing failed")
	}
	if err := s.Users.UpdatePassword(ctx, userID, h); err != nil {
		return errorsx.Internal("password update failed")
	}
	// revoke all sessions
	_ = s.Tokens.RevokeAllForUser(ctx, userID)
	return nil
}

// ResetPassword is the admin-driven password reset.
func (s *AuthService) ResetPassword(ctx context.Context, userID int64, req models.ResetPasswordRequest) error {
	if err := passwordValid(req.NewPassword); err != nil {
		return err
	}
	h, err := hash.HashPassword(req.NewPassword)
	if err != nil {
		return errorsx.Internal("password hashing failed")
	}
	if err := s.Users.UpdatePassword(ctx, userID, h); err != nil {
		return errorsx.Internal("password update failed")
	}
	_ = s.Tokens.RevokeAllForUser(ctx, userID)
	return nil
}

// Me returns the full current user profile (with roles & permissions).
func (s *AuthService) Me(ctx context.Context, userID int64) (*models.User, error) {
	u, err := s.Users.GetByID(ctx, userID)
	if err != nil {
		return nil, errorsx.NotFoundID("user", userID)
	}
	_ = s.Users.LoadRolesAndPerms(ctx, u)
	return u, nil
}

// findRoleByCode looks up a role by its code across all roles.
func findRoleByCode(ctx context.Context, roles *repository.RoleRepo, code string) (*models.Role, error) {
	all, err := roles.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	for i := range all {
		if all[i].Code == code {
			return &all[i], nil
		}
	}
	return nil, errorsx.NotFoundID("role", code)
}

func (s *AuthService) findRoleByCode(ctx context.Context, code string) (*models.Role, error) {
	return findRoleByCode(ctx, s.Roles, code)
}

// --- local validators (avoid import cycle with validator pkg naming) ---

func requireFields(m map[string]string) error {
	for k, v := range m {
		if strings.TrimSpace(v) == "" {
			return errorsx.BadRequest(k + " is required")
		}
	}
	return nil
}

func passwordValid(p string) error {
	if len(p) < 8 {
		return errorsx.BadRequest("password must be at least 8 characters")
	}
	hasLetter, hasDigit := false, false
	for _, r := range p {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'):
			hasLetter = true
		case r >= '0' && r <= '9':
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return errorsx.BadRequest("password must contain letters and digits")
	}
	return nil
}

func usernameValid(u string) error {
	const allowed = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_.-"
	if len(u) < 3 || len(u) > 32 {
		return errorsx.BadRequest("username must be 3-32 chars")
	}
	for _, r := range u {
		if !strings.ContainsRune(allowed, r) {
			return errorsx.BadRequest("username has invalid characters")
		}
	}
	return nil
}

func emailValid(e string) error {
	if !strings.Contains(e, "@") || !strings.Contains(e, ".") || len(e) < 5 {
		return errorsx.BadRequest("invalid email format")
	}
	return nil
}

func validateAll(_ ...string) error { return nil }
