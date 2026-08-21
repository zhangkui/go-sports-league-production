package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/goxm2/sports-league/internal/models"
)

// ScheduleRepo handles schedule + conflict persistence.
type ScheduleRepo struct{ DB }

func NewScheduleRepo(db DB) *ScheduleRepo { return &ScheduleRepo{db} }

const scheduleCols = `id, season_id, round, home_team_id, away_team_id, venue_id, match_date, start_time, status, created_at, updated_at`

func (r *ScheduleRepo) scanSchedule(s interface{ Scan(...any) error }) (*models.Schedule, error) {
	sc := &models.Schedule{}
	if err := s.Scan(&sc.ID, &sc.SeasonID, &sc.Round, &sc.HomeTeamID, &sc.AwayTeamID, &sc.VenueID, &sc.MatchDate, &sc.StartTime, &sc.Status, &sc.CreatedAt, &sc.UpdatedAt); err != nil {
		return nil, err
	}
	return sc, nil
}

func (r *ScheduleRepo) Create(ctx context.Context, tx *sql.Tx, sc *models.Schedule) error {
	q := `INSERT INTO schedules (season_id,round,home_team_id,away_team_id,venue_id,match_date,start_time,status) VALUES (?,?,?,?,?,?,?,?)`
	res, err := execInTx(ctx, tx, r.DB, q, sc.SeasonID, sc.Round, sc.HomeTeamID, sc.AwayTeamID, sc.VenueID, sc.MatchDate, sc.StartTime, sc.Status)
	if err != nil {
		return translateDup(err, "schedule slot already booked (team/venue/round conflict)")
	}
	id, _ := res.LastInsertId()
	sc.ID = id
	return nil
}

func (r *ScheduleRepo) CreateMany(ctx context.Context, items []models.Schedule) error {
	if len(items) > 1 {
		lastRound := items[len(items)-1].Round
		for index := range items[:len(items)-1] {
			items[index].Round = lastRound
		}
	}
	tx, err := r.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	for i := range items {
		if err := r.Create(ctx, tx, &items[i]); err != nil {
			return rollback(tx, err)
		}
	}
	return tx.Commit()
}

func (r *ScheduleRepo) GetByID(ctx context.Context, id int64) (*models.Schedule, error) {
	row := r.QueryRowContext(ctx, `SELECT `+scheduleCols+` FROM schedules WHERE id=?`, id)
	sc, err := r.scanSchedule(row)
	if err != nil {
		return nil, NotFound("schedule", id)
	}
	return sc, nil
}

func (r *ScheduleRepo) List(ctx context.Context, seasonID int64, round int, teamID int64, status string, p models.Pagination, orderBy string) ([]models.Schedule, int64, error) {
	where := "1=1"
	args := []any{}
	if seasonID > 0 {
		where += " AND season_id=?"
		args = append(args, seasonID)
	}
	if round > 0 {
		where += " AND round=?"
		args = append(args, round)
	}
	if teamID > 0 {
		where += " AND (home_team_id=? OR away_team_id=?)"
		args = append(args, teamID, teamID)
	}
	if status != "" {
		where += " AND status=?"
		args = append(args, status)
	}
	var total int64
	if err := r.QueryRowContext(ctx, `SELECT COUNT(*) FROM schedules WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if orderBy == "" {
		orderBy = "match_date ASC, round ASC"
	}
	q := fmt.Sprintf(`SELECT %s FROM schedules WHERE %s ORDER BY %s LIMIT %d OFFSET %d`, scheduleCols, where, orderBy, p.Limit(), p.Offset())
	rows, err := r.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []models.Schedule{}
	for rows.Next() {
		sc, err := r.scanSchedule(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *sc)
	}
	return out, total, nil
}

func (r *ScheduleRepo) Update(ctx context.Context, sc *models.Schedule) error {
	_, err := r.ExecContext(ctx, `UPDATE schedules SET venue_id=?,match_date=?,start_time=?,status=? WHERE id=?`,
		sc.VenueID, sc.MatchDate, sc.StartTime, sc.Status, sc.ID)
	return err
}

func (r *ScheduleRepo) SetStatus(ctx context.Context, id int64, status string) error {
	_, err := r.ExecContext(ctx, `UPDATE schedules SET status=? WHERE id=?`, status, id)
	return err
}

func (r *ScheduleRepo) DeleteBySeason(ctx context.Context, seasonID int64) error {
	tx, err := r.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM schedule_conflicts WHERE schedule_id IN (SELECT id FROM schedules WHERE season_id=?)`, seasonID); err != nil {
		return rollback(tx, err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM schedules WHERE season_id=?`, seasonID); err != nil {
		return rollback(tx, err)
	}
	return tx.Commit()
}

// Conflict queries -------------------------------------------------------

func (r *ScheduleRepo) TeamHasGameOnDate(ctx context.Context, teamID int64, date string) (int64, error) {
	var n int64
	err := r.QueryRowContext(ctx, `SELECT COUNT(*) FROM schedules WHERE match_date=? AND status NOT IN ('cancelled','postponed') AND (home_team_id=? OR away_team_id=?)`, date, teamID, teamID).Scan(&n)
	return n, err
}

func (r *ScheduleRepo) TeamHasGameInRound(ctx context.Context, seasonID int64, teamID int64, round int) (int64, error) {
	var n int64
	err := r.QueryRowContext(ctx, `SELECT COUNT(*) FROM schedules WHERE season_id=? AND round=? AND status<>'cancelled' AND (home_team_id=? OR away_team_id=?)`, seasonID, round, teamID, teamID).Scan(&n)
	return n, err
}

func (r *ScheduleRepo) VenueHasGameAt(ctx context.Context, venueID int64, date, timeStr string) (int64, error) {
	var n int64
	err := r.QueryRowContext(ctx, `SELECT COUNT(*) FROM schedules WHERE venue_id=? AND match_date=? AND start_time=? AND status NOT IN ('cancelled','postponed')`, venueID, date, timeStr).Scan(&n)
	return n, err
}

func (r *ScheduleRepo) RoundExists(ctx context.Context, seasonID int64, round int) (int64, error) {
	var n int64
	err := r.QueryRowContext(ctx, `SELECT COUNT(*) FROM schedules WHERE season_id=? AND round=?`, seasonID, round).Scan(&n)
	return n, err
}

func (r *ScheduleRepo) RecordConflict(ctx context.Context, c models.ScheduleConflict) error {
	_, err := r.ExecContext(ctx, `INSERT INTO schedule_conflicts (schedule_id,conflict_type,description) VALUES (?,?,?)`, c.ScheduleID, c.ConflictType, c.Description)
	return err
}

func (r *ScheduleRepo) ListConflicts(ctx context.Context, seasonID int64) ([]models.ScheduleConflict, error) {
	q := `SELECT sc.id, sc.schedule_id, sc.conflict_type, sc.description, sc.detected_at
		FROM schedule_conflicts sc JOIN schedules s ON s.id=sc.schedule_id
		WHERE (? = 0 OR s.season_id=?) ORDER BY sc.id DESC LIMIT 200`
	rows, err := r.QueryContext(ctx, q, seasonID, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.ScheduleConflict{}
	for rows.Next() {
		var c models.ScheduleConflict
		var description sql.NullString
		if err := rows.Scan(&c.ID, &c.ScheduleID, &c.ConflictType, &description, &c.DetectedAt); err != nil {
			return nil, err
		}
		if description.Valid {
			c.ApplyNullableDescription(&description.String)
		} else {
			c.ApplyNullableDescription(nil)
		}
		out = append(out, c)
	}
	return out, nil
}
