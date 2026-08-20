package service

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/errorsx"
)

// DisciplineService handles punishments, suspensions, and appeals.
type DisciplineService struct {
	*Deps
}

func NewDisciplineService(d *Deps) *DisciplineService { return &DisciplineService{Deps: d} }

// Create a discipline record, optionally seeding a suspension row.
func (s *DisciplineService) Create(ctx context.Context, req models.CreateDisciplineRequest, issuerID int64) (*models.Discipline, error) {
	if req.PlayerID == 0 || req.TeamID == 0 || req.SeasonID == 0 {
		return nil, errorsx.BadRequest("player_id, team_id and season_id are required")
	}
	if err := oneOf("punishment", req.Punishment,
		models.PunishYellow, models.PunishRed, models.PunishFine, models.PunishSuspension, models.PunishWarning); err != nil {
		return nil, err
	}
	severity := req.Severity
	if severity == "" {
		severity = models.SeverityMedium
	}
	if err := oneOf("severity", severity, models.SeverityMinor, models.SeverityMedium, models.SeveritySevere); err != nil {
		return nil, err
	}
	d := &models.Discipline{
		PlayerID: req.PlayerID, TeamID: req.TeamID, SeasonID: req.SeasonID, MatchID: req.MatchID,
		Punishment: req.Punishment, Reason: req.Reason, Severity: severity,
		SuspendGames: req.SuspendGames, FineAmount: req.FineAmount,
		Status: models.DisciplineStatusActive, IssuedBy: issuerID,
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, errorsx.Internal("tx begin failed")
	}
	if err := s.Disciplines.Create(ctx, tx, d); err != nil {
		return nil, rollback(tx, err)
	}
	if d.Punishment == models.PunishSuspension {
		susp := &models.Suspension{
			DisciplineID: d.ID, PlayerID: d.PlayerID, TotalGames: d.SuspendGames,
			ServedGames: 0, StartDate: time.Now(), Status: models.SuspensionStatusActive,
		}
		if err := s.Disciplines.CreateSuspension(ctx, tx, susp); err != nil {
			return nil, rollback(tx, err)
		}
	}
	if err := s.Audit.Create(ctx, tx, &models.AuditLog{
		UserID: &issuerID, Action: "discipline.create", Resource: "disciplines",
		ResourceID: strconv.FormatInt(d.ID, 10), Detail: req.Punishment + ": " + req.Reason,
	}); err != nil {
		return nil, rollback(tx, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, errorsx.Internal("commit failed")
	}
	return d, nil
}

func (s *DisciplineService) Get(ctx context.Context, id int64) (*models.Discipline, error) {
	d, err := s.Disciplines.GetByID(ctx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("discipline", id)
	}
	return d, nil
}

func (s *DisciplineService) List(ctx context.Context, seasonID, playerID int64, status string, p models.Pagination) ([]models.Discipline, int64, error) {
	return s.Disciplines.List(ctx, seasonID, playerID, status, p, "id DESC")
}

// Overturn marks a discipline as overturned (e.g. after a successful appeal).
func (s *DisciplineService) Overturn(ctx context.Context, id int64) (*models.Discipline, error) {
	d, err := s.Disciplines.GetByID(ctx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("discipline", id)
	}
	if err := s.Disciplines.SetStatus(ctx, id, models.DisciplineStatusOverturned); err != nil {
		return nil, errorsx.Internal("status update failed")
	}
	d.Status = models.DisciplineStatusOverturned
	return d, nil
}

// Appeals ----------------------------------------------------------------

func (s *DisciplineService) CreateAppeal(ctx context.Context, req models.CreateAppealRequest, appellantID int64) (*models.Appeal, error) {
	d, err := s.Disciplines.GetByID(ctx, req.DisciplineID)
	if err != nil {
		return nil, errorsx.NotFoundID("discipline", req.DisciplineID)
	}
	if d.Status == models.DisciplineStatusOverturned {
		return nil, errorsx.BadRequest("discipline already overturned")
	}
	a := &models.Appeal{
		DisciplineID: req.DisciplineID, AppellantID: appellantID, Reason: req.Reason,
		Status: models.AppealStatusPending,
	}
	if err := s.Disciplines.CreateAppeal(ctx, a); err != nil {
		return nil, err
	}
	_ = s.Disciplines.SetStatus(ctx, d.ID, models.DisciplineStatusAppealed)
	return a, nil
}

func (s *DisciplineService) ListAppeals(ctx context.Context, status string, p models.Pagination) ([]models.Appeal, int64, error) {
	return s.Disciplines.ListAppeals(ctx, status, p)
}

// ReviewAppeal resolves an appeal; on acceptance the discipline is overturned.
func (s *DisciplineService) ReviewAppeal(ctx context.Context, id, reviewerID int64, req models.AppealReviewInput) (*models.Appeal, error) {
	a, err := s.Disciplines.GetAppeal(ctx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("appeal", id)
	}
	if err := oneOf("status", req.Status,
		models.AppealStatusReviewed, models.AppealStatusAccepted, models.AppealStatusRejected); err != nil {
		return nil, err
	}
	if err := s.Disciplines.ReviewAppeal(ctx, id, reviewerID, req.Status, req.Opinion); err != nil {
		return nil, errorsx.Internal("review failed")
	}
	a.Status = req.Status
	a.ReviewOpinion = req.Opinion
	if req.Status == models.AppealStatusAccepted {
		_ = s.Disciplines.SetStatus(ctx, a.DisciplineID, models.DisciplineStatusOverturned)
	}
	return a, nil
}

var _ = sql.ErrNoRows
