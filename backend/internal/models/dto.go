package models

import "time"

// ===== Auth & User DTOs =====

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type ResetPasswordRequest struct {
	NewPassword string `json:"new_password"`
}

type UpdateUserRequest struct {
	FullName *string `json:"full_name,omitempty"`
	Phone    *string `json:"phone,omitempty"`
	Email    *string `json:"email,omitempty"`
	Status   *int    `json:"status,omitempty"`
}

type AssignRolesRequest struct {
	RoleIDs []int64 `json:"role_ids"`
}

type CreateRoleRequest struct {
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Permissions []int64 `json:"permission_ids,omitempty"`
}

type UpdateRoleRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Permissions []int64 `json:"permission_ids,omitempty"`
}

// ===== Season DTOs =====

type CreateSeasonRequest struct {
	Code              string `json:"code"`
	Name              string `json:"name"`
	Sport             string `json:"sport"`
	Division          string `json:"division,omitempty"`
	Format            string `json:"format,omitempty"`
	TeamCount         int    `json:"team_count,omitempty"`
	Rounds            int    `json:"rounds,omitempty"`
	StartDate         string `json:"start_date,omitempty"`
	EndDate           string `json:"end_date,omitempty"`
	RegistrationStart string `json:"registration_start,omitempty"`
	RegistrationEnd   string `json:"registration_end,omitempty"`
}

type UpdateSeasonRequest struct {
	Name              *string `json:"name,omitempty"`
	Division          *string `json:"division,omitempty"`
	TeamCount         *int    `json:"team_count,omitempty"`
	Rounds            *int    `json:"rounds,omitempty"`
	StartDate         *string `json:"start_date,omitempty"`
	EndDate           *string `json:"end_date,omitempty"`
	RegistrationStart *string `json:"registration_start,omitempty"`
	RegistrationEnd   *string `json:"registration_end,omitempty"`
}

type SeasonStatusChange struct {
	Status string `json:"status"`
}

type ScoringRuleInput struct {
	WinPoints   int    `json:"win_points"`
	DrawPoints  int    `json:"draw_points"`
	LossPoints  int    `json:"loss_points"`
	Tiebreakers string `json:"tiebreakers,omitempty"`
}

// ===== Team & Player DTOs =====

type CreateTeamRequest struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	SeasonID  int64  `json:"season_id"`
	LogoURL   string `json:"logo_url,omitempty"`
	CaptainID *int64 `json:"captain_id,omitempty"`
	Contact   string `json:"contact,omitempty"`
}

type UpdateTeamRequest struct {
	Name      *string `json:"name,omitempty"`
	LogoURL   *string `json:"logo_url,omitempty"`
	CaptainID *int64  `json:"captain_id,omitempty"`
	Contact   *string `json:"contact,omitempty"`
	Status    *string `json:"status,omitempty"`
}

type TeamRegistrationReview struct {
	Status string `json:"status"`
	Note   string `json:"note,omitempty"`
}

type CreatePlayerRequest struct {
	TeamID    int64  `json:"team_id"`
	Name      string `json:"name"`
	Number    int    `json:"number"`
	Position  string `json:"position,omitempty"`
	BirthDate string `json:"birth_date,omitempty"`
	HeightCM  *int   `json:"height_cm,omitempty"`
	WeightKG  *int   `json:"weight_kg,omitempty"`
}

type UpdatePlayerRequest struct {
	Name       *string `json:"name,omitempty"`
	Number     *int    `json:"number,omitempty"`
	Position   *string `json:"position,omitempty"`
	BirthDate  *string `json:"birth_date,omitempty"`
	HeightCM   *int    `json:"height_cm,omitempty"`
	WeightKG   *int    `json:"weight_kg,omitempty"`
	Status     *string `json:"status,omitempty"`
	Eligibility *string `json:"eligibility,omitempty"`
}

type CreateTransferRequest struct {
	PlayerID   int64  `json:"player_id"`
	ToTeamID   int64  `json:"to_team_id"`
	Reason     string `json:"reason,omitempty"`
	EffectiveAt string `json:"effective_at,omitempty"`
}

type TransferReview struct {
	Status string `json:"status"`
	Note   string `json:"note,omitempty"`
}

// ===== Venue & Schedule DTOs =====

type CreateVenueRequest struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Address  string `json:"address,omitempty"`
	Capacity *int   `json:"capacity,omitempty"`
	Sport    string `json:"sport,omitempty"`
	Status   string `json:"status,omitempty"`
}

