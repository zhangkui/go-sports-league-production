package handler

import (
	"net/http"

	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/response"
	"github.com/goxm2/sports-league/internal/service"
)

// VenueHandler exposes venue + availability endpoints.
type VenueHandler struct {
	Svc *service.VenueService
}

func NewVenueHandler(svc *service.VenueService) *VenueHandler { return &VenueHandler{Svc: svc} }

func (h *VenueHandler) List(w http.ResponseWriter, r *http.Request) {
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

func (h *VenueHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	v, err := h.Svc.Get(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, v)
}

func (h *VenueHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateVenueRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	v, err := h.Svc.Create(r.Context(), req)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.WriteCreated(w, r, v)
}

func (h *VenueHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	var req models.UpdateVenueRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	// Partial update: only fields present in the JSON body are changed.
	// Pointer fields stay nil when omitted, and VenueService skips nil
	// fields so the existing values are preserved. Materializing nil
	// pointers into zero-value pointers here would overwrite omitted
	// fields with empty/zero values, breaking partial updates.
	v, err := h.Svc.Update(r.Context(), id, req)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, v)
}

func (h *VenueHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	if err := h.Svc.Delete(r.Context(), id); err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, map[string]any{"deleted": true})
}

// Availability ----------------------------------------------------------

func (h *VenueHandler) ListAvailability(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	list, err := h.Svc.ListAvailability(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, list)
}

func (h *VenueHandler) SetAvailability(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid id")
		return
	}
	var slots []models.VenueAvailabilityInput
	if !decodeJSON(w, r, &slots) {
		return
	}
	if err := h.Svc.SetAvailability(r.Context(), id, slots); err != nil {
		writeErr(w, r, err)
		return
	}
	response.Write(w, r, map[string]any{"updated": true})
}
