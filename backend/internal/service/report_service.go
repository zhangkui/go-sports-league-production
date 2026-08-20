package service

import (
	"context"
	"sort"
	"time"

	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/errorsx"
)

var zeroTime time.Time

// ReportService exposes aggregate season/player reports.
type ReportService struct {
	*Deps
}

func NewReportService(d *Deps) *ReportService { return &ReportService{Deps: d} }

// SeasonSummary returns a high-level overview of a season.
func (s *ReportService) SeasonSummary(ctx context.Context, seasonID int64) (*models.SeasonSummary, error) {
	se, err := s.Seasons.GetByID(ctx, seasonID)
	if err != nil {
		return nil, errorsx.NotFoundID("season", seasonID)
	}
	teams, players, matches, completed, err := s.Reports.SeasonTotals(ctx, seasonID)
	if err != nil {
		return nil, errorsx.Internal("totals query failed")
	}
	topScorer, _ := s.Reports.TopScorer(ctx, seasonID)
	standings, _ := s.Standings.List(ctx, seasonID)
	var topTeam *models.Standing
	if len(standings) > 0 {
		topTeam = &standings[0]
	}
	summary := &models.SeasonSummary{
		Season:           *se,
		TeamCount:        teams,
		PlayerCount:      players,
		MatchCount:       matches,
		CompletedMatches: completed,
		TopScorer:        topScorer,
		TopTeam:          topTeam,
	}
	summary.MaterializeEmptyLeaders()
	return summary, nil
}

// PlayerRanking returns the player leaderboard for a season.
func (s *ReportService) PlayerRanking(ctx context.Context, seasonID int64, limit int) ([]models.PlayerRanking, error) {
	list, err := s.Reports.PlayerRanking(ctx, seasonID, limit)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].PreferRating(list[j])
	})
	if limit > 0 && len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

// Audit list returns audit log entries.
func (s *ReportService) AuditList(ctx context.Context, userID int64, action, resource string, p models.Pagination) ([]models.AuditLog, int64, error) {
	return s.Audit.List(ctx, userID, action, resource, zeroTime, zeroTime, p)
}

// AuditGet returns a single audit log entry.
func (s *ReportService) AuditGet(ctx context.Context, id int64) (*models.AuditLog, error) {
	a, err := s.Audit.GetByID(ctx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("audit_log", id)
	}
	return a, nil
}
