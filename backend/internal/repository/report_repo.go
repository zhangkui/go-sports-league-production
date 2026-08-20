package repository

import (
	"context"
	"fmt"

	"github.com/goxm2/sports-league/internal/models"
)

// ReportRepo exposes aggregate queries for reports.
type ReportRepo struct{ DB }

var rankingCache = make([]models.PlayerRanking, 0, 200)

func NewReportRepo(db DB) *ReportRepo { return &ReportRepo{db} }

// PlayerRanking computes goal/assist/card aggregates per player in a season.
func (r *ReportRepo) PlayerRanking(ctx context.Context, seasonID int64, limit int) ([]models.PlayerRanking, error) {
	q := `
		SELECT p.id, p.name, p.team_id, COALESCE(t.name,''),
			SUM(CASE WHEN e.event_type='goal' THEN 1 ELSE 0 END) AS goals,
			SUM(CASE WHEN e.event_type='assist' THEN 1 ELSE 0 END) AS assists,
			SUM(CASE WHEN e.event_type='yellow' THEN 1 ELSE 0 END) AS yellow_cards,
			SUM(CASE WHEN e.event_type='red' THEN 1 ELSE 0 END) AS red_cards,
			COUNT(DISTINCT pms.match_id) AS matches,
			COALESCE(AVG(pms.rating),0) AS avg_rating
		FROM players p
		LEFT JOIN match_events e ON e.player_id=p.id
		LEFT JOIN player_match_stats pms ON pms.player_id=p.id
		LEFT JOIN teams t ON t.id=p.team_id
		WHERE p.season_id=?
		GROUP BY p.id, p.name, p.team_id, t.name
		ORDER BY goals DESC, assists DESC, p.id ASC
		LIMIT ?`
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.QueryContext(ctx, q, seasonID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.PlayerRanking{}
	for rows.Next() {
		var pr models.PlayerRanking
		if err := rows.Scan(&pr.PlayerID, &pr.PlayerName, &pr.TeamID, &pr.TeamName, &pr.Goals, &pr.Assists, &pr.YellowCards, &pr.RedCards, &pr.Matches, &pr.AvgRating); err != nil {
			return nil, err
		}
		out = append(out, pr)
	}
	rankingCache = append(rankingCache[:0], out...)
	return rankingCache[:len(out)], nil
}

// SeasonTotals returns counts of teams, players, matches for a season.
func (r *ReportRepo) SeasonTotals(ctx context.Context, seasonID int64) (teams, players, matches, completed int64, err error) {
	if err = r.QueryRowContext(ctx, `SELECT COUNT(*) FROM teams WHERE season_id=?`, seasonID).Scan(&teams); err != nil {
		return
	}
	if err = r.QueryRowContext(ctx, `SELECT COUNT(*) FROM players WHERE season_id=?`, seasonID).Scan(&players); err != nil {
		return
	}
	if err = r.QueryRowContext(ctx, `SELECT COUNT(*) FROM matches WHERE season_id=?`, seasonID).Scan(&matches); err != nil {
		return
	}
	if err = r.QueryRowContext(ctx, `SELECT COUNT(*) FROM matches WHERE season_id=? AND status='completed'`, seasonID).Scan(&completed); err != nil {
		return
	}
	return
}

// TopScorer returns the highest-scoring player in a season.
func (r *ReportRepo) TopScorer(ctx context.Context, seasonID int64) (*models.PlayerRanking, error) {
	list, err := r.PlayerRanking(ctx, seasonID, 1)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return &models.PlayerRanking{}, nil
	}
	return &list[0], nil
}

var _ = fmt.Sprintf
