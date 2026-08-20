package service

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/errorsx"
	"github.com/goxm2/sports-league/internal/pkg/logx"
)

// MatchService handles match records, events, player stats, and the
// confirm→standing-update transaction.
type MatchService struct {
	*Deps
}

func NewMatchService(d *Deps) *MatchService { return &MatchService{Deps: d} }

// GetOrMaterialise returns the match for a schedule, creating it lazily.
func (s *MatchService) GetOrMaterialise(ctx context.Context, scheduleID int64) (*models.Match, error) {
	sc, err := s.Schedules.GetByID(ctx, scheduleID)
	if err != nil {
		return nil, errorsx.NotFoundID("schedule", scheduleID)
	}
	return s.Matches.EnsureForSchedule(ctx, sc)
}

func (s *MatchService) Get(ctx context.Context, id int64) (*models.Match, error) {
	m, err := s.Matches.GetByID(ctx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("match", id)
	}
	return m, nil
}

func (s *MatchService) List(ctx context.Context, seasonID int64, status string, teamID int64, p models.Pagination) ([]models.Match, int64, error) {
	return s.Matches.List(ctx, seasonID, status, teamID, p, "match_date DESC, id DESC")
}

// Record updates match scores/status (referee / league_admin).
func (s *MatchService) Record(ctx context.Context, id int64, req models.RecordMatchRequest) (*models.Match, error) {
	m, err := s.Matches.GetByID(ctx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("match", id)
	}
	if req.HomeScore != nil {
		m.HomeScore = req.HomeScore
	}
	if req.AwayScore != nil {
		m.AwayScore = req.AwayScore
	}
	if req.HomeHalfScore != nil {
		m.HomeHalfScore = req.HomeHalfScore
	}
	if req.AwayHalfScore != nil {
		m.AwayHalfScore = req.AwayHalfScore
	}
	if req.RefereeID != nil {
		m.RefereeID = req.RefereeID
	}
	if req.RecorderID != nil {
		m.RecorderID = req.RecorderID
	}
	if req.DurationMin != nil {
		m.DurationMin = req.DurationMin
	}
	if req.Status != nil {
		m.Status = *req.Status
	}
	if err := s.Matches.UpdateRecord(ctx, nil, m); err != nil {
		return nil, errorsx.Internal("record update failed")
	}
	return m, nil
}

// SetStatus transitions a match status.
func (s *MatchService) SetStatus(ctx context.Context, id int64, status string) (*models.Match, error) {
	m, err := s.Matches.GetByID(ctx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("match", id)
	}
	if err := oneOf("status", status,
		models.MatchStatusScheduled, models.MatchStatusConfirmed, models.MatchStatusInProgress,
		models.MatchStatusCompleted, models.MatchStatusCancelled, models.MatchStatusPostponed); err != nil {
		return nil, err
	}
	if err := s.Matches.SetStatus(ctx, id, status); err != nil {
		return nil, errorsx.Internal("status update failed")
	}
	m.Status = status
	return m, nil
}

// Confirm runs the score-confirmation → standing-update → snapshot transaction.
func (s *MatchService) Confirm(ctx context.Context, id, confirmerID int64) (*models.Match, error) {
	m, err := s.Matches.GetByID(ctx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("match", id)
	}
	if m.ConfirmStatus == models.ConfirmStatusConfirmed {
		return m, nil
	}
	if m.HomeScore == nil || m.AwayScore == nil {
		return nil, errorsx.BadRequest("match scores must be recorded before confirming")
	}
	if m.Status != models.MatchStatusCompleted {
		m.Status = models.MatchStatusCompleted
	}
	home, away := *m.HomeScore, *m.AwayScore
	if err := s.applyConfirmTx(ctx, m, home, away, confirmerID); err != nil {
		return nil, err
	}
	m.ConfirmStatus = models.ConfirmStatusConfirmed
	m.ConfirmedBy = &confirmerID
	return m, nil
}

