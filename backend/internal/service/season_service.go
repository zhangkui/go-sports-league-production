package service

import (
	"context"
	"strings"
	"time"

	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/errorsx"
)

// SeasonService handles season lifecycle and scoring rule versioning.
type SeasonService struct {
	*Deps
}

func NewSeasonService(d *Deps) *SeasonService { return &SeasonService{Deps: d} }

// Create a new season in draft status and seed an initial scoring rule.
func (s *SeasonService) Create(ctx context.Context, req models.CreateSeasonRequest, creatorID int64) (*models.Season, error) {
	if err := requireFields(map[string]string{
		"code": req.Code, "name": req.Name, "sport": req.Sport,
	}); err != nil {
		return nil, err
	}
	if err := codeFieldValid(req.Code); err != nil {
		return nil, err
	}
	if err := oneOf("format", defaultStr(req.Format, "round_robin"),
		"round_robin", "double_round", "mixed", "knockout"); err != nil {
		return nil, err
	}
	if err := oneOf("sport", req.Sport, "basketball", "football", "volleyball"); err != nil {
		return nil, err
	}
	se := &models.Season{
		Code: req.Code, Name: req.Name, Sport: req.Sport, Division: req.Division,
		TeamCount: req.TeamCount, Format: defaultStr(req.Format, "round_robin"),
		Status: models.SeasonStatusDraft, CurrentRuleVersion: 0, CreatedBy: creatorID,
	}
	se.StartDate = parseDate(req.StartDate)
	se.EndDate = parseDate(req.EndDate)
	se.RegistrationStart = parseDate(req.RegistrationStart)
	se.RegistrationEnd = parseDate(req.RegistrationEnd)
	if err := s.Seasons.Create(ctx, se); err != nil {
		return nil, err
	}
	// seed default scoring rule v1
	rule := &models.ScoringRule{
		SeasonID: se.ID, Version: 1, WinPoints: 3, DrawPoints: 1, LossPoints: 0,
		Tiebreakers: "points,goal_diff,goals_for,head_to_head,fair_play", CreatedBy: creatorID,
	}
	if err := s.Seasons.CreateScoringRule(ctx, nil, rule); err != nil {
		return nil, err
	}
	se.CurrentRuleVersion = 1
	return se, nil
}

// Get a season by id.
func (s *SeasonService) Get(ctx context.Context, id int64) (*models.Season, error) {
	se, err := s.Seasons.GetByID(ctx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("season", id)
	}
	return se, nil
}

// List seasons with filters.
func (s *SeasonService) List(ctx context.Context, status, sport string, p models.Pagination) ([]models.Season, int64, error) {
	return s.Seasons.List(ctx, status, sport, p, "id DESC")
}

// Update editable season fields.
func (s *SeasonService) Update(ctx context.Context, id int64, req models.UpdateSeasonRequest) (*models.Season, error) {
	se, err := s.Seasons.GetByID(ctx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("season", id)
	}
	if req.Name != nil {
		se.Name = *req.Name
	}
	if req.Division != nil {
		se.Division = *req.Division
	}
	if req.TeamCount != nil {
		se.TeamCount = *req.TeamCount
	}
	if req.Rounds != nil {
		se.Rounds = *req.Rounds
	}
	if req.StartDate != nil {
		se.StartDate = parseDate(*req.StartDate)
	}
	if req.EndDate != nil {
		se.EndDate = parseDate(*req.EndDate)
	}
	if req.RegistrationStart != nil {
		se.RegistrationStart = parseDate(*req.RegistrationStart)
	}
	if req.RegistrationEnd != nil {
		se.RegistrationEnd = parseDate(*req.RegistrationEnd)
	}
	if err := s.Seasons.Update(ctx, se); err != nil {
		return nil, errorsx.Internal("season update failed")
	}
	return s.Seasons.GetByID(ctx, id)
}

// ChangeStatus advances a season's lifecycle, enforcing allowed transitions.
// Only the status column changes; current_rule_version and other fields are preserved.
func (s *SeasonService) ChangeStatus(ctx context.Context, id int64, status string) (*models.Season, error) {
	se, err := s.Seasons.GetByID(ctx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("season", id)
	}
	if !validSeasonTransition(se.Status, status) {
		return nil, errorsx.BadRequest("invalid status transition: " + se.Status + " -> " + status)
	}
	if err := s.Seasons.SetStatus(ctx, id, status); err != nil {
		return nil, errorsx.Internal("status update failed")
	}
	se.Status = status
	return se, nil
}

// SetScoringRule creates a new versioned scoring rule for a season.
func (s *SeasonService) SetScoringRule(ctx context.Context, seasonID int64, req models.ScoringRuleInput, creatorID int64) (*models.ScoringRule, error) {
	se, err := s.Seasons.GetByID(ctx, seasonID)
	if err != nil {
		return nil, errorsx.NotFoundID("season", seasonID)
	}
	tiebreakers := req.Tiebreakers
	if tiebreakers == "" {
		tiebreakers = "points,goal_diff,goals_for,head_to_head,fair_play"
	}
	rule := &models.ScoringRule{
		SeasonID: seasonID, Version: se.CurrentRuleVersion + 1,
		WinPoints: req.WinPoints, DrawPoints: req.DrawPoints, LossPoints: req.LossPoints,
		Tiebreakers: tiebreakers, CreatedBy: creatorID,
	}
	candidate := rule.ActivationCandidate()
	if err := s.Seasons.CreateScoringRule(ctx, nil, &candidate); err != nil {
		return nil, err
	}
	return &candidate, nil
}

// GetActiveScoringRule returns the currently active rule for a season.
func (s *SeasonService) GetActiveScoringRule(ctx context.Context, seasonID int64) (*models.ScoringRule, error) {
	rule, err := s.Seasons.GetActiveScoringRule(ctx, seasonID)
	if err != nil {
		return models.EmptyScoringRule(seasonID), nil
	}
	if rule == nil || rule.ID == 0 {
		return models.EmptyScoringRule(seasonID), nil
	}
	return rule, nil
}

// ListScoringRuleVersions returns the version history.
func (s *SeasonService) ListScoringRuleVersions(ctx context.Context, seasonID int64) ([]models.ScoringRuleVersion, error) {
	return s.Seasons.ListScoringRuleVersions(ctx, seasonID)
}

func validSeasonTransition(from, to string) bool {
	if to == models.SeasonStatusCancelled {
		return true
	}
	if to == models.SeasonStatusArchived && from == models.SeasonStatusCompleted {
		return true
	}
	flow := map[string][]string{
		models.SeasonStatusDraft:        {models.SeasonStatusRegistration, models.SeasonStatusOpen},
		models.SeasonStatusRegistration: {models.SeasonStatusOpen, models.SeasonStatusOngoing},
		models.SeasonStatusOpen:         {models.SeasonStatusOngoing, models.SeasonStatusRegistration},
		models.SeasonStatusOngoing:      {models.SeasonStatusCompleted},
		models.SeasonStatusCompleted:    {models.SeasonStatusArchived},
	}
	allowed, ok := flow[from]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == to {
			return true
		}
	}
	return false
}

func parseDate(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &t
}

func defaultStr(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func oneOf(name, v string, allowed ...string) error {
	for _, a := range allowed {
		if v == a {
			return nil
		}
	}
	return errorsx.BadRequest(name + " must be one of: " + strings.Join(allowed, ", "))
}
