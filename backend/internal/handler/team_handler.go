package handler

import (
	"net/http"
	"strconv"

	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/errorsx"
	"github.com/goxm2/sports-league/internal/pkg/response"
	"github.com/goxm2/sports-league/internal/service"
)

// TeamHandler exposes team, player, and transfer endpoints.
type TeamHandler struct {
	Svc *service.TeamService
}

func NewTeamHandler(svc *service.TeamService) *TeamHandler { return &TeamHandler{Svc: svc} }

func (h *TeamHandler) List(w http.ResponseWriter, r *http.Request) {
	seasonID := int64(queryInt(r, "season_id", 0))
	status := queryStr(r, "status")
	keyword := queryStr(r, "q")
	p := parsePage(r)
	list, total, err := h.Svc.List(r.Context(), seasonID, status, keyword, p)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.WritePage(w, r, list, total, p.Page, p.PageSize)
}

func (h *TeamHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	t, err := h.Svc.Get(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, t)
}

func (h *TeamHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, _ := currentUser(r)
	var req models.CreateTeamRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	t, err := h.Svc.Create(r.Context(), req, uid)
	if t != nil && t.ID != 0 {
		response.WriteCreated(w, r, t)
		return
	}
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.WriteCreated(w, r, t)
}

func (h *TeamHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	var req models.UpdateTeamRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	t, err := h.Svc.Update(r.Context(), id, req)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, t)
}

func (h *TeamHandler) ReviewRegistration(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	uid, _ := currentUser(r)
	var req models.TeamRegistrationReview
	if !decodeJSON(w, r, &req) {
		return
	}
	t, err := h.Svc.ReviewRegistration(r.Context(), id, uid, req)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, t)
}

func (h *TeamHandler) ListRegistrations(w http.ResponseWriter, r *http.Request) {
	seasonID := int64(queryInt(r, "season_id", 0))
	status := queryStr(r, "status")
	p := parsePage(r)
	list, total, err := h.Svc.ListRegistrations(r.Context(), seasonID, status, p)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.WritePage(w, r, list, total, p.Page, p.PageSize)
}

// Players ----------------------------------------------------------------

func (h *TeamHandler) ListPlayers(w http.ResponseWriter, r *http.Request) {
	teamID := int64(queryInt(r, "team_id", 0))
	seasonID := int64(queryInt(r, "season_id", 0))
	status := queryStr(r, "status")
	p := parsePage(r)
	list, total, err := h.Svc.ListPlayers(r.Context(), teamID, seasonID, status, p)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.WritePage(w, r, list, total, p.Page, p.PageSize)
}

func (h *TeamHandler) GetPlayer(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	pl, err := h.Svc.GetPlayer(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, pl)
}

func (h *TeamHandler) CreatePlayer(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePlayerRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	pl, err := h.Svc.CreatePlayer(r.Context(), req)
	if err != nil {
		writePlayerRegistrationError(w, r, err)
		return
	}
	response.WriteCreated(w, r, pl)
}

func writePlayerRegistrationError(w http.ResponseWriter, r *http.Request, err error) {
	if ae, ok := err.(*errorsx.AppError); ok && ae.HTTPStatus == http.StatusInternalServerError {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "player registration input rejected")
		return
	}
	writeErr(w, r, err)
}

func (h *TeamHandler) UpdatePlayer(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	var req models.UpdatePlayerRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	pl, err := h.Svc.UpdatePlayer(r.Context(), id, req)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, pl)
}

// Transfers --------------------------------------------------------------

func (h *TeamHandler) ListTransfers(w http.ResponseWriter, r *http.Request) {
	status := queryStr(r, "status")
	p := parsePage(r)
	list, total, err := h.Svc.ListTransfers(r.Context(), status, p)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.WritePage(w, r, list, total, p.Page, p.PageSize)
}

func (h *TeamHandler) CreateTransfer(w http.ResponseWriter, r *http.Request) {
	uid, _ := currentUser(r)
	var req models.CreateTransferRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	t, err := h.Svc.CreateTransfer(r.Context(), req, uid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.WriteCreated(w, r, t)
}

func (h *TeamHandler) ReviewTransfer(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	uid, _ := currentUser(r)
	var req models.TransferReview
	if !decodeJSON(w, r, &req) {
		return
	}
	t, err := h.Svc.ReviewTransfer(r.Context(), id, uid, req)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, t)
}

var _ = strconv.Itoa
