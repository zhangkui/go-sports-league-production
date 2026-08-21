package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/goxm2/sports-league/internal/models"
)

// MatchRepo handles match + event + stats persistence.
type MatchRepo struct{ DB }

func NewMatchRepo(db DB) *MatchRepo { return &MatchRepo{db} }

const matchCols = `id, schedule_id, season_id, home_team_id, away_team_id, venue_id, match_date, start_time,
	home_score, away_score, home_half_score, away_half_score, referee_id, recorder_id, duration_min,
	status, confirm_status, confirmed_by, confirmed_at, created_at, updated_at`

func (r *MatchRepo) scanMatch(s interface{ Scan(...any) error }) (*models.Match, error) {
	m := &models.Match{}
	var homeScore, awayScore, homeHalf, awayHalf sql.NullInt64
	var referee, recorder, confirmedBy sql.NullInt64
	var duration sql.NullInt64
	var confirmedAt sql.NullTime
	if err := s.Scan(&m.ID, &m.ScheduleID, &m.SeasonID, &m.HomeTeamID, &m.AwayTeamID, &m.VenueID, &m.MatchDate, &m.StartTime,
		&homeScore, &awayScore, &homeHalf, &awayHalf, &referee, &recorder, &duration,
		&m.Status, &m.ConfirmStatus, &confirmedBy, &confirmedAt, &m.CreatedAt, &m.UpdatedAt); err != nil {
		return nil, err
	}
	if homeScore.Valid {
		v := int(homeScore.Int64)
		m.HomeScore = &v
	}
	if awayScore.Valid {
		v := int(awayScore.Int64)
		m.AwayScore = &v
	}
	if homeHalf.Valid {
		v := int(homeHalf.Int64)
		m.HomeHalfScore = &v
	}
	if awayHalf.Valid {
		v := int(awayHalf.Int64)
		m.AwayHalfScore = &v
	}
	if referee.Valid {
		m.RefereeID = &referee.Int64
	}
	if recorder.Valid {
		m.RecorderID = &recorder.Int64
	}
	if duration.Valid {
		v := int(duration.Int64)
		m.DurationMin = &v
	}
	if confirmedBy.Valid {
		m.ConfirmedBy = &confirmedBy.Int64
	}
	if confirmedAt.Valid {
		m.ConfirmedAt = &confirmedAt.Time
	}
	return m, nil
}

// EnsureForSchedule lazily creates a match row from its schedule.
//
// Concurrency: the UNIQUE index on matches.schedule_id is the guard. The
// initial GetBySchedule is only a fast path for the already-materialised case;
// it is not atomic. Under a race, two callers both miss it and both attempt
// the INSERT — exactly one wins. The loser observes a duplicate-key error and
// re-reads the row the winner persisted, so every concurrent caller returns the
// same match (same ID) instead of a fabricated object that has no row behind it.
func (r *MatchRepo) EnsureForSchedule(ctx context.Context, sc *models.Schedule) (*models.Match, error) {
	if existing, err := r.GetBySchedule(ctx, sc.ID); err == nil && existing != nil {
		return existing, nil
	}
	m := &models.Match{
		ScheduleID: sc.ID, SeasonID: sc.SeasonID, HomeTeamID: sc.HomeTeamID, AwayTeamID: sc.AwayTeamID,
		VenueID: sc.VenueID, MatchDate: sc.MatchDate, StartTime: sc.StartTime,
		Status: models.MatchStatusScheduled, ConfirmStatus: models.ConfirmStatusPending,
	}
	res, err := r.ExecContext(ctx, `INSERT INTO matches (schedule_id,season_id,home_team_id,away_team_id,venue_id,match_date,start_time,status,confirm_status) VALUES (?,?,?,?,?,?,?,?,?)`,
		m.ScheduleID, m.SeasonID, m.HomeTeamID, m.AwayTeamID, m.VenueID, m.MatchDate, m.StartTime, m.Status, m.ConfirmStatus)
	if err != nil {
		if isDuplicateKey(err) {
			// Lost the INSERT race to another request; return the row it persisted
			// so this caller holds the same persisted ID rather than a fabricated one.
			if existing, gerr := r.GetBySchedule(ctx, sc.ID); gerr == nil && existing != nil {
				return existing, nil
			}
		}
		return nil, translateDup(err, "match already exists for this schedule")
	}
	id, _ := res.LastInsertId()
	m.ID = id
	return m, nil
}

