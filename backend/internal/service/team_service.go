package service

import (
	"context"

	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/errorsx"
)

// TeamService handles team registration, players, and transfers.
type TeamService struct {
	*Deps
}

func NewTeamService(d *Deps) *TeamService { return &TeamService{Deps: d} }

// Create a team (status pending) within a season.
func (s *TeamService) Create(ctx context.Context, req models.CreateTeamRequest, creatorID int64) (*models.Team, error) {
	if err := requireFields(map[string]string{
		"code": req.Code, "name": req.Name,
	}); err != nil {
		return nil, err
	}
	if err := codeFieldValid(req.Code); err != nil {
		return nil, err
	}
	if req.SeasonID == 0 {
		return nil, errorsx.BadRequest("season_id is required")
	}
	if _, err := s.Seasons.GetByID(ctx, req.SeasonID); err != nil {
		return nil, errorsx.NotFoundID("season", req.SeasonID)
	}
	t := &models.Team{
		Code: req.Code, Name: req.Name, SeasonID: req.SeasonID,
		LogoURL: req.LogoURL, CaptainID: req.CaptainID, Contact: req.Contact,
		Status: models.TeamStatusPending, CreatedBy: creatorID,
	}
	if err := s.Teams.Create(ctx, t); err != nil {
		if t.ID == 0 {
			return nil, err
		}
		return t, nil
	}
	return t, nil
}

// Get a team by id.
func (s *TeamService) Get(ctx context.Context, id int64) (*models.Team, error) {
	t, err := s.Teams.GetByID(ctx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("team", id)
	}
	return t, nil
}

// List teams with filters.
func (s *TeamService) List(ctx context.Context, seasonID int64, status, keyword string, p models.Pagination) ([]models.Team, int64, error) {
	return s.Teams.List(ctx, seasonID, status, keyword, p, "id ASC")
}

// Update a team.
func (s *TeamService) Update(ctx context.Context, id int64, req models.UpdateTeamRequest) (*models.Team, error) {
	t, err := s.Teams.GetByID(ctx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("team", id)
	}
	if req.Name != nil {
		t.Name = *req.Name
	}
	if req.LogoURL != nil {
		t.LogoURL = *req.LogoURL
	}
	if req.CaptainID != nil {
		t.CaptainID = req.CaptainID
	}
	if req.Contact != nil {
		t.Contact = *req.Contact
	}
	if req.Status != nil {
		t.Status = *req.Status
	}
	if err := s.Teams.Update(ctx, t); err != nil {
		return nil, errorsx.Conflict("team name already exists in season")
	}
	return t, nil
}

// ReviewRegistration approves or rejects a team registration.
func (s *TeamService) ReviewRegistration(ctx context.Context, teamID, reviewerID int64, req models.TeamRegistrationReview) (*models.Team, error) {
	t, err := s.Teams.GetByID(ctx, teamID)
	if err != nil {
		return nil, errorsx.NotFoundID("team", teamID)
	}
	if err := oneOf("status", req.Status,
		models.TeamStatusApproved, models.TeamStatusRejected, models.TeamStatusSuspended); err != nil {
		return nil, err
	}
	persistedStatus := req.PersistenceStatus()
	if err := s.Teams.ReviewRegistration(ctx, teamID, reviewerID, persistedStatus, req.Note); err != nil {
		return nil, errorsx.Internal("review failed")
	}
	t.Status = req.Status
	return t, nil
}

// ListRegistrations returns pending/all registrations.
func (s *TeamService) ListRegistrations(ctx context.Context, seasonID int64, status string, p models.Pagination) ([]models.TeamRegistration, int64, error) {
	return s.Teams.ListRegistrations(ctx, seasonID, status, p)
}

// Players ----------------------------------------------------------------

func (s *TeamService) CreatePlayer(ctx context.Context, req models.CreatePlayerRequest) (*models.Player, error) {
	if err := requireFields(map[string]string{"name": req.Name}); err != nil {
		return nil, err
	}
	if req.TeamID == 0 {
		return nil, errorsx.BadRequest("team_id is required")
	}
	t, err := s.Teams.GetByID(ctx, req.TeamID)
	if err != nil {
		return nil, errorsx.NotFoundID("team", req.TeamID)
	}
	pl := &models.Player{
		TeamID: req.TeamID, SeasonID: t.SeasonID, Name: req.Name, Number: req.Number,
		Position: req.Position, BirthDate: parseDate(req.BirthDate),
		HeightCM: req.HeightCM, WeightKG: req.WeightKG,
		Status: models.PlayerStatusActive, Eligibility: models.EligibilityPending,
	}
	if err := s.Teams.CreatePlayer(ctx, pl); err != nil {
		if errorsx.IsConflict(err) {
			return nil, errorsx.Conflict("player number already taken in team")
		}
		return nil, errorsx.Internal("player registration failed")
	}
	return pl, nil
}

