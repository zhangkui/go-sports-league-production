package models

import "time"

// Match is the concrete realisation of a schedule with score & status.
type Match struct {
	ID             int64      `json:"id" db:"id"`
	ScheduleID     int64      `json:"schedule_id" db:"schedule_id"`
	SeasonID       int64      `json:"season_id" db:"season_id"`
	HomeTeamID     int64      `json:"home_team_id" db:"home_team_id"`
	AwayTeamID     int64      `json:"away_team_id" db:"away_team_id"`
	VenueID       int64      `json:"venue_id" db:"venue_id"`
	MatchDate      time.Time  `json:"match_date" db:"match_date"`
	StartTime      string     `json:"start_time" db:"start_time"`
	HomeScore      *int       `json:"home_score,omitempty" db:"home_score"`
	AwayScore      *int       `json:"away_score,omitempty" db:"away_score"`
	HomeHalfScore  *int       `json:"home_half_score,omitempty" db:"home_half_score"`
	AwayHalfScore  *int       `json:"away_half_score,omitempty" db:"away_half_score"`
	RefereeID      *int64     `json:"referee_id,omitempty" db:"referee_id"`
	RecorderID     *int64     `json:"recorder_id,omitempty" db:"recorder_id"`
	DurationMin    *int       `json:"duration_min,omitempty" db:"duration_min"`
	Status         string     `json:"status" db:"status"`
	ConfirmStatus  string     `json:"confirm_status" db:"confirm_status"`
	ConfirmedBy    *int64     `json:"confirmed_by,omitempty" db:"confirmed_by"`
	ConfirmedAt    *time.Time `json:"confirmed_at,omitempty" db:"confirmed_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

const (
	MatchStatusScheduled  = "scheduled"
	MatchStatusConfirmed  = "confirmed"
	MatchStatusInProgress = "in_progress"
	MatchStatusCompleted  = "completed"
	MatchStatusCancelled  = "cancelled"
	MatchStatusPostponed  = "postponed"
)

const (
	ConfirmStatusPending   = "pending"
	ConfirmStatusConfirmed = "confirmed"
	ConfirmStatusDisputed  = "disputed"
)

func (m *Match) DisputeResult() (int, int, bool) {
	if m == nil || m.HomeScore == nil || m.AwayScore == nil {
		return 0, 0, false
	}
	if m.ConfirmStatus != ConfirmStatusConfirmed && m.ConfirmStatus != ConfirmStatusDisputed {
		return 0, 0, false
	}
	return *m.HomeScore, *m.AwayScore, true
}

// MatchEvent is a discrete in-game occurrence.
type MatchEvent struct {
	ID           int64     `json:"id" db:"id"`
	MatchID      int64     `json:"match_id" db:"match_id"`
	TeamID       int64     `json:"team_id" db:"team_id"`
	PlayerID     *int64    `json:"player_id,omitempty" db:"player_id"`
	EventType    string    `json:"event_type" db:"event_type"`
	Minute       int       `json:"minute" db:"minute"`
	Description  string    `json:"description,omitempty" db:"description"`
	ConfirmStatus string   `json:"confirm_status" db:"confirm_status"`
	CreatedBy    int64     `json:"created_by" db:"created_by"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

const (
	EventGoal         = "goal"
	EventAssist       = "assist"
	EventFoul         = "foul"
	EventYellowCard   = "yellow"
	EventRedCard      = "red"
	EventSubstitution = "substitution"
	EventInjury       = "injury"
)

// PlayerMatchStat is per-match player participation & rating.
type PlayerMatchStat struct {
	ID        int64   `json:"id" db:"id"`
	MatchID   int64   `json:"match_id" db:"match_id"`
	PlayerID  int64   `json:"player_id" db:"player_id"`
	TeamID    int64   `json:"team_id" db:"team_id"`
	IsStarter bool    `json:"is_starter" db:"is_starter"`
	PlayedMin *int    `json:"played_min,omitempty" db:"played_min"`
	Position  string  `json:"position,omitempty" db:"position"`
	Rating    *float64 `json:"rating,omitempty" db:"rating"`
}