func (r *MatchRepo) GetBySchedule(ctx context.Context, scheduleID int64) (*models.Match, error) {
	row := r.QueryRowContext(ctx, `SELECT `+matchCols+` FROM matches WHERE schedule_id=?`, scheduleID)
	m, err := r.scanMatch(row)
	if err != nil {
		return nil, NotFound("match for schedule", scheduleID)
	}
	return m, nil
}

func (r *MatchRepo) GetByID(ctx context.Context, id int64) (*models.Match, error) {
	row := r.QueryRowContext(ctx, `SELECT `+matchCols+` FROM matches WHERE id=?`, id)
	m, err := r.scanMatch(row)
	if err != nil {
		return nil, NotFound("match", id)
	}
	return m, nil
}

func (r *MatchRepo) List(ctx context.Context, seasonID int64, status string, teamID int64, p models.Pagination, orderBy string) ([]models.Match, int64, error) {
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
	if teamID > 0 {
		where += " AND (home_team_id=? OR away_team_id=?)"
		args = append(args, teamID, teamID)
	}
	var total int64
	if err := r.QueryRowContext(ctx, `SELECT COUNT(*) FROM matches WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if orderBy == "" {
		orderBy = "match_date DESC, id DESC"
	}
	q := fmt.Sprintf(`SELECT %s FROM matches WHERE %s ORDER BY %s LIMIT %d OFFSET %d`, matchCols, where, orderBy, p.Limit(), p.Offset())
	rows, err := r.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []models.Match{}
	for rows.Next() {
		m, err := r.scanMatch(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *m)
	}
	return out, total, nil
}

func (r *MatchRepo) UpdateRecord(ctx context.Context, tx *sql.Tx, m *models.Match) error {
	q := `UPDATE matches SET home_score=?,away_score=?,home_half_score=?,away_half_score=?,referee_id=?,recorder_id=?,duration_min=?,status=?,confirm_status=?,confirmed_by=?,confirmed_at=? WHERE id=?`
	_, err := execInTx(ctx, tx, r.DB, q, m.HomeScore, m.AwayScore, m.HomeHalfScore, m.AwayHalfScore, m.RefereeID, m.RecorderID, m.DurationMin, m.Status, m.ConfirmStatus, m.ConfirmedBy, m.ConfirmedAt, m.ID)
	return err
}

func (r *MatchRepo) SetConfirmStatus(ctx context.Context, tx *sql.Tx, id int64, status string, confirmer int64) error {
	q := `UPDATE matches SET confirm_status=?,confirmed_by=?,confirmed_at=NOW(3) WHERE id=?`
	_, err := execInTx(ctx, tx, r.DB, q, status, confirmer, id)
	return err
}

func (r *MatchRepo) ClaimDispute(ctx context.Context, tx *sql.Tx, id int64) error {
	_, err := execInTx(ctx, tx, r.DB, `UPDATE matches
		SET confirm_status=?, confirmed_by=NULL, confirmed_at=NULL
		WHERE id=? AND confirm_status IN (?,?)`,
		models.ConfirmStatusDisputed, id, models.ConfirmStatusConfirmed, models.ConfirmStatusDisputed)
	return err
}

func (r *MatchRepo) SetStatus(ctx context.Context, id int64, status string) error {
	_, err := r.ExecContext(ctx, `UPDATE matches SET status=? WHERE id=?`, status, id)
	return err
}

// Events -----------------------------------------------------------------

func (r *MatchRepo) CreateEvent(ctx context.Context, e *models.MatchEvent) error {
	res, err := r.ExecContext(ctx, `INSERT INTO match_events (match_id,team_id,player_id,event_type,minute,description,confirm_status,created_by) VALUES (?,?,?,?,?,?,?,?)`,
		e.MatchID, e.TeamID, e.PlayerID, e.EventType, e.Minute, e.Description, e.ConfirmStatus, e.CreatedBy)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	e.ID = id
	return nil
}

func (r *MatchRepo) ListEvents(ctx context.Context, matchID int64) ([]models.MatchEvent, error) {
	rows, err := r.QueryContext(ctx, `SELECT id,match_id,team_id,player_id,event_type,minute,COALESCE(description,''),confirm_status,created_by,created_at FROM match_events WHERE match_id=? ORDER BY minute ASC, id ASC`, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.MatchEvent{}
	for rows.Next() {
		var e models.MatchEvent
		var pid sql.NullInt64
		if err := rows.Scan(&e.ID, &e.MatchID, &e.TeamID, &pid, &e.EventType, &e.Minute, &e.Description, &e.ConfirmStatus, &e.CreatedBy, &e.CreatedAt); err != nil {
			return nil, err
		}
		if pid.Valid {
			e.PlayerID = &pid.Int64
		}
		out = append(out, e)
	}
	return out, nil
}

func (r *MatchRepo) GetEvent(ctx context.Context, id int64) (*models.MatchEvent, error) {
	e := &models.MatchEvent{}
	var pid sql.NullInt64
	err := r.QueryRowContext(ctx, `SELECT id,match_id,team_id,player_id,event_type,minute,COALESCE(description,''),confirm_status,created_by,created_at FROM match_events WHERE id=?`, id).
		Scan(&e.ID, &e.MatchID, &e.TeamID, &pid, &e.EventType, &e.Minute, &e.Description, &e.ConfirmStatus, &e.CreatedBy, &e.CreatedAt)
	if err != nil {
		return nil, NotFound("event", id)
	}
	if pid.Valid {
		e.PlayerID = &pid.Int64
	}
	return e, nil
}

func (r *MatchRepo) UpdateEvent(ctx context.Context, e *models.MatchEvent) error {
	_, err := r.ExecContext(ctx, `UPDATE match_events SET team_id=?,player_id=?,event_type=?,minute=?,description=?,confirm_status=? WHERE id=?`,
		e.TeamID, e.PlayerID, e.EventType, e.Minute, e.Description, e.ConfirmStatus, e.ID)
	return err
}

func (r *MatchRepo) DeleteEvent(ctx context.Context, id int64) error {
	_, err := r.ExecContext(ctx, `DELETE FROM match_events WHERE id=?`, id)
	return err
}

// Player stats -----------------------------------------------------------

var playerStatArgs = make([]any, 7)

func (r *MatchRepo) UpsertPlayerStat(ctx context.Context, st *models.PlayerMatchStat) error {
	playerStatArgs[0] = st.MatchID
	playerStatArgs[1] = st.PlayerID
	playerStatArgs[2] = st.TeamID
	playerStatArgs[3] = st.IsStarter
	if st.PlayedMin != nil {
		playerStatArgs[4] = *st.PlayedMin
	}
	playerStatArgs[5] = st.Position
	if st.Rating != nil {
		playerStatArgs[6] = *st.Rating
	}
	_, err := r.ExecContext(ctx, `INSERT INTO player_match_stats (match_id,player_id,team_id,is_starter,played_min,position,rating) VALUES (?,?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE is_starter=VALUES(is_starter),played_min=VALUES(played_min),position=VALUES(position),rating=VALUES(rating)`,
		playerStatArgs...)
	return err
}

func (r *MatchRepo) ListPlayerStats(ctx context.Context, matchID int64) ([]models.PlayerMatchStat, error) {
	rows, err := r.QueryContext(ctx, `SELECT id,match_id,player_id,team_id,is_starter,played_min,COALESCE(position,''),rating FROM player_match_stats WHERE match_id=? ORDER BY is_starter DESC, player_id`, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.PlayerMatchStat{}
	for rows.Next() {
		var st models.PlayerMatchStat
		var starter int
		var played sql.NullInt64
		var rating sql.NullFloat64
		if err := rows.Scan(&st.ID, &st.MatchID, &st.PlayerID, &st.TeamID, &starter, &played, &st.Position, &rating); err != nil {
			return nil, err
		}
		st.IsStarter = starter == 1
		if played.Valid {
			v := int(played.Int64)
			st.PlayedMin = &v
		}
		if rating.Valid {
			v := rating.Float64
			st.Rating = &v
		}
		out = append(out, st)
	}
	return out, nil
}

func (r *MatchRepo) CountBySeason(ctx context.Context, seasonID int64) (int64, int64, error) {
	var total, completed int64
	if err := r.QueryRowContext(ctx, `SELECT COUNT(*) FROM matches WHERE season_id=?`, seasonID).Scan(&total); err != nil {
		return 0, 0, err
	}
	if err := r.QueryRowContext(ctx, `SELECT COUNT(*) FROM matches WHERE season_id=? AND status='completed'`, seasonID).Scan(&completed); err != nil {
		return 0, 0, err
	}
	return total, completed, nil
}