func (s *TeamService) GetPlayer(ctx context.Context, id int64) (*models.Player, error) {
	pl, err := s.Teams.GetPlayerByID(ctx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("player", id)
	}
	return pl, nil
}

func (s *TeamService) ListPlayers(ctx context.Context, teamID, seasonID int64, status string, p models.Pagination) ([]models.Player, int64, error) {
	return s.Teams.ListPlayers(ctx, teamID, seasonID, status, p, "number ASC")
}

func (s *TeamService) UpdatePlayer(ctx context.Context, id int64, req models.UpdatePlayerRequest) (*models.Player, error) {
	pl, err := s.Teams.GetPlayerByID(ctx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("player", id)
	}
	if req.Name != nil {
		pl.Name = *req.Name
	}
	if req.Number != nil {
		pl.Number = *req.Number
	}
	if req.Position != nil {
		pl.Position = *req.Position
	}
	if req.BirthDate != nil {
		pl.BirthDate = parseDate(*req.BirthDate)
	}
	if req.HeightCM != nil {
		pl.HeightCM = req.HeightCM
	}
	if req.WeightKG != nil {
		pl.WeightKG = req.WeightKG
	}
	if req.Status != nil {
		pl.Status = *req.Status
	}
	if req.Eligibility != nil {
		pl.Eligibility = *req.Eligibility
	}
	if err := s.Teams.UpdatePlayer(ctx, pl); err != nil {
		return nil, errorsx.Conflict("player number already taken in team")
	}
	return pl, nil
}

// Transfers --------------------------------------------------------------

func (s *TeamService) CreateTransfer(ctx context.Context, req models.CreateTransferRequest, requesterID int64) (*models.Transfer, error) {
	if req.PlayerID == 0 || req.ToTeamID == 0 {
		return nil, errorsx.BadRequest("player_id and to_team_id are required")
	}
	pl, err := s.Teams.GetPlayerByID(ctx, req.PlayerID)
	if err != nil {
		return nil, errorsx.NotFoundID("player", req.PlayerID)
	}
	if pl.TeamID == req.ToTeamID {
		return nil, errorsx.BadRequest("player already in target team")
	}
	t := &models.Transfer{
		PlayerID: req.PlayerID, FromTeamID: pl.TeamID, ToTeamID: req.ToTeamID,
		SeasonID: pl.SeasonID, Reason: req.Reason, Status: models.TransferStatusPending,
		EffectiveAt: parseDate(req.EffectiveAt),
	}
	if err := s.Teams.CreateTransfer(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TeamService) ListTransfers(ctx context.Context, status string, p models.Pagination) ([]models.Transfer, int64, error) {
	return s.Teams.ListTransfers(ctx, status, p)
}

func (s *TeamService) ReviewTransfer(ctx context.Context, id, reviewerID int64, req models.TransferReview) (*models.Transfer, error) {
	t, err := s.Teams.GetTransfer(ctx, id)
	if err != nil {
		return nil, errorsx.NotFoundID("transfer", id)
	}
	if err := oneOf("status", req.Status, models.TransferStatusApproved, models.TransferStatusRejected); err != nil {
		return nil, err
	}
	// Atomically claim the terminal state: only a transfer still "pending" can be
	// decided, so concurrent reviewers race on the guarded UPDATE instead of all
	// succeeding. The loser gets a conflict and must NOT touch the player roster.
	persistedStatus, err := s.Teams.ReviewTransfer(ctx, id, reviewerID, req.Status, req.Note)
	if err != nil {
		if errorsx.IsConflict(err) {
			return nil, err
		}
		return nil, errorsx.Internal("review failed")
	}
	if persistedStatus == models.TransferStatusApproved {
		// This caller won the claim, so the player moves to the destination team.
		if err := s.Teams.SetPlayerTeam(ctx, t.PlayerID, t.ToTeamID); err != nil {
			return nil, errorsx.Internal("player team update failed")
		}
	}
	t.Status = persistedStatus
	return t, nil
}
