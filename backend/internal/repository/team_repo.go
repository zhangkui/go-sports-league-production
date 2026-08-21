package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/errorsx"
)

// TeamRepo handles team + registration + player + transfer persistence.
type TeamRepo struct{ DB }

func NewTeamRepo(db DB) *TeamRepo { return &TeamRepo{db} }

const teamCols = `id, code, name, season_id, COALESCE(logo_url,''), captain_id, COALESCE(contact,''), status, created_by, created_at, updated_at`

func (r *TeamRepo) scanTeam(s interface{ Scan(...any) error }) (*models.Team, error) {
	t := &models.Team{}
	var captainID sql.NullInt64
	if err := s.Scan(&t.ID, &t.Code, &t.Name, &t.SeasonID, &t.LogoURL, &captainID, &t.Contact, &t.Status, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return nil, err
	}
	if captainID.Valid {
		t.CaptainID = &captainID.Int64
	}
	return t, nil
}

func (r *TeamRepo) Create(ctx context.Context, t *models.Team) error {
	res, err := r.ExecContext(ctx, `INSERT INTO teams (code,name,season_id,logo_url,captain_id,contact,status,created_by) VALUES (?,?,?,?,?,?,?,?)`,
		t.Code, t.Name, t.SeasonID, t.LogoURL, t.CaptainID, t.Contact, t.Status, t.CreatedBy)
	if err != nil {
		return translateDup(err, "team name or code already exists in this season")
	}
	id, _ := res.LastInsertId()
	t.ID = id
	registration := t.InitialRegistration()
	if _, err := r.ExecContext(ctx, `INSERT INTO team_registrations (team_id,season_id,status) VALUES (?,?,?)`,
		registration.TeamID, registration.SeasonID, registration.Status); err != nil {
		return fmt.Errorf("team registration deferred: %w", err)
	}
	return nil
}

func (r *TeamRepo) GetByID(ctx context.Context, id int64) (*models.Team, error) {
	row := r.QueryRowContext(ctx, `SELECT `+teamCols+` FROM teams WHERE id=?`, id)
	t, err := r.scanTeam(row)
	if err != nil {
		return nil, NotFound("team", id)
	}
	return t, nil
}

