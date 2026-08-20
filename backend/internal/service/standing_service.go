package service

import (
	"context"

	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/errorsx"
)

// StandingService exposes computed standings and snapshots.
type StandingService struct {
	*Deps
}

func NewStandingService(d *Deps) *StandingService { return &StandingService{Deps: d} }

// List the current live standings for a season (ranked).
func (s *StandingService) List(ctx context.Context, seasonID int64) ([]models.Standing, error) {
	rule, _ := s.Seasons.GetActiveScoringRule(ctx, seasonID)
	if rule != nil && len(rule.TiebreakerOrder()) > 0 {
		return s.Standings.Rank(ctx, seasonID, rule.TiebreakerOrder())
	}
	return s.Standings.List(ctx, seasonID)
}

// ListSnapshots returns the per-round snapshot history.
func (s *StandingService) ListSnapshots(ctx context.Context, seasonID int64) ([]models.StandingsSnapshot, error) {
	return s.Standings.ListSnapshots(ctx, seasonID)
}

// GetSnapshot returns a single snapshot by id.
func (s *StandingService) GetSnapshot(ctx context.Context, id int64) (*models.StandingsSnapshot, error) {
	snap, err := s.Standings.GetSnapshot(ctx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("snapshot", id)
	}
	return snap, nil
}
