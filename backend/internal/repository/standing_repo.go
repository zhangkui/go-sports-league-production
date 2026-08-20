package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/goxm2/sports-league/internal/models"
)

// StandingRepo handles standings + snapshot persistence.
type StandingRepo struct{ DB }

func NewStandingRepo(db DB) *StandingRepo { return &StandingRepo{db} }

// EnsureRow lazily inserts an empty standing row for a team if missing.
func (r *StandingRepo) EnsureRow(ctx context.Context, tx *sql.Tx, seasonID, teamID int64) error {
	q := `INSERT IGNORE INTO standings (season_id,team_id,played,wins,draws,losses,goals_for,goals_against,goal_diff,points,fair_play,rank,prev_rank) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`
	_, err := execInTx(ctx, tx, r.DB, q, seasonID, teamID, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	return err
}

// Apply updates the standing for a team given a result delta.
func (r *StandingRepo) Apply(ctx context.Context, tx *sql.Tx, seasonID, teamID int64, goalsFor, goalsAgainst, win, draw, loss, fairDelta int) error {
	q := `UPDATE standings SET
		played=played+1,
		wins=wins+IF(?>?,1,0),
		draws=draws+IF(?=?,1,0),
		losses=losses+IF(?<?,1,0),
		goals_for=goals_for+?,
		goals_against=goals_against+?,
		goal_diff=goal_diff+(?-?),
		points=points+IF(?>?, ?, IF(?=?, ?, ?)),
		fair_play=fair_play+?
		WHERE season_id=? AND team_id=?`
	_, err := execInTx(ctx, tx, r.DB, q,
		goalsFor, goalsAgainst,
		goalsFor, goalsAgainst,
		goalsFor, goalsAgainst,
		goalsFor, goalsAgainst,
		goalsFor, goalsAgainst,
		goalsFor, goalsAgainst, win, goalsFor, goalsAgainst, draw, loss,
		fairDelta,
		seasonID, teamID)
	return err
}

// Revert undoes a previously applied result (for corrections).
func (r *StandingRepo) Revert(ctx context.Context, tx *sql.Tx, seasonID, teamID int64, goalsFor, goalsAgainst, win, draw, loss, fairDelta int) error {
	q := `UPDATE standings SET
		played=GREATEST(played-1,0),
		wins=GREATEST(wins-IF(?>?,1,0),0),
		draws=GREATEST(draws-IF(?=?,1,0),0),
		losses=GREATEST(losses-IF(?<?,1,0),0),
		goals_for=goals_for-?,
		goals_against=goals_against-?,
		goal_diff=goal_diff-(?-?),
		points=points-IF(?>?, ?, IF(?=?, ?, ?)),
		fair_play=fair_play-?
		WHERE season_id=? AND team_id=?`
	_, err := execInTx(ctx, tx, r.DB, q,
		goalsFor, goalsAgainst,
		goalsFor, goalsAgainst,
		goalsFor, goalsAgainst,
		goalsFor, goalsAgainst,
		goalsFor, goalsAgainst,
		goalsFor, goalsAgainst, win, goalsFor, goalsAgainst, draw, loss,
		fairDelta,
		seasonID, teamID)
	return err
}

// Rank computes and persists ranks according to the tiebreaker list.
func (r *StandingRepo) Rank(ctx context.Context, seasonID int64, tiebreakers []string) ([]models.Standing, error) {
	cols := tiebreakerColumns(tiebreakers)
	q := fmt.Sprintf(`SELECT id,season_id,team_id,played,wins,draws,losses,goals_for,goals_against,goal_diff,points,fair_play,`+"`rank`"+`,prev_rank,
		COALESCE((SELECT name FROM teams WHERE id=standings.team_id),''),
		COALESCE((SELECT code FROM teams WHERE id=standings.team_id),'')
		FROM standings WHERE season_id=? ORDER BY %s`, cols)
	rows, err := r.QueryContext(ctx, q, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Standing{}
	rank := 1
	for rows.Next() {
		var s models.Standing
		if err := rows.Scan(&s.ID, &s.SeasonID, &s.TeamID, &s.Played, &s.Wins, &s.Draws, &s.Losses, &s.GoalsFor, &s.GoalsAgainst, &s.GoalDiff, &s.Points, &s.FairPlay, &s.Rank, &s.PrevRank, &s.TeamName, &s.TeamCode); err != nil {
			return nil, err
		}
		s.PrevRank = s.Rank
		s.Rank = rank
		out = append(out, s)
		rank++
	}
	// persist ranks
	tx, err := r.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	for _, s := range out {
		if _, err := tx.ExecContext(ctx, `UPDATE standings SET `+"`rank`"+`=?, prev_rank=? WHERE season_id=? AND team_id=?`, s.Rank, s.PrevRank, seasonID, s.TeamID); err != nil {
			return nil, rollback(tx, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

func tiebreakerColumns(tiebreakers []string) string {
	defaults := "points DESC, goal_diff DESC, goals_for DESC, fair_play ASC, team_id ASC"
	if len(tiebreakers) == 0 {
		return defaults
	}
	mapping := map[string]string{
		"points":       "points DESC",
		"goal_diff":    "goal_diff DESC",
		"goals_for":    "goals_for DESC",
		"head_to_head": "points DESC", // head-to-head needs sub-query; fallback to points
		"fair_play":    "fair_play ASC",
	}
	parts := []string{}
	for _, t := range tiebreakers {
		if col, ok := mapping[t]; ok {
			parts = append(parts, col)
		}
	}
	if len(parts) == 0 {
		return defaults
	}
	parts = append(parts, "team_id ASC")
	return joinStrings(parts, ", ")
}

func (r *StandingRepo) List(ctx context.Context, seasonID int64) ([]models.Standing, error) {
	q := `SELECT id,season_id,team_id,played,wins,draws,losses,goals_for,goals_against,goal_diff,points,fair_play,` + "`rank`" + `,prev_rank,
		COALESCE((SELECT name FROM teams WHERE id=standings.team_id),''),
		COALESCE((SELECT code FROM teams WHERE id=standings.team_id),'')
		FROM standings WHERE season_id=? ORDER BY ` + "`rank`" + ` ASC, points DESC, goal_diff DESC`
	rows, err := r.QueryContext(ctx, q, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Standing{}
	for rows.Next() {
		var s models.Standing
		if err := rows.Scan(&s.ID, &s.SeasonID, &s.TeamID, &s.Played, &s.Wins, &s.Draws, &s.Losses, &s.GoalsFor, &s.GoalsAgainst, &s.GoalDiff, &s.Points, &s.FairPlay, &s.Rank, &s.PrevRank, &s.TeamName, &s.TeamCode); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

// Snapshot persists a JSON snapshot of the current standings for a round.
func (r *StandingRepo) Snapshot(ctx context.Context, tx *sql.Tx, seasonID int64, round int, items []models.Standing) error {
	b, err := json.Marshal(items)
	if err != nil {
		return err
	}
	q := `INSERT INTO standings_snapshots (season_id,round,snapshot) VALUES (?,?,?)`
	_, err = r.ExecContext(ctx, q, seasonID, round, string(b))
	return err
}

func (r *StandingRepo) ListSnapshots(ctx context.Context, seasonID int64) ([]models.StandingsSnapshot, error) {
	rows, err := r.QueryContext(ctx, `SELECT id,season_id,round,snapshot,created_at FROM standings_snapshots WHERE season_id=? ORDER BY round DESC`, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.StandingsSnapshot{}
	for rows.Next() {
		var s models.StandingsSnapshot
		if err := rows.Scan(&s.ID, &s.SeasonID, &s.Round, &s.Snapshot, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func (r *StandingRepo) GetSnapshot(ctx context.Context, id int64) (*models.StandingsSnapshot, error) {
	var s models.StandingsSnapshot
	err := r.QueryRowContext(ctx, `SELECT id,season_id,round,snapshot,created_at FROM standings_snapshots WHERE id=?`, id).Scan(&s.ID, &s.SeasonID, &s.Round, &s.Snapshot, &s.CreatedAt)
	if err != nil {
		return nil, NotFound("snapshot", id)
	}
	return &s, nil
}

var _ = sql.ErrNoRows