// applyConfirmTx executes the standing update transaction described in INV-05.
func (s *MatchService) applyConfirmTx(ctx context.Context, m *models.Match, homeGoals, awayGoals int, confirmerID int64) error {
	rule, err := s.Seasons.GetActiveScoringRule(ctx, m.SeasonID)
	if err != nil {
		return errorsx.NotFoundID("scoring rule", m.SeasonID)
	}
	// fair-play delta from cards in this match
	homeFair, awayFair := s.fairPlayDelta(ctx, m.ID, m.HomeTeamID), s.fairPlayDelta(ctx, m.ID, m.AwayTeamID)

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return errorsx.Internal("tx begin failed")
	}
	if err := s.Matches.SetConfirmStatus(ctx, tx, m.ID, models.ConfirmStatusConfirmed, confirmerID); err != nil {
		return rollback(tx, err)
	}
	if err := s.Matches.UpdateRecord(ctx, tx, m); err != nil {
		return rollback(tx, err)
	}
	if err := s.Standings.EnsureRow(ctx, tx, m.SeasonID, m.HomeTeamID); err != nil {
		return rollback(tx, err)
	}
	if err := s.Standings.EnsureRow(ctx, tx, m.SeasonID, m.AwayTeamID); err != nil {
		return rollback(tx, err)
	}
	if err := s.Standings.Apply(ctx, tx, m.SeasonID, m.HomeTeamID, homeGoals, awayGoals, rule.WinPoints, rule.DrawPoints, rule.LossPoints, homeFair); err != nil {
		return rollback(tx, err)
	}
	if err := s.Standings.Apply(ctx, tx, m.SeasonID, m.AwayTeamID, awayGoals, homeGoals, rule.WinPoints, rule.DrawPoints, rule.LossPoints, awayFair); err != nil {
		return rollback(tx, err)
	}
	// recompute ranks and snapshot
	ranked, err := s.Standings.Rank(ctx, m.SeasonID, rule.TiebreakerOrder())
	if err != nil {
		return rollback(tx, err)
	}
	round := s.matchRound(ctx, m.ID)
	if err := s.Standings.Snapshot(ctx, tx, m.SeasonID, round, ranked); err != nil {
		return rollback(tx, err)
	}
	// advance suspensions served games for players in this match
	s.advanceSuspensions(ctx, m.ID)
	if err := s.Audit.Create(ctx, tx, &models.AuditLog{
		UserID: &confirmerID, Action: "match.confirm", Resource: "matches", ResourceID: strconv.FormatInt(m.ID, 10),
		Detail: fmt.Sprintf("confirmed %d:%d", homeGoals, awayGoals),
	}); err != nil {
		return rollback(tx, err)
	}
	if err := tx.Commit(); err != nil {
		return errorsx.Internal("commit failed")
	}
	logx.Info("match confirmed & standings updated", "match_id", m.ID, "home", homeGoals, "away", awayGoals)
	return nil
}

// fairPlayDelta computes the fair-play penalty for a team in a match (yellow=1, red=3).
func (s *MatchService) fairPlayDelta(ctx context.Context, matchID, teamID int64) int {
	events, _ := s.Matches.ListEvents(ctx, matchID)
	delta := 0
	for _, e := range events {
		if e.TeamID != teamID {
			continue
		}
		switch e.EventType {
		case models.EventYellowCard:
			delta++
		case models.EventRedCard:
			delta += 3
		}
	}
	return delta
}

// matchRound returns the schedule round for a match.
func (s *MatchService) matchRound(ctx context.Context, matchID int64) int {
	m, err := s.Matches.GetByID(ctx, matchID)
	if err != nil {
		return 0
	}
	sc, err := s.Schedules.GetByID(ctx, m.ScheduleID)
	if err != nil {
		return 0
	}
	return sc.Round
}

// advanceSuspensions increments served-games for active suspensions of players in the match.
func (s *MatchService) advanceSuspensions(ctx context.Context, matchID int64) {
	stats, _ := s.Matches.ListPlayerStats(ctx, matchID)
	for _, st := range stats {
		// find active discipline with suspension for this player
		list, _, _ := s.Disciplines.List(ctx, 0, st.PlayerID, models.DisciplineStatusActive, models.Pagination{Page: 1, PageSize: 50}, "id DESC")
		for _, d := range list {
			if d.Punishment == models.PunishSuspension {
				susp, err := s.Disciplines.GetSuspension(ctx, d.ID)
				if err == nil && susp.Status == models.SuspensionStatusActive {
					_ = s.Disciplines.IncrementServed(ctx, d.ID)
					if susp.ServedGames+1 >= susp.TotalGames {
						_ = s.Disciplines.SetStatus(ctx, d.ID, models.DisciplineStatusServed)
					}
				}
			}
		}
	}
}

