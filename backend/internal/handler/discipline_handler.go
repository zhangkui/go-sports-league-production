package handler

import (
	"net/http"

	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/response"
	"github.com/goxm2/sports-league/internal/service"
)

// DisciplineHandler exposes discipline + appeal endpoints.
type DisciplineHandler struct {
	Svc *service.DisciplineService
}

func NewDisciplineHandler(svc *service.DisciplineService) *DisciplineHandler {
	return &DisciplineHandler{Svc: svc}
}

func (h *DisciplineHandler) List(w http.ResponseWriter, r *http.Request) {
	seasonID := int64(queryInt(r, "season_id", 0))
	playerID := int64(queryInt(r, "player_id", 0))
	status := queryStr(r, "status")
	p := parsePage(r)
	list, total, err := h.Svc.List(r.Context(), seasonID, playerID, status, p)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.WritePage(w, r, list, total, p.Page, p.PageSize)
}

func (h *DisciplineHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	d, err := h.Svc.Get(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, d)
}

func (h *DisciplineHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, _ := currentUser(r)
	var req models.CreateDisciplineRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	d, err := h.Svc.Create(r.Context(), req, uid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.WriteCreated(w, r, d)
}

func (h *DisciplineHandler) PlayerHistory(w http.ResponseWriter, r *http.Request) {
	pid, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	p := parsePage(r)
	list, total, err := h.Svc.List(r.Context(), 0, pid, "", p)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.WritePage(w, r, list, total, p.Page, p.PageSize)
}

func (h *DisciplineHandler) Overturn(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	d, err := h.Svc.Overturn(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, d)
}

// Appeals ----------------------------------------------------------------

func (h *DisciplineHandler) ListAppeals(w http.ResponseWriter, r *http.Request) {
	status := queryStr(r, "status")
	p := parsePage(r)
	list, total, err := h.Svc.ListAppeals(r.Context(), status, p)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.WritePage(w, r, list, total, p.Page, p.PageSize)
}

func (h *DisciplineHandler) CreateAppeal(w http.ResponseWriter, r *http.Request) {
	uid, _ := currentUser(r)
	var req models.CreateAppealRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	a, err := h.Svc.CreateAppeal(r.Context(), req, uid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.WriteCreated(w, r, a)
}

func (h *DisciplineHandler) ReviewAppeal(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	uid, _ := currentUser(r)
	var req models.AppealReviewInput
	if !decodeJSON(w, r, &req) {
		return
	}
	a, err := h.Svc.ReviewAppeal(r.Context(), id, uid, req)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, a)
}
