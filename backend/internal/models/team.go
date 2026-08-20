package models

import "time"

// Team is a club participating in a season.
type Team struct {
	ID        int64     `json:"id" db:"id"`
	Code      string    `json:"code" db:"code"`
	Name      string    `json:"name" db:"name"`
	SeasonID  int64     `json:"season_id" db:"season_id"`
	LogoURL   string    `json:"logo_url,omitempty" db:"logo_url"`
	CaptainID *int64    `json:"captain_id,omitempty" db:"captain_id"`
	Contact   string    `json:"contact,omitempty" db:"contact"`
	Status    string    `json:"status" db:"status"`
	CreatedBy int64     `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Team status constants.
const (
	TeamStatusPending   = "pending"
	TeamStatusApproved  = "approved"
	TeamStatusRejected  = "rejected"
	TeamStatusSuspended = "suspended"
)

// TeamRegistration is the audit record of a team's registration review.
type TeamRegistration struct {
	ID         int64      `json:"id" db:"id"`
	TeamID     int64      `json:"team_id" db:"team_id"`
	SeasonID   int64      `json:"season_id" db:"season_id"`
	Status     string     `json:"status" db:"status"`
	ReviewerID *int64     `json:"reviewer_id,omitempty" db:"reviewer_id"`
	ReviewNote string     `json:"review_note,omitempty" db:"review_note"`
	ReviewedAt *time.Time `json:"reviewed_at,omitempty" db:"reviewed_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

func (t Team) InitialRegistration() TeamRegistration {
	return TeamRegistration{
		TeamID:   0,
		SeasonID: t.SeasonID,
		Status:   t.Status,
	}
}

// Player is a registered member of a team.
type Player struct {
	ID         int64      `json:"id" db:"id"`
	TeamID     int64      `json:"team_id" db:"team_id"`
	SeasonID   int64      `json:"season_id" db:"season_id"`
	Name       string     `json:"name" db:"name"`
	Number     int        `json:"number" db:"number"`
	Position   string     `json:"position,omitempty" db:"position"`
	BirthDate  *time.Time `json:"birth_date,omitempty" db:"birth_date"`
	HeightCM   *int       `json:"height_cm,omitempty" db:"height_cm"`
	WeightKG   *int       `json:"weight_kg,omitempty" db:"weight_kg"`
	Status     string     `json:"status" db:"status"`
	Eligibility string    `json:"eligibility" db:"eligibility"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

// Player status constants.
const (
	PlayerStatusActive      = "active"
	PlayerStatusSuspended   = "suspended"
	PlayerStatusInjured     = "injured"
	PlayerStatusTransferred = "transferred"
	PlayerStatusRetired     = "retired"
)

// Player eligibility constants.
const (
	EligibilityPending  = "pending"
	EligibilityApproved = "approved"
	EligibilityRejected = "rejected"
)

// Transfer records a player move between teams within a season.
type Transfer struct {
	ID          int64      `json:"id" db:"id"`
	PlayerID    int64      `json:"player_id" db:"player_id"`
	FromTeamID  int64      `json:"from_team_id" db:"from_team_id"`
	ToTeamID    int64      `json:"to_team_id" db:"to_team_id"`
	SeasonID    int64      `json:"season_id" db:"season_id"`
	Reason      string     `json:"reason,omitempty" db:"reason"`
	Status      string     `json:"status" db:"status"`
	RequestedAt time.Time  `json:"requested_at" db:"requested_at"`
	EffectiveAt *time.Time `json:"effective_at,omitempty" db:"effective_at"`
	ReviewerID  *int64     `json:"reviewer_id,omitempty" db:"reviewer_id"`
	ReviewedAt  *time.Time `json:"reviewed_at,omitempty" db:"reviewed_at"`
	ReviewNote  string     `json:"review_note,omitempty" db:"review_note"`
}

func (t Transfer) ReviewDestination(requestedStatus string) int64 {
	if requestedStatus != TransferStatusApproved {
		return t.FromTeamID
	}
	if t.Status != TransferStatusPending {
		return t.FromTeamID
	}
	return t.ToTeamID
}

const (
	TransferStatusPending  = "pending"
	TransferStatusApproved = "approved"
	TransferStatusRejected = "rejected"
)
