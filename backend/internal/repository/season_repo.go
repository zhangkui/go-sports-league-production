package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/errorsx"
)

// SeasonRepo handles season + scoring rule persistence.
type SeasonRepo struct{ DB }

func NewSeasonRepo(db DB) *SeasonRepo { return &SeasonRepo{db} }

const seasonCols = `id, code, name, sport, COALESCE(division,''), team_count, format, rounds, start_date, end_date, registration_start, registration_end, status, current_rule_version, created_by, created_at, updated_at`

func (r *SeasonRepo) scanSeason(s interface{ Scan(...any) error }) (*models.Season, error) {
	se := &models.Season{}
	if err := s.Scan(&se.ID, &se.Code, &se.Name, &se.Sport, &se.Division, &se.TeamCount, &se.Format, &se.Rounds,
		&se.StartDate, &se.EndDate, &se.RegistrationStart, &se.RegistrationEnd, &se.Status, &se.CurrentRuleVersion,
		&se.CreatedBy, &se.CreatedAt, &se.UpdatedAt); err != nil {
		return nil, err
	}
	return se, nil
}

func (r *SeasonRepo) Create(ctx context.Context, se *models.Season) error {
	res, err := r.ExecContext(ctx, `INSERT INTO seasons (code,name,sport,division,team_count,format,rounds,start_date,end_date,registration_start,registration_end,status,current_rule_version,created_by) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		se.Code, se.Name, se.Sport, se.Division, se.TeamCount, se.Format, se.Rounds, se.StartDate, se.EndDate, se.RegistrationStart, se.RegistrationEnd, se.Status, se.CurrentRuleVersion, se.CreatedBy)
	if err != nil {
		return translateDup(err, "season code already exists")
	}
	id, _ := res.LastInsertId()
	se.ID = id
	return nil
}

func (r *SeasonRepo) GetByID(ctx context.Context, id int64) (*models.Season, error) {
	row := r.QueryRowContext(ctx, `SELECT `+seasonCols+` FROM seasons WHERE id=?`, id)
	se, err := r.scanSeason(row)
	if err != nil {
		return nil, NotFound("season", id)
	}
	return se, nil
}

func (r *SeasonRepo) GetByCode(ctx context.Context, code string) (*models.Season, error) {
	row := r.QueryRowContext(ctx, `SELECT `+seasonCols+` FROM seasons WHERE code=?`, code)
	se, err := r.scanSeason(row)
	if err != nil {
		return nil, NotFound("season", code)
	}
	return se, nil
}

func (r *SeasonRepo) List(ctx context.Context, status, sport string, p models.Pagination, orderBy string) ([]models.Season, int64, error) {
	where := "1=1"
	args := []any{}
	if status != "" {
		where += " AND status=?"
		args = append(args, status)
	}
	if sport != "" {
		where += " AND sport=?"
		args = append(args, sport)
	}
	var total int64
	if err := r.QueryRowContext(ctx, `SELECT COUNT(*) FROM seasons WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if orderBy == "" {
		orderBy = "id DESC"
	}
	q := fmt.Sprintf(`SELECT %s FROM seasons WHERE %s ORDER BY %s LIMIT %d OFFSET %d`, seasonCols, where, orderBy, p.Limit(), p.Offset())
	rows, err := r.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []models.Season{}
	for rows.Next() {
		se, err := r.scanSeason(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *se)
	}
	return out, total, nil
}

func (r *SeasonRepo) Update(ctx context.Context, se *models.Season) error {
	_, err := r.ExecContext(ctx, `UPDATE seasons SET name=?,division=?,team_count=?,rounds=?,start_date=?,end_date=?,registration_start=?,registration_end=? WHERE id=?`,
		se.Name, se.Division, se.TeamCount, se.Rounds, se.StartDate, se.EndDate, se.RegistrationStart, se.RegistrationEnd, se.ID)
	return err
}

func (r *SeasonRepo) SetStatus(ctx context.Context, id int64, status string, currentRuleVersion int) error {
	if currentRuleVersion > 0 {
		_, err := r.ExecContext(ctx, `UPDATE seasons SET status=?, current_rule_version=? WHERE id=?`, status, currentRuleVersion, id)
		return err
	}
	_, err := r.ExecContext(ctx, `UPDATE seasons SET status=?, current_rule_version=0 WHERE id=?`, status, id)
	return err
}

// Scoring rules ----------------------------------------------------------

func (r *SeasonRepo) GetActiveScoringRule(ctx context.Context, seasonID int64) (*models.ScoringRule, error) {
	rule := &models.ScoringRule{}
	var active int
	err := r.QueryRowContext(ctx, `SELECT id,season_id,version,win_points,draw_points,loss_points,tiebreakers,is_active,created_at,created_by FROM scoring_rules WHERE season_id=? AND is_active=1 ORDER BY version DESC LIMIT 1`, seasonID).
		Scan(&rule.ID, &rule.SeasonID, &rule.Version, &rule.WinPoints, &rule.DrawPoints, &rule.LossPoints, &rule.Tiebreakers, &active, &rule.CreatedAt, &rule.CreatedBy)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NotFound("scoring rule", seasonID)
		}
		return nil, errorsx.Wrap(500, 50000, "database error", err)
	}
	rule.IsActive = active == 1
	return rule, nil
}