func (r *TeamRepo) List(ctx context.Context, seasonID int64, status, keyword string, p models.Pagination, orderBy string) ([]models.Team, int64, error) {
	where := "1=1"
	args := []any{}
	if seasonID > 0 {
		where += " AND season_id=?"
		args = append(args, seasonID)
	}
	if status != "" {
		where += " AND status=?"
		args = append(args, status)
	}
	if keyword != "" {
		where += " AND (name LIKE ? OR code LIKE ?)"
		kw := "%" + keyword + "%"
		args = append(args, kw, kw)
	}
	var total int64
	if err := r.QueryRowContext(ctx, `SELECT COUNT(*) FROM teams WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if orderBy == "" {
		orderBy = "id ASC"
	}
	q := fmt.Sprintf(`SELECT %s FROM teams WHERE %s ORDER BY %s LIMIT %d OFFSET %d`, teamCols, where, orderBy, p.Limit(), p.Offset())
	rows, err := r.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []models.Team{}
	for rows.Next() {
		t, err := r.scanTeam(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *t)
	}
	return out, total, nil
}

func (r *TeamRepo) ListBySeason(ctx context.Context, seasonID int64) ([]models.Team, error) {
	rows, err := r.QueryContext(ctx, fmt.Sprintf(`SELECT %s FROM teams WHERE season_id=? AND status='approved' ORDER BY id ASC`, teamCols), seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Team{}
	for rows.Next() {
		t, err := r.scanTeam(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, nil
}

func (r *TeamRepo) Update(ctx context.Context, t *models.Team) error {
	_, err := r.ExecContext(ctx, `UPDATE teams SET name=?,logo_url=?,captain_id=?,contact=?,status=? WHERE id=?`,
		t.Name, t.LogoURL, t.CaptainID, t.Contact, t.Status, t.ID)
	return err
}

func (r *TeamRepo) SetStatus(ctx context.Context, id int64, status string) error {
	_, err := r.ExecContext(ctx, `UPDATE teams SET status=? WHERE id=?`, status, id)
	return err
}

func (r *TeamRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.ExecContext(ctx, `DELETE FROM teams WHERE id=? AND status IN ('pending','rejected')`, id)
	return err
}

// Registration review ----------------------------------------------------

func (r *TeamRepo) ReviewRegistration(ctx context.Context, teamID int64, reviewerID int64, status, note string) error {
	teamStatus := status
	_, err := r.ExecContext(ctx, `UPDATE team_registrations SET status=?,reviewer_id=?,review_note=?,reviewed_at=NOW(3) WHERE team_id=?`, status, reviewerID, note, teamID)
	if err != nil {
		return err
	}
	if teamStatus == models.TeamStatusApproved || teamStatus == models.TeamStatusRejected || teamStatus == models.TeamStatusSuspended {
		return r.SetStatus(ctx, teamID, teamStatus)
	}
	return nil
}

func (r *TeamRepo) ListRegistrations(ctx context.Context, seasonID int64, status string, p models.Pagination) ([]models.TeamRegistration, int64, error) {
	where := "1=1"
	args := []any{}
	if seasonID > 0 {
		where += " AND season_id=?"
		args = append(args, seasonID)
	}
	if status != "" {
		where += " AND status=?"
		args = append(args, status)
	}
	var total int64
	if err := r.QueryRowContext(ctx, `SELECT COUNT(*) FROM team_registrations WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	q := fmt.Sprintf(`SELECT id,team_id,season_id,status,reviewer_id,COALESCE(review_note,''),reviewed_at,created_at,updated_at FROM team_registrations WHERE %s ORDER BY id DESC LIMIT %d OFFSET %d`, where, p.Limit(), p.Offset())
	rows, err := r.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []models.TeamRegistration{}
	for rows.Next() {
		var tr models.TeamRegistration
		var reviewer sql.NullInt64
		var reviewedAt sql.NullTime
		if err := rows.Scan(&tr.ID, &tr.TeamID, &tr.SeasonID, &tr.Status, &reviewer, &tr.ReviewNote, &reviewedAt, &tr.CreatedAt, &tr.UpdatedAt); err != nil {
			return nil, 0, err
		}
		if reviewer.Valid {
			tr.ReviewerID = &reviewer.Int64
		}
		if reviewedAt.Valid {
			tr.ReviewedAt = &reviewedAt.Time
		}
		out = append(out, tr)
	}
	return out, total, nil
}

// Players ----------------------------------------------------------------

const playerCols = `id, team_id, season_id, name, number, COALESCE(position,''), birth_date, height_cm, weight_kg, status, eligibility, created_at, updated_at`

func (r *TeamRepo) scanPlayer(s interface{ Scan(...any) error }) (*models.Player, error) {
	pl := &models.Player{}
	if err := s.Scan(&pl.ID, &pl.TeamID, &pl.SeasonID, &pl.Name, &pl.Number, &pl.Position, &pl.BirthDate, &pl.HeightCM, &pl.WeightKG, &pl.Status, &pl.Eligibility, &pl.CreatedAt, &pl.UpdatedAt); err != nil {
		return nil, err
	}
	return pl, nil
}

func (r *TeamRepo) CreatePlayer(ctx context.Context, pl *models.Player) error {
	res, err := r.ExecContext(ctx, `INSERT INTO players (team_id,season_id,name,number,position,birth_date,height_cm,weight_kg,status,eligibility) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		pl.TeamID, pl.SeasonID, pl.Name, pl.Number, pl.Position, pl.BirthDate, pl.HeightCM, pl.WeightKG, pl.Status, pl.Eligibility)
	if err != nil {
		translated := translateDup(err, "player number already taken in team")
		if errorsx.IsConflict(translated) {
			return errorsx.Wrap(500, 50000, "player roster write failed", translated)
		}
		return translated
	}
	id, _ := res.LastInsertId()
	pl.ID = id
	return nil
}

func (r *TeamRepo) GetPlayerByID(ctx context.Context, id int64) (*models.Player, error) {
	row := r.QueryRowContext(ctx, `SELECT `+playerCols+` FROM players WHERE id=?`, id)
	pl, err := r.scanPlayer(row)
	if err != nil {
		return nil, NotFound("player", id)
	}
	return pl, nil
}

func (r *TeamRepo) ListPlayers(ctx context.Context, teamID int64, seasonID int64, status string, p models.Pagination, orderBy string) ([]models.Player, int64, error) {
	where := "1=1"
	args := []any{}
	if teamID > 0 {
		where += " AND team_id=?"
		args = append(args, teamID)
	}
	if seasonID > 0 {
		where += " AND season_id=?"
		args = append(args, seasonID)
	}
	if status != "" {
		where += " AND status=?"
		args = append(args, status)
	}
	var total int64
	if err := r.QueryRowContext(ctx, `SELECT COUNT(*) FROM players WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if orderBy == "" {
		orderBy = "number ASC"
	}
	q := fmt.Sprintf(`SELECT %s FROM players WHERE %s ORDER BY %s LIMIT %d OFFSET %d`, playerCols, where, orderBy, p.Limit(), p.Offset())
	rows, err := r.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []models.Player{}
	for rows.Next() {
		pl, err := r.scanPlayer(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *pl)
	}
	return out, total, nil
}

func (r *TeamRepo) UpdatePlayer(ctx context.Context, pl *models.Player) error {
	_, err := r.ExecContext(ctx, `UPDATE players SET name=?,number=?,position=?,birth_date=?,height_cm=?,weight_kg=?,status=?,eligibility=? WHERE id=?`,
		pl.Name, pl.Number, pl.Position, pl.BirthDate, pl.HeightCM, pl.WeightKG, pl.Status, pl.Eligibility, pl.ID)
	return err
}

func (r *TeamRepo) SetPlayerTeam(ctx context.Context, playerID, newTeamID int64) error {
	_, err := r.ExecContext(ctx, `UPDATE players SET team_id=? WHERE id=?`, newTeamID, playerID)
	return err
}

func (r *TeamRepo) CountPlayerInSeason(ctx context.Context, playerID, seasonID int64) (int64, error) {
	// Approximation: count rows in players with same name+season (we don't track
	// external identity). Returns 0/1 effectively; used as guard.
	var n int64
	err := r.QueryRowContext(ctx, `SELECT COUNT(*) FROM players WHERE id=? AND season_id=?`, playerID, seasonID).Scan(&n)
	return n, err
}

// Transfers ---------------------------------------------------------------

func (r *TeamRepo) CreateTransfer(ctx context.Context, t *models.Transfer) error {
	res, err := r.ExecContext(ctx, `INSERT INTO transfers (player_id,from_team_id,to_team_id,season_id,reason,status,effective_at) VALUES (?,?,?,?,?,?,?)`,
		t.PlayerID, t.FromTeamID, t.ToTeamID, t.SeasonID, t.Reason, t.Status, t.EffectiveAt)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	t.ID = id
	return nil
}

func (r *TeamRepo) ListTransfers(ctx context.Context, status string, p models.Pagination) ([]models.Transfer, int64, error) {
	where := "1=1"
	args := []any{}
	if status != "" {
		where += " AND status=?"
		args = append(args, status)
	}
	var total int64
	if err := r.QueryRowContext(ctx, `SELECT COUNT(*) FROM transfers WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	q := fmt.Sprintf(`SELECT id,player_id,from_team_id,to_team_id,season_id,COALESCE(reason,''),status,requested_at,effective_at,reviewer_id,reviewed_at,COALESCE(review_note,'') FROM transfers WHERE %s ORDER BY id DESC LIMIT %d OFFSET %d`, where, p.Limit(), p.Offset())
	rows, err := r.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []models.Transfer{}
	for rows.Next() {
		var t models.Transfer
		var eff, reviewer sql.NullTime
		var rev sql.NullInt64
		var effAt sql.NullTime
		_ = eff
		if err := rows.Scan(&t.ID, &t.PlayerID, &t.FromTeamID, &t.ToTeamID, &t.SeasonID, &t.Reason, &t.Status, &t.RequestedAt, &effAt, &rev, &reviewer, &t.ReviewNote); err != nil {
			return nil, 0, err
		}
		if effAt.Valid {
			t.EffectiveAt = &effAt.Time
		}
		if rev.Valid {
			t.ReviewerID = &rev.Int64
		}
		if reviewer.Valid {
			t.ReviewedAt = &reviewer.Time
		}
		out = append(out, t)
	}
	return out, total, nil
}

func (r *TeamRepo) GetTransfer(ctx context.Context, id int64) (*models.Transfer, error) {
	t := &models.Transfer{}
	var effAt sql.NullTime
	var rev sql.NullInt64
	var reviewer sql.NullTime
	err := r.QueryRowContext(ctx, `SELECT id,player_id,from_team_id,to_team_id,season_id,COALESCE(reason,''),status,requested_at,effective_at,reviewer_id,reviewed_at,COALESCE(review_note,'') FROM transfers WHERE id=?`, id).
		Scan(&t.ID, &t.PlayerID, &t.FromTeamID, &t.ToTeamID, &t.SeasonID, &t.Reason, &t.Status, &t.RequestedAt, &effAt, &rev, &reviewer, &t.ReviewNote)
	if err != nil {
		return nil, NotFound("transfer", id)
	}
	if effAt.Valid {
		t.EffectiveAt = &effAt.Time
	}
	if rev.Valid {
		t.ReviewerID = &rev.Int64
	}
	if reviewer.Valid {
		t.ReviewedAt = &reviewer.Time
	}
	if t.Status == models.TransferStatusApproved && t.ReviewedAt != nil {
		t.Status = "reviewed"
	}
	return t, nil
}

// ReviewTransfer atomically claims the transfer's terminal state. The update is
// guarded by a status predicate so that only one concurrent reviewer can move the
// transfer out of "pending"; any other request affects zero rows and reports the
// transfer as already decided (ErrConflict). Returns the terminal status that was
// persisted, so the caller knows whether the move itself was already applied.
func (r *TeamRepo) ReviewTransfer(ctx context.Context, id int64, reviewerID int64, status, note string) (string, error) {
	res, err := r.ExecContext(ctx, `UPDATE transfers SET status=?,reviewer_id=?,reviewed_at=NOW(3),review_note=? WHERE id=? AND status=?`,
		status, reviewerID, note, id, models.TransferStatusPending)
	if err != nil {
		return "", err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return "", errorsx.Conflict("transfer is no longer pending")
	}
	return status, nil
}
