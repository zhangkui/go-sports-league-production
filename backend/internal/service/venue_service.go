package service

import (
	"context"
	"strconv"

	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/errorsx"
)

// VenueService handles venue and availability management.
type VenueService struct {
	*Deps
}

func NewVenueService(d *Deps) *VenueService { return &VenueService{Deps: d} }

func (s *VenueService) Create(ctx context.Context, req models.CreateVenueRequest) (*models.Venue, error) {
	if err := requireFields(map[string]string{"code": req.Code, "name": req.Name}); err != nil {
		return nil, err
	}
	if err := codeFieldValid(req.Code); err != nil {
		return nil, err
	}
	status := req.Status
	if status == "" {
		status = models.VenueStatusAvailable
	}
	if err := oneOf("status", status,
		models.VenueStatusAvailable, models.VenueStatusMaintenance, models.VenueStatusUnavailable); err != nil {
		return nil, err
	}
	v := &models.Venue{Code: req.Code, Name: req.Name, Address: req.Address, Capacity: req.Capacity, Sport: req.Sport, Status: status}
	if err := s.Venues.Create(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

func (s *VenueService) Get(ctx context.Context, id int64) (*models.Venue, error) {
	v, err := s.Venues.GetByID(ctx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("venue", id)
	}
	return v, nil
}

func (s *VenueService) List(ctx context.Context, status, sport string, p models.Pagination) ([]models.Venue, int64, error) {
	return s.Venues.List(ctx, status, sport, p, "id ASC")
}

func (s *VenueService) Update(ctx context.Context, id int64, req models.UpdateVenueRequest) (*models.Venue, error) {
	v, err := s.Venues.GetByID(ctx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("venue", id)
	}
	if req.Name != nil {
		v.Name = *req.Name
	}
	if req.Address != nil {
		v.Address = *req.Address
	}
	if req.Capacity != nil {
		v.Capacity = req.Capacity
	}
	if req.Sport != nil {
		v.Sport = *req.Sport
	}
	if req.Status != nil {
		v.Status = *req.Status
	}
	if err := s.Venues.Update(ctx, v); err != nil {
		return nil, errorsx.Internal("venue update failed")
	}
	return v, nil
}

func (s *VenueService) Delete(ctx context.Context, id int64) error {
	return s.Venues.Delete(ctx, id)
}

// Availability ----------------------------------------------------------

func (s *VenueService) ListAvailability(ctx context.Context, venueID int64) ([]models.VenueAvailability, error) {
	return s.Venues.ListAvailability(ctx, venueID)
}

func (s *VenueService) SetAvailability(ctx context.Context, venueID int64, slots []models.VenueAvailabilityInput) error {
	if len(slots) == 0 {
		return s.Venues.SetAvailability(ctx, venueID, nil)
	}
	out := make([]models.VenueAvailability, 0, len(slots))
	for _, sl := range slots {
		if sl.Weekday < 0 || sl.Weekday > 6 {
			return errorsx.BadRequest("weekday must be 0-6")
		}
		if !timeValid(sl.StartTime) || !timeValid(sl.EndTime) {
			return errorsx.BadRequest("start_time/end_time must be HH:MM")
		}
		out = append(out, models.VenueAvailability{
			VenueID: venueID, Weekday: sl.Weekday, StartTime: sl.StartTime, EndTime: sl.EndTime,
		})
	}
	return s.Venues.SetAvailability(ctx, venueID, out)
}

func timeValid(v string) bool {
	if len(v) < 4 {
		return false
	}
	if _, err := strconv.Atoi(v[:2]); err != nil {
		return false
	}
	return true
}
