package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/goxm2/sports-league/internal/models"
)

// DisciplineRepo handles discipline + suspension + appeal persistence.
type DisciplineRepo struct{ DB }

func NewDisciplineRepo(db DB) *DisciplineRepo { return &DisciplineRepo{db} }

const disciplineCols = `id, player_id, team_id, season_id, match_id, punishment, reason, severity, suspend_games, fine_amount, status, issued_by, created_at, updated_at`

func (r *DisciplineRepo) scan(s interface{ Scan(...any) error }) (*models.Discipline, error) {
	d := &models.Discipline{}
	var matchID sql.NullInt64
	if err := s.Scan(&d.ID, &d.PlayerID, &d.TeamID, &d.SeasonID, &matchID, &d.Punishment, &d.Reason, &d.Severity, &d.SuspendGames, &d.FineAmount, &d.Status, &d.IssuedBy, &d.CreatedAt, &d.UpdatedAt); err != nil {
		return nil, err
	}
	if matchID.Valid {
		d.MatchID = &matchID.Int64
	}
	return d, nil
}

func (r *DisciplineRepo) Create(ctx context.Context, tx *sql.Tx, d *models.Discipline) error {
	q := `INSERT INTO disciplines (player_id,team_id,season_id,match_id,punishment,reason,severity,suspend_games,fine_amount,status,issued_by) VALUES (?,?,?,?,?,?,?,?,?,?,?)`
	res, err := execInTx(ctx, tx, r.DB, q, d.PlayerID, d.TeamID, d.SeasonID, d.MatchID, d.Punishment, d.Reason, d.Severity, d.SuspendGames, d.FineAmount, d.Status, d.IssuedBy)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	d.ID = id
	return nil
}

func (r *DisciplineRepo) GetByID(ctx context.Context, id int64) (*models.Discipline, error) {
	row := r.QueryRowContext(ctx, `SELECT `+disciplineCols+` FROM disciplines WHERE id=?`, id)
	d, err := r.scan(row)
	if err != nil {
		return nil, NotFound("discipline", id)
	}
	return d, nil
}

