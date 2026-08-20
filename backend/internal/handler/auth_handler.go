package handler

import (
	"net/http"

	"github.com/goxm2/sports-league/internal/middleware"
	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/response"
	"github.com/goxm2/sports-league/internal/service"
)

// AuthHandler exposes auth endpoints.
type AuthHandler struct {
	Auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler { return &AuthHandler{Auth: auth} }

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	u, err := h.Auth.Register(r.Context(), req)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.WriteCreated(w, r, u)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	pair, err := h.Auth.Login(r.Context(), req, middleware.ClientIPFromCtx(r), r.UserAgent())
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, pair)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	pair, err := h.Auth.Refresh(r.Context(), req.RefreshToken, middleware.ClientIPFromCtx(r), r.UserAgent())
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, pair)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshRequest
	_ = decodeJSON(w, r, &req)
	_ = h.Auth.Logout(r.Context(), req.RefreshToken)
	response.Write(w, r, map[string]any{"logged_out": true})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	uid, _ := currentUser(r)
	u, err := h.Auth.Me(r.Context(), uid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, u)
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	uid, _ := currentUser(r)
	var req models.ChangePasswordRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := h.Auth.ChangePassword(r.Context(), uid, req); err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, map[string]any{"changed": true})
}
