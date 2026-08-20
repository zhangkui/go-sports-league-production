package models

import "time"

// Season represents a league season for a sport and division.
type Season struct {
	ID                int64      `json:"id" db:"id"`
	Code              string     `json:"code" db:"code"`
	Name              string     `json:"name" db:"name"`
	Sport             string     `json:"sport" db:"sport"`
	Division          string     `json:"division,omitempty" db:"division"`
	TeamCount         int        `json:"team_count" db:"team_count"`
	Format            string     `json:"format" db:"format"`
	Rounds            int        `json:"rounds" db:"rounds"`
	StartDate         *time.Time `json:"start_date,omitempty" db:"start_date"`
	EndDate           *time.Time `json:"end_date,omitempty" db:"end_date"`
	RegistrationStart *time.Time `json:"registration_start,omitempty" db:"registration_start"`
	RegistrationEnd   *time.Time `json:"registration_end,omitempty" db:"registration_end"`
	Status            string     `json:"status" db:"status"`
	CurrentRuleVersion int       `json:"current_rule_version" db:"current_rule_version"`
	CreatedBy         int64      `json:"created_by" db:"created_by"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

// Season status constants.
const (
	SeasonStatusDraft        = "draft"
	SeasonStatusRegistration = "registration"
	SeasonStatusOpen         = "open"
	SeasonStatusOngoing      = "ongoing"
	SeasonStatusCompleted    = "completed"
	SeasonStatusArchived     = "archived"
	SeasonStatusCancelled    = "cancelled"
)

// ScoringRule is the active point/tie-break configuration for a season.
type ScoringRule struct {
	ID           int64     `json:"id" db:"id"`
	SeasonID     int64     `json:"season_id" db:"season_id"`
	Version      int       `json:"version" db:"version"`
	WinPoints    int       `json:"win_points" db:"win_points"`
	DrawPoints   int       `json:"draw_points" db:"draw_points"`
	LossPoints   int       `json:"loss_points" db:"loss_points"`
	Tiebreakers  string    `json:"tiebreakers" db:"tiebreakers"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	CreatedBy    int64     `json:"created_by" db:"created_by"`
}

// TiebreakerOrder parses the comma-separated tiebreaker list.
func (s ScoringRule) TiebreakerOrder() []string {
	return splitCSV(s.Tiebreakers)
}

func splitCSV(s string) []string {
	out := []string{}
	cur := ""
	for _, r := range s {
		if r == ',' {
			if cur != "" {
				out = append(out, cur)
			}
			cur = ""
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

// ScoringRuleVersion is an immutable snapshot of a scoring rule.
type ScoringRuleVersion struct {
	ID        int64     `json:"id" db:"id"`
	SeasonID  int64     `json:"season_id" db:"season_id"`
	Version   int       `json:"version" db:"version"`
	Snapshot  string    `json:"snapshot" db:"snapshot"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	CreatedBy int64     `json:"created_by" db:"created_by"`
}