type UpdateVenueRequest struct {
	Name     *string `json:"name,omitempty"`
	Address  *string `json:"address,omitempty"`
	Capacity *int    `json:"capacity,omitempty"`
	Sport    *string `json:"sport,omitempty"`
	Status   *string `json:"status,omitempty"`
}

type VenueAvailabilityInput struct {
	Weekday   int    `json:"weekday"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type GenerateScheduleRequest struct {
	SeasonID    int64 `json:"season_id"`
	Format      string `json:"format,omitempty"`
	VenueID     int64 `json:"venue_id,omitempty"`
	StartDate   string `json:"start_date,omitempty"`
	GamesPerDay int   `json:"games_per_day,omitempty"`
}

type UpdateScheduleRequest struct {
	VenueID    *int64  `json:"venue_id,omitempty"`
	MatchDate  *string `json:"match_date,omitempty"`
	StartTime *string `json:"start_time,omitempty"`
	Status    *string `json:"status,omitempty"`
}

// ===== Match DTOs =====

type RecordMatchRequest struct {
	HomeScore      *int `json:"home_score,omitempty"`
	AwayScore      *int `json:"away_score,omitempty"`
	HomeHalfScore  *int `json:"home_half_score,omitempty"`
	AwayHalfScore  *int `json:"away_half_score,omitempty"`
	RefereeID      *int64 `json:"referee_id,omitempty"`
	RecorderID     *int64 `json:"recorder_id,omitempty"`
	DurationMin    *int `json:"duration_min,omitempty"`
	Status         *string `json:"status,omitempty"`
}

type CreateMatchEventRequest struct {
	TeamID     int64  `json:"team_id"`
	PlayerID   *int64 `json:"player_id,omitempty"`
	EventType  string `json:"event_type"`
	Minute     int    `json:"minute"`
	Description string `json:"description,omitempty"`
}

type CreatePlayerStatRequest struct {
	PlayerID  int64   `json:"player_id"`
	TeamID    int64   `json:"team_id"`
	IsStarter bool    `json:"is_starter"`
	PlayedMin *int    `json:"played_min,omitempty"`
	Position  string  `json:"position,omitempty"`
	Rating    *float64 `json:"rating,omitempty"`
}

// ===== Discipline DTOs =====

type CreateDisciplineRequest struct {
	PlayerID     int64   `json:"player_id"`
	TeamID       int64   `json:"team_id"`
	SeasonID     int64   `json:"season_id"`
	MatchID      *int64  `json:"match_id,omitempty"`
	Punishment   string  `json:"punishment"`
	Reason       string  `json:"reason"`
	Severity     string  `json:"severity,omitempty"`
	SuspendGames int     `json:"suspend_games,omitempty"`
	FineAmount   float64 `json:"fine_amount,omitempty"`
}

type AppealReviewInput struct {
	Status   string `json:"status"`
	Opinion  string `json:"opinion,omitempty"`
}

type CreateAppealRequest struct {
	DisciplineID int64  `json:"discipline_id"`
	Reason       string  `json:"reason"`
}

// SeasonSummary is the aggregate report row.
type SeasonSummary struct {
	Season        Season         `json:"season"`
	TeamCount     int64          `json:"team_count"`
	PlayerCount   int64          `json:"player_count"`
	MatchCount    int64          `json:"match_count"`
	CompletedMatches int64       `json:"completed_matches"`
	TopScorer     *PlayerRanking `json:"top_scorer,omitempty"`
	TopTeam       *Standing      `json:"top_team,omitempty"`
}

// PlayerRanking is a derived stats row for the player leaderboard.
type PlayerRanking struct {
	PlayerID    int64  `json:"player_id" db:"player_id"`
	PlayerName  string `json:"player_name" db:"player_name"`
	TeamID      int64  `json:"team_id" db:"team_id"`
	TeamName    string `json:"team_name" db:"team_name"`
	Goals       int    `json:"goals" db:"goals"`
	Assists     int    `json:"assists" db:"assists"`
	YellowCards int    `json:"yellow_cards" db:"yellow_cards"`
	RedCards    int    `json:"red_cards" db:"red_cards"`
	Matches     int    `json:"matches" db:"matches"`
	AvgRating   float64 `json:"avg_rating" db:"avg_rating"`
}

var _ = time.Now
