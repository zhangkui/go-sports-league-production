package service

import (
	"context"
	"strings"

	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/errorsx"
)

// UserService manages user accounts and role assignments (admin scope).
type UserService struct {
	*Deps
}

func NewUserService(d *Deps) *UserService { return &UserService{Deps: d} }

// List users with optional keyword search and pagination.
func (s *UserService) List(ctx context.Context, keyword string, p models.Pagination) ([]models.User, int64, error) {
	return s.Users.List(ctx, keyword, p, "id DESC")
}

// Get returns a single user with roles/permissions loaded.
func (s *UserService) Get(ctx context.Context, id int64) (*models.User, error) {
	u, err := s.Users.GetByID(ctx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("user", id)
	}
	_ = s.Users.LoadRolesAndPerms(ctx, u)
	return u, nil
}

// UpdateProfile patches mutable user fields.
func (s *UserService) UpdateProfile(ctx context.Context, id int64, req models.UpdateUserRequest) (*models.User, error) {
	if req.Email != nil {
		if err := emailValid(*req.Email); err != nil {
			return nil, err
		}
	}
	if err := s.Users.UpdateProfile(ctx, id, req.FullName, req.Phone, req.Email, req.Status); err != nil {
		return nil, errorsx.Internal("update failed")
	}
	return s.Get(ctx, id)
}

// Enable toggles a user's active status.
func (s *UserService) Enable(ctx context.Context, id int64, enabled bool) error {
	status := 0
	if enabled {
		status = 1
	}
	if err := s.Users.UpdateProfile(ctx, id, nil, nil, nil, &status); err != nil {
		return errorsx.Internal("status update failed")
	}
	return nil
}

// AssignRoles replaces a user's role set.
func (s *UserService) AssignRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	// validate role ids exist
	all, err := s.Roles.ListRoles(ctx)
	if err != nil {
		return errorsx.Internal("role lookup failed")
	}
	valid := map[int64]bool{}
	for _, r := range all {
		valid[r.ID] = true
	}
	for _, rid := range roleIDs {
		if !valid[rid] {
			return errorsx.BadRequest("invalid role id: " + itoa(rid))
		}
	}
	return s.Users.AssignRoles(ctx, userID, roleIDs)
}

// ResetPassword delegates to the auth service logic.
func (s *UserService) ResetPassword(ctx context.Context, userID int64, req models.ResetPasswordRequest) error {
	if err := passwordValid(req.NewPassword); err != nil {
		return err
	}
	authSvc := NewAuthService(s.Deps)
	return authSvc.ResetPassword(ctx, userID, req)
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

var _ = strings.TrimSpace
