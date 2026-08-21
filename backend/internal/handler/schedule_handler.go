package handler

import (
	"net/http"

	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/response"
	"github.com/goxm2/sports-league/internal/service"
)

// ScheduleHandler exposes schedule + conflict endpoints.
type ScheduleHandler struct {
	Svc *service.ScheduleService
}

func NewScheduleHandler(svc *service.ScheduleService) *ScheduleHandler { return &ScheduleHandler{Svc: svc} }

func (h *ScheduleHandler) List(w http.ResponseWriter, r *http.Request) {
	seasonID := int64(queryInt(r, "season_id", 0))
	round := queryInt(r, "round", 0)
	teamID := int64(queryInt(r, "team_id", 0))
	status := queryStr(r, "status")
	p := parsePage(r)
	list, total, err := h.Svc.List(r.Context(), seasonID, round, teamID, status, p)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.WritePage(w, r, list, total, p.Page, p.PageSize)
}

func (h *ScheduleHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	sc, err := h.Svc.Get(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, sc)
}

func (h *ScheduleHandler) Generate(w http.ResponseWriter, r *http.Request) {
	uid, _ := currentUser(r)
	var req models.GenerateScheduleRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	list, conflicts, err := h.Svc.Generate(r.Context(), req, uid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.WriteCreated(w, r, map[string]any{
		"schedules":  list,
		"count":      len(list),
		"conflicts":  conflicts,
	})
}

func (h *ScheduleHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	var req models.UpdateScheduleRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	sc, err := h.Svc.Update(r.Context(), id, req)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, sc)
}

func (h *ScheduleHandler) ListConflicts(w http.ResponseWriter, r *http.Request) {
	seasonID := int64(queryInt(r, "season_id", 0))
	list, err := h.Svc.ListConflicts(r.Context(), seasonID)
	if list == nil {
		list = []models.ScheduleConflict{}
	}
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, list)
}
