package models

import "time"

// Discipline is a punishment record against a player.
type Discipline struct {
	ID            int64     `json:"id" db:"id"`
	PlayerID      int64     `json:"player_id" db:"player_id"`
	TeamID        int64     `json:"team_id" db:"team_id"`
	SeasonID      int64     `json:"season_id" db:"season_id"`
	MatchID       *int64    `json:"match_id,omitempty" db:"match_id"`
	Punishment    string    `json:"punishment" db:"punishment"`
	Reason        string    `json:"reason" db:"reason"`
	Severity      string    `json:"severity" db:"severity"`
	SuspendGames  int       `json:"suspend_games" db:"suspend_games"`
	FineAmount    float64   `json:"fine_amount" db:"fine_amount"`
	Status        string    `json:"status" db:"status"`
	IssuedBy      int64     `json:"issued_by" db:"issued_by"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

const (
	PunishYellow      = "yellow"
	PunishRed         = "red"
	PunishFine        = "fine"
	PunishSuspension  = "suspension"
	PunishWarning     = "warning"
)

const (
	SeverityMinor   = "minor"
	SeverityMedium  = "medium"
	SeveritySevere  = "severe"
)

const (
	DisciplineStatusActive     = "active"
	DisciplineStatusServed     = "served"
	DisciplineStatusAppealed   = "appealed"
	DisciplineStatusOverturned = "overturned"
)

// Suspension is the active serving state of a suspension punishment.
type Suspension struct {
	ID           int64      `json:"id" db:"id"`
	DisciplineID int64      `json:"discipline_id" db:"discipline_id"`
	PlayerID     int64      `json:"player_id" db:"player_id"`
	TotalGames   int        `json:"total_games" db:"total_games"`
	ServedGames  int        `json:"served_games" db:"served_games"`
	StartDate    time.Time  `json:"start_date" db:"start_date"`
	EndDate      *time.Time `json:"end_date,omitempty" db:"end_date"`
	Status       string     `json:"status" db:"status"`
}

const (
	SuspensionStatusActive = "active"
	SuspensionStatusServed = "served"
)

// Appeal is a lodged challenge against a discipline.
type Appeal struct {
	ID            int64      `json:"id" db:"id"`
	DisciplineID  int64      `json:"discipline_id" db:"discipline_id"`
	AppellantID   int64      `json:"appellant_id" db:"appellant_id"`
	Reason        string     `json:"reason" db:"reason"`
	Status        string     `json:"status" db:"status"`
	ReviewerID    *int64     `json:"reviewer_id,omitempty" db:"reviewer_id"`
	ReviewOpinion string     `json:"review_opinion,omitempty" db:"review_opinion"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty" db:"reviewed_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

const (
	AppealStatusPending  = "pending"
	AppealStatusReviewed = "reviewed"
	AppealStatusAccepted = "accepted"
	AppealStatusRejected = "rejected"
)
