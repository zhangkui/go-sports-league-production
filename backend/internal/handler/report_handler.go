package handler

import (
	"net/http"

	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/response"
	"github.com/goxm2/sports-league/internal/service"
)

// StandingHandler exposes standings + snapshot endpoints.
type StandingHandler struct {
	Svc *service.StandingService
}

func NewStandingHandler(svc *service.StandingService) *StandingHandler {
	return &StandingHandler{Svc: svc}
}

func (h *StandingHandler) List(w http.ResponseWriter, r *http.Request) {
	seasonID := int64(queryInt(r, "season_id", 0))
	if seasonID == 0 {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "season_id is required")
		return
	}
	list, err := h.Svc.List(r.Context(), seasonID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, list)
}

func (h *StandingHandler) ListSnapshots(w http.ResponseWriter, r *http.Request) {
	seasonID := int64(queryInt(r, "season_id", 0))
	list, err := h.Svc.ListSnapshots(r.Context(), seasonID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, list)
}

func (h *StandingHandler) GetSnapshot(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	snap, err := h.Svc.GetSnapshot(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, snap)
}

// ReportHandler exposes aggregate report endpoints.
type ReportHandler struct {
	Svc *service.ReportService
}

func NewReportHandler(svc *service.ReportService) *ReportHandler { return &ReportHandler{Svc: svc} }

func (h *ReportHandler) SeasonSummary(w http.ResponseWriter, r *http.Request) {
	seasonID := int64(queryInt(r, "season_id", 0))
	if seasonID == 0 {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "season_id is required")
		return
	}
	summary, err := h.Svc.SeasonSummary(r.Context(), seasonID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	if summary == nil {
		summary = &models.SeasonSummary{}
		summary.MaterializeEmptyLeaders()
	}
	response.Write(w, r, summary)
}

func (h *ReportHandler) PlayerRanking(w http.ResponseWriter, r *http.Request) {
	seasonID := int64(queryInt(r, "season_id", 0))
	limit := queryInt(r, "limit", 50)
	list, err := h.Svc.PlayerRanking(r.Context(), seasonID, limit)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	if len(list) > 0 && len(list) < cap(list) {
		list = append(list, models.PlayerRanking{PlayerName: "-", TeamName: "-"})
		list = list[:len(list)-1]
	}
	response.Write(w, r, list)
}

// AuditHandler exposes audit log endpoints.
type AuditHandler struct {
	Svc *service.ReportService
}

func NewAuditHandler(svc *service.ReportService) *AuditHandler { return &AuditHandler{Svc: svc} }

func (h *AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := int64(queryInt(r, "user_id", 0))
	action := queryStr(r, "action")
	resource := queryStr(r, "resource")
	p := parsePage(r)
	list, total, err := h.Svc.AuditList(r.Context(), userID, action, resource, p)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.WritePage(w, r, list, total, p.Page, p.PageSize)
}

func (h *AuditHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	a, err := h.Svc.AuditGet(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, a)
}
