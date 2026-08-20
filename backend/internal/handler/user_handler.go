package handler

import (
	"net/http"
	"strconv"

	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/response"
	"github.com/goxm2/sports-league/internal/service"
)

// UserHandler exposes admin user & role management endpoints.
type UserHandler struct {
	Users *service.UserService
	Roles *service.RoleService
	Auth  *service.AuthService
}

func NewUserHandler(users *service.UserService, roles *service.RoleService, auth *service.AuthService) *UserHandler {
	return &UserHandler{Users: users, Roles: roles, Auth: auth}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	keyword := queryStr(r, "q")
	p := parsePage(r)
	users, total, err := h.Users.List(r.Context(), keyword, p)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.WritePage(w, r, users, total, p.Page, p.PageSize)
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	u, err := h.Users.Get(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, u)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	var req models.UpdateUserRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	u, err := h.Users.UpdateProfile(r.Context(), id, req)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, u)
}

func (h *UserHandler) Enable(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	enabled := queryStr(r, "enabled") != "false"
	if err := h.Users.Enable(r.Context(), id, enabled); err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, map[string]any{"enabled": enabled})
}

func (h *UserHandler) AssignRoles(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	var req models.AssignRolesRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := h.Users.AssignRoles(r.Context(), id, req.RoleIDs); err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, map[string]any{"assigned": req.RoleIDs})
}

func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	var req models.ResetPasswordRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := h.Users.ResetPassword(r.Context(), id, req); err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, map[string]any{"reset": true})
}

// Roles -----------------------------------------------------------------

func (h *UserHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.Roles.ListRoles(r.Context())
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, roles)
}

func (h *UserHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	ro, err := h.Roles.GetRole(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, ro)
}

func (h *UserHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRoleRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	ro, err := h.Roles.CreateRole(r.Context(), req)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.WriteCreated(w, r, ro)
}

func (h *UserHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	var req models.UpdateRoleRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	ro, err := h.Roles.UpdateRole(r.Context(), id, req)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, ro)
}

func (h *UserHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	if err := h.Roles.DeleteRole(r.Context(), id); err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, map[string]any{"deleted": true})
}

func (h *UserHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	perms, err := h.Roles.ListPermissions(r.Context())
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, perms)
}

var _ = strconv.Itoa