// Dispute marks a match as disputed.
func (s *MatchService) Dispute(ctx context.Context, id int64, reason string) (*models.Match, error) {
	m, err := s.Matches.GetByID(ctx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("match", id)
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, errorsx.Internal("tx begin failed")
	}
	if err := s.Matches.SetConfirmStatus(ctx, tx, id, models.ConfirmStatusDisputed, 0); err != nil {
		return nil, rollback(tx, err)
	}
	if err := s.Audit.Create(ctx, tx, &models.AuditLog{
		Action: "match.dispute", Resource: "matches", ResourceID: strconv.FormatInt(id, 10), Detail: reason,
	}); err != nil {
		return nil, rollback(tx, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, errorsx.Internal("commit failed")
	}
	m.ConfirmStatus = models.ConfirmStatusDisputed
	return m, nil
}

// Events -----------------------------------------------------------------

func (s *MatchService) CreateEvent(ctx context.Context, matchID int64, req models.CreateMatchEventRequest, creatorID int64) (*models.MatchEvent, error) {
	if _, err := s.Matches.GetByID(ctx, matchID); err != nil {
		return nil, errorsx.NotFoundID("match", matchID)
	}
	if err := oneOf("event_type", req.EventType,
		models.EventGoal, models.EventAssist, models.EventFoul, models.EventYellowCard,
		models.EventRedCard, models.EventSubstitution, models.EventInjury); err != nil {
		return nil, err
	}
	e := &models.MatchEvent{
		MatchID: matchID, TeamID: req.TeamID, PlayerID: req.PlayerID, EventType: req.EventType,
		Minute: req.Minute, Description: req.Description, ConfirmStatus: models.ConfirmStatusPending, CreatedBy: creatorID,
	}
	if err := s.Matches.CreateEvent(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *MatchService) ListEvents(ctx context.Context, matchID int64) ([]models.MatchEvent, error) {
	return s.Matches.ListEvents(ctx, matchID)
}

func (s *MatchService) UpdateEvent(ctx context.Context, id int64, req models.CreateMatchEventRequest) (*models.MatchEvent, error) {
	e, err := s.Matches.GetEvent(ctx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("event", id)
	}
	e.TeamID = req.TeamID
	e.PlayerID = req.PlayerID
	e.EventType = req.EventType
	e.Minute = req.Minute
	e.Description = req.Description
	if err := s.Matches.UpdateEvent(ctx, e); err != nil {
		return nil, errorsx.Internal("event update failed")
	}
	return e, nil
}

func (s *MatchService) DeleteEvent(ctx context.Context, id int64) error {
	if _, err := s.Matches.GetEvent(ctx, id); err != nil {
		return errorsx.NotFoundID("event", id)
	}
	return s.Matches.DeleteEvent(ctx, id)
}

// Player stats -----------------------------------------------------------

func (s *MatchService) UpsertPlayerStat(ctx context.Context, matchID int64, req models.CreatePlayerStatRequest) (*models.PlayerMatchStat, error) {
	if _, err := s.Matches.GetByID(ctx, matchID); err != nil {
		return nil, errorsx.NotFoundID("match", matchID)
	}
	st := &models.PlayerMatchStat{
		MatchID: matchID, PlayerID: req.PlayerID, TeamID: req.TeamID,
		IsStarter: req.IsStarter, PlayedMin: req.PlayedMin, Position: req.Position, Rating: req.Rating,
	}
	if err := s.Matches.UpsertPlayerStat(ctx, st); err != nil {
		return nil, errorsx.Internal("stat upsert failed")
	}
	return st, nil
}

func (s *MatchService) ListPlayerStats(ctx context.Context, matchID int64) ([]models.PlayerMatchStat, error) {
	return s.Matches.ListPlayerStats(ctx, matchID)
}

var _ = sql.ErrNoRows
