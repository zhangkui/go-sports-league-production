package handler

import (
	"net/http"

	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/response"
	"github.com/goxm2/sports-league/internal/service"
)

// MatchHandler exposes match, event, and player-stat endpoints.
type MatchHandler struct {
	Svc *service.MatchService
}

func NewMatchHandler(svc *service.MatchService) *MatchHandler { return &MatchHandler{Svc: svc} }

func (h *MatchHandler) List(w http.ResponseWriter, r *http.Request) {
	seasonID := int64(queryInt(r, "season_id", 0))
	status := queryStr(r, "status")
	teamID := int64(queryInt(r, "team_id", 0))
	p := parsePage(r)
	list, total, err := h.Svc.List(r.Context(), seasonID, status, teamID, p)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.WritePage(w, r, list, total, p.Page, p.PageSize)
}

func (h *MatchHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	m, err := h.Svc.Get(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, m)
}

func (h *MatchHandler) Record(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	var req models.RecordMatchRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	m, err := h.Svc.Record(r.Context(), id, req)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, m)
}

func (h *MatchHandler) SetStatus(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	status := queryStr(r, "status")
	m, err := h.Svc.SetStatus(r.Context(), id, status)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, m)
}

func (h *MatchHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	uid, _ := currentUser(r)
	m, err := h.Svc.Confirm(r.Context(), id, uid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, m)
}

func (h *MatchHandler) Dispute(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	m, err := h.Svc.Dispute(r.Context(), id, queryStr(r, "reason"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, m)
}

// Events ----------------------------------------------------------------

func (h *MatchHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	list, err := h.Svc.ListEvents(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, list)
}

func (h *MatchHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	uid, _ := currentUser(r)
	var req models.CreateMatchEventRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	e, err := h.Svc.CreateEvent(r.Context(), id, req, uid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.WriteCreated(w, r, e)
}

func (h *MatchHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	eid, err := pathID(r, "event_id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	var req models.CreateMatchEventRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	e, err := h.Svc.UpdateEvent(r.Context(), eid, req)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, e)
}

func (h *MatchHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	eid, err := pathID(r, "event_id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	if err := h.Svc.DeleteEvent(r.Context(), eid); err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, map[string]any{"deleted": true})
}

// Player stats ----------------------------------------------------------

func (h *MatchHandler) ListPlayerStats(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	list, err := h.Svc.ListPlayerStats(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, list)
}

func (h *MatchHandler) UpsertPlayerStat(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	var req models.CreatePlayerStatRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	st, err := h.Svc.UpsertPlayerStat(r.Context(), id, req)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, st)
}
