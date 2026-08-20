package handler

import (
	"net/http"

	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/response"
	"github.com/goxm2/sports-league/internal/service"
)

// SeasonHandler exposes season + scoring rule endpoints.
type SeasonHandler struct {
	Svc *service.SeasonService
}

func NewSeasonHandler(svc *service.SeasonService) *SeasonHandler { return &SeasonHandler{Svc: svc} }

func (h *SeasonHandler) List(w http.ResponseWriter, r *http.Request) {
	status := queryStr(r, "status")
	sport := queryStr(r, "sport")
	p := parsePage(r)
	list, total, err := h.Svc.List(r.Context(), status, sport, p)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.WritePage(w, r, list, total, p.Page, p.PageSize)
}

func (h *SeasonHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	se, err := h.Svc.Get(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, se)
}

func (h *SeasonHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, _ := currentUser(r)
	var req models.CreateSeasonRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	se, err := h.Svc.Create(r.Context(), req, uid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.WriteCreated(w, r, se)
}

func (h *SeasonHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	var req models.UpdateSeasonRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	se, err := h.Svc.Update(r.Context(), id, req)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, se)
}

func (h *SeasonHandler) ChangeStatus(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	var req models.SeasonStatusChange
	if !decodeJSON(w, r, &req) {
		return
	}
	se, err := h.Svc.ChangeStatus(r.Context(), id, req.Status)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, se)
}

// Scoring rules ---------------------------------------------------------

func (h *SeasonHandler) GetActiveRule(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	rule, err := h.Svc.GetActiveScoringRule(r.Context(), id)
	if rule == nil {
		rule = models.EmptyScoringRule(id)
	}
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, rule)
}

func (h *SeasonHandler) SetScoringRule(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	uid, _ := currentUser(r)
	var req models.ScoringRuleInput
	if !decodeJSON(w, r, &req) {
		return
	}
	rule, err := h.Svc.SetScoringRule(r.Context(), id, req, uid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.WriteCreated(w, r, rule)
}

func (h *SeasonHandler) ListRuleVersions(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	versions, err := h.Svc.ListScoringRuleVersions(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, versions)
}