func (r *DisciplineRepo) List(ctx context.Context, seasonID int64, playerID int64, status string, p models.Pagination, orderBy string) ([]models.Discipline, int64, error) {
	where := "1=1"
	args := []any{}
	if seasonID > 0 {
		where += " AND season_id=?"
		args = append(args, seasonID)
	}
	if playerID > 0 {
		where += " AND player_id=?"
		args = append(args, playerID)
	}
	if status != "" {
		where += " AND status=?"
		args = append(args, status)
	}
	var total int64
	if err := r.QueryRowContext(ctx, `SELECT COUNT(*) FROM disciplines WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if orderBy == "" {
		orderBy = "id DESC"
	}
	q := fmt.Sprintf(`SELECT %s FROM disciplines WHERE %s ORDER BY %s LIMIT %d OFFSET %d`, disciplineCols, where, orderBy, p.Limit(), p.Offset())
	rows, err := r.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []models.Discipline{}
	for rows.Next() {
		d, err := r.scan(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *d)
	}
	return out, total, nil
}

func (r *DisciplineRepo) SetStatus(ctx context.Context, id int64, status string) error {
	_, err := r.ExecContext(ctx, `UPDATE disciplines SET status=? WHERE id=?`, status, id)
	return err
}

// Suspension -------------------------------------------------------------

func (r *DisciplineRepo) CreateSuspension(ctx context.Context, tx *sql.Tx, s *models.Suspension) error {
	q := `INSERT INTO suspensions (discipline_id,player_id,total_games,served_games,start_date,end_date,status) VALUES (?,?,?,?,?,?,?)`
	res, err := execInTx(ctx, tx, r.DB, q, s.DisciplineID, s.PlayerID, s.TotalGames, s.ServedGames, s.StartDate, s.EndDate, s.Status)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	s.ID = id
	return nil
}

func (r *DisciplineRepo) GetSuspension(ctx context.Context, disciplineID int64) (*models.Suspension, error) {
	s := &models.Suspension{}
	var end sql.NullTime
	err := r.QueryRowContext(ctx, `SELECT id,discipline_id,player_id,total_games,served_games,start_date,end_date,status FROM suspensions WHERE discipline_id=?`, disciplineID).
		Scan(&s.ID, &s.DisciplineID, &s.PlayerID, &s.TotalGames, &s.ServedGames, &s.StartDate, &end, &s.Status)
	if err != nil {
		return nil, err
	}
	if end.Valid {
		s.EndDate = &end.Time
	}
	return s, nil
}

func (r *DisciplineRepo) IncrementServed(ctx context.Context, disciplineID int64) error {
	_, err := r.ExecContext(ctx, `UPDATE suspensions SET served_games=served_games+1, status=CASE WHEN served_games+1>=total_games THEN 'served' ELSE status END WHERE discipline_id=?`, disciplineID)
	return err
}

// Appeals ----------------------------------------------------------------

func (r *DisciplineRepo) CreateAppeal(ctx context.Context, a *models.Appeal) error {
	res, err := r.ExecContext(ctx, `INSERT INTO appeals (discipline_id,appellant_id,reason,status) VALUES (?,?,?,?)`,
		a.DisciplineID, a.AppellantID, a.Reason, a.Status)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	a.ID = id
	return nil
}

func (r *DisciplineRepo) ListAppeals(ctx context.Context, status string, p models.Pagination) ([]models.Appeal, int64, error) {
	where := "1=1"
	args := []any{}
	if status != "" {
		where += " AND status=?"
		args = append(args, status)
	}
	var total int64
	if err := r.QueryRowContext(ctx, `SELECT COUNT(*) FROM appeals WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	q := fmt.Sprintf(`SELECT id,discipline_id,appellant_id,reason,status,reviewer_id,COALESCE(review_opinion,''),reviewed_at,created_at FROM appeals WHERE %s ORDER BY id DESC LIMIT %d OFFSET %d`, where, p.Limit(), p.Offset())
	rows, err := r.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []models.Appeal{}
	for rows.Next() {
		var a models.Appeal
		var reviewer sql.NullInt64
		var reviewed sql.NullTime
		if err := rows.Scan(&a.ID, &a.DisciplineID, &a.AppellantID, &a.Reason, &a.Status, &reviewer, &a.ReviewOpinion, &reviewed, &a.CreatedAt); err != nil {
			return nil, 0, err
		}
		if reviewer.Valid {
			a.ReviewerID = &reviewer.Int64
		}
		if reviewed.Valid {
			a.ReviewedAt = &reviewed.Time
		}
		out = append(out, a)
	}
	return out, total, nil
}

func (r *DisciplineRepo) GetAppeal(ctx context.Context, id int64) (*models.Appeal, error) {
	a := &models.Appeal{}
	var reviewer sql.NullInt64
	var reviewed sql.NullTime
	err := r.QueryRowContext(ctx, `SELECT id,discipline_id,appellant_id,reason,status,reviewer_id,COALESCE(review_opinion,''),reviewed_at,created_at FROM appeals WHERE id=?`, id).
		Scan(&a.ID, &a.DisciplineID, &a.AppellantID, &a.Reason, &a.Status, &reviewer, &a.ReviewOpinion, &reviewed, &a.CreatedAt)
	if err != nil {
		return nil, NotFound("appeal", id)
	}
	if reviewer.Valid {
		a.ReviewerID = &reviewer.Int64
	}
	if reviewed.Valid {
		a.ReviewedAt = &reviewed.Time
	}
	return a, nil
}

func (r *DisciplineRepo) ReviewAppeal(ctx context.Context, id int64, reviewerID int64, status, opinion string) error {
	_, err := r.ExecContext(ctx, `UPDATE appeals SET status=?,reviewer_id=?,review_opinion=?,reviewed_at=NOW(3) WHERE id=?`, status, reviewerID, opinion, id)
	return err
}
