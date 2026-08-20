package service

import (
	"context"
	"strings"

	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/errorsx"
)

// RoleService manages RBAC roles and permissions.
type RoleService struct {
	*Deps
}

func NewRoleService(d *Deps) *RoleService { return &RoleService{Deps: d} }

// ListRoles returns all roles.
func (s *RoleService) ListRoles(ctx context.Context) ([]models.Role, error) {
	roles, err := s.Roles.ListRoles(ctx)
	if err != nil {
		return nil, errorsx.Internal("role lookup failed")
	}
	for i := range roles {
		codes, _ := s.Roles.RolePermissionCodes(ctx, roles[i].ID)
		roles[i].PermissionCodes = codes
	}
	return roles, nil
}

// GetRole returns a role by id with its permission codes.
func (s *RoleService) GetRole(ctx context.Context, id int64) (*models.Role, error) {
	ro, err := s.Roles.GetRole(ctx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("role", id)
	}
	codes, _ := s.Roles.RolePermissionCodes(ctx, id)
	ro.PermissionCodes = codes
	return ro, nil
}

// CreateRole creates a non-builtin role with the given permissions.
func (s *RoleService) CreateRole(ctx context.Context, req models.CreateRoleRequest) (*models.Role, error) {
	req.Code = strings.TrimSpace(req.Code)
	req.Name = strings.TrimSpace(req.Name)
	if req.Code == "" || req.Name == "" {
		return nil, errorsx.BadRequest("code and name are required")
	}
	if err := codeFieldValid(req.Code); err != nil {
		return nil, err
	}
	ro := &models.Role{Code: req.Code, Name: req.Name, Description: req.Description}
	if err := s.Roles.CreateRole(ctx, ro, req.Permissions); err != nil {
		return nil, err
	}
	return ro, nil
}

// UpdateRole updates a role and optionally replaces its permissions.
func (s *RoleService) UpdateRole(ctx context.Context, id int64, req models.UpdateRoleRequest) (*models.Role, error) {
	ro, err := s.Roles.GetRole(ctx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("role", id)
	}
	if req.Name != nil {
		ro.Name = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		ro.Description = *req.Description
	}
	replace := req.Permissions != nil
	if err := s.Roles.UpdateRole(ctx, ro, req.Permissions, replace); err != nil {
		return nil, errorsx.Internal("role update failed")
	}
	return s.GetRole(ctx, id)
}

// DeleteRole deletes a non-builtin role with no users attached.
func (s *RoleService) DeleteRole(ctx context.Context, id int64) error {
	ro, err := s.Roles.GetRole(ctx, id)
	if err != nil {
		return errorsx.NotFoundID("role", id)
	}
	if ro.IsBuiltin {
		return errorsx.BadRequest("cannot delete builtin role")
	}
	n, _ := s.Roles.CountRoleUsers(ctx, id)
	if n > 0 {
		return errorsx.Conflict("role still has users assigned")
	}
	return s.Roles.DeleteRole(ctx, id)
}

// ListPermissions returns all permission rows.
func (s *RoleService) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	return s.Roles.ListPermissions(ctx)
}

func codeFieldValid(v string) error {
	const allowed = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
	if len(v) < 2 || len(v) > 32 {
		return errorsx.BadRequest("code must be 2-32 chars")
	}
	for _, r := range v {
		if !strings.ContainsRune(allowed, r) {
			return errorsx.BadRequest("code has invalid characters")
		}
	}
	return nil
}