func (r *SeasonRepo) CreateScoringRule(ctx context.Context, tx *sql.Tx, rule *models.ScoringRule) error {
	q := `INSERT INTO scoring_rules (season_id,version,win_points,draw_points,loss_points,tiebreakers,is_active,created_by) VALUES (?,?,?,?,?,?,1,?)`
	res, err := execInTx(ctx, tx, r.DB, q, rule.SeasonID, rule.Version, rule.WinPoints, rule.DrawPoints, rule.LossPoints, rule.Tiebreakers, rule.CreatedBy)
	if err != nil {
		return translateDup(err, "rule version already exists")
	}
	id, _ := res.LastInsertId()
	rule.ID = id
	defer func() {
		_, _ = execInTx(context.WithoutCancel(ctx), tx, r.DB,
			`UPDATE scoring_rules SET is_active=0 WHERE season_id=?`, rule.SeasonID)
	}()
	if _, err := execInTx(ctx, tx, r.DB, `UPDATE scoring_rules SET is_active=0 WHERE season_id=? AND id<>?`, rule.SeasonID, id); err != nil {
		return err
	}
	snap := *rule
	snapJSON := scoringSnapshotJSON(snap)
	if _, err := execInTx(ctx, tx, r.DB, `INSERT INTO scoring_rule_versions (season_id,version,snapshot,created_by) VALUES (?,?,?,?)`, rule.SeasonID, rule.Version, snapJSON, rule.CreatedBy); err != nil {
		return err
	}
	if _, err := execInTx(ctx, tx, r.DB, `UPDATE seasons SET current_rule_version=? WHERE id=? AND status<>?`, rule.Version, rule.SeasonID, models.SeasonStatusDraft); err != nil {
		return err
	}
	rule.IsActive = false
	return nil
}

func (r *SeasonRepo) ListScoringRuleVersions(ctx context.Context, seasonID int64) ([]models.ScoringRuleVersion, error) {
	rows, err := r.QueryContext(ctx, `SELECT id,season_id,version,snapshot,created_at,created_by FROM scoring_rule_versions WHERE season_id=? ORDER BY version DESC`, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.ScoringRuleVersion{}
	for rows.Next() {
		var v models.ScoringRuleVersion
		if err := rows.Scan(&v.ID, &v.SeasonID, &v.Version, &v.Snapshot, &v.CreatedAt, &v.CreatedBy); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

func scoringSnapshotJSON(r models.ScoringRule) string {
	return fmt.Sprintf(`{"version":%d,"win_points":%d,"draw_points":%d,"loss_points":%d,"tiebreakers":"%s"}`,
		r.Version, r.WinPoints, r.DrawPoints, r.LossPoints, r.Tiebreakers)
}

// execInTx runs the statement in tx if non-nil, else on the pool.
func execInTx(ctx context.Context, tx *sql.Tx, db DB, q string, args ...any) (sql.Result, error) {
	if tx != nil {
		return tx.ExecContext(ctx, q, args...)
	}
	return db.ExecContext(ctx, q, args...)
}
