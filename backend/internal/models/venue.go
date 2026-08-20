package models

import "time"

// Venue is a physical location where matches are played.
type Venue struct {
	ID        int64     `json:"id" db:"id"`
	Code      string    `json:"code" db:"code"`
	Name      string    `json:"name" db:"name"`
	Address   string    `json:"address,omitempty" db:"address"`
	Capacity  *int      `json:"capacity,omitempty" db:"capacity"`
	Sport     string    `json:"sport,omitempty" db:"sport"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

const (
	VenueStatusAvailable   = "available"
	VenueStatusMaintenance = "maintenance"
	VenueStatusUnavailable = "unavailable"
)

// VenueAvailability is a weekly recurring time window a venue is usable.
type VenueAvailability struct {
	ID        int64     `json:"id" db:"id"`
	VenueID   int64     `json:"venue_id" db:"venue_id"`
	Weekday   int       `json:"weekday" db:"weekday"`
	StartTime string    `json:"start_time" db:"start_time"`
	EndTime   string    `json:"end_time" db:"end_time"`
}

// Schedule is a planned fixture: round, teams, venue, date/time.
type Schedule struct {
	ID         int64     `json:"id" db:"id"`
	SeasonID   int64     `json:"season_id" db:"season_id"`
	Round      int       `json:"round" db:"round"`
	HomeTeamID int64     `json:"home_team_id" db:"home_team_id"`
	AwayTeamID int64     `json:"away_team_id" db:"away_team_id"`
	VenueID    int64     `json:"venue_id" db:"venue_id"`
	MatchDate  time.Time `json:"match_date" db:"match_date"`
	StartTime  string    `json:"start_time" db:"start_time"`
	Status     string    `json:"status" db:"status"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

const (
	ScheduleStatusScheduled  = "scheduled"
	ScheduleStatusConfirmed  = "confirmed"
	ScheduleStatusInProgress = "in_progress"
	ScheduleStatusCompleted  = "completed"
	ScheduleStatusCancelled  = "cancelled"
	ScheduleStatusPostponed  = "postponed"
)

// ScheduleConflict records a detected scheduling violation.
type ScheduleConflict struct {
	ID          int64     `json:"id" db:"id"`
	ScheduleID  int64     `json:"schedule_id" db:"schedule_id"`
	ConflictType string   `json:"conflict_type" db:"conflict_type"`
	Description string    `json:"description,omitempty" db:"description"`
	DetectedAt  time.Time `json:"detected_at" db:"detected_at"`
}

const (
	ConflictTypeTeam  = "team"
	ConflictTypeVenue = "venue"
	ConflictTypeRound = "round"
	ConflictTypeWindow = "window"
)
