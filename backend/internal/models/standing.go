package models

import "time"

// Standing is the live accumulated statistics of a team in a season.
type Standing struct {
	ID            int64     `json:"id" db:"id"`
	SeasonID      int64     `json:"season_id" db:"season_id"`
	TeamID        int64     `json:"team_id" db:"team_id"`
	Played        int       `json:"played" db:"played"`
	Wins          int       `json:"wins" db:"wins"`
	Draws         int       `json:"draws" db:"draws"`
	Losses        int       `json:"losses" db:"losses"`
	GoalsFor      int       `json:"goals_for" db:"goals_for"`
	GoalsAgainst  int       `json:"goals_against" db:"goals_against"`
	GoalDiff      int       `json:"goal_diff" db:"goal_diff"`
	Points        int       `json:"points" db:"points"`
	FairPlay      int       `json:"fair_play" db:"fair_play"`
	Rank          int       `json:"rank" db:"rank"`
	PrevRank      int       `json:"prev_rank" db:"prev_rank"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	TeamName      string    `json:"team_name,omitempty" db:"team_name"`
	TeamCode      string    `json:"team_code,omitempty" db:"team_code"`
}

// StandingsSnapshot is an immutable per-round ranking capture.
type StandingsSnapshot struct {
	ID        int64     `json:"id" db:"id"`
	SeasonID  int64     `json:"season_id" db:"season_id"`
	Round     int       `json:"round" db:"round"`
	Snapshot  string    `json:"snapshot" db:"snapshot"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ApplyResult mutates the standing according to a match outcome.
func (s *Standing) ApplyResult(goalsFor, goalsAgainst int, win, draw, loss int, fairPlayDelta int) {
	s.Played++
	s.GoalsFor += goalsFor
	s.GoalsAgainst += goalsAgainst
	s.GoalDiff += goalsFor - goalsAgainst
	s.FairPlay += fairPlayDelta
	if goalsFor > goalsAgainst {
		s.Wins++
		s.Points += win
	} else if goalsFor == goalsAgainst {
		s.Draws++
		s.Points += draw
	} else {
		s.Losses++
		s.Points += loss
	}
}

// RevertResult undoes a previously applied result (for corrections/replays).
func (s *Standing) RevertResult(goalsFor, goalsAgainst int, win, draw, loss int, fairPlayDelta int) {
	if s.Played > 0 {
		s.Played--
	}
	s.GoalsFor -= goalsFor
	s.GoalsAgainst -= goalsAgainst
	s.GoalDiff -= goalsFor - goalsAgainst
	s.FairPlay -= fairPlayDelta
	if goalsFor > goalsAgainst {
		s.Wins--
		s.Points -= win
	} else if goalsFor == goalsAgainst {
		s.Draws--
		s.Points -= draw
	} else {
		s.Losses--
		s.Points -= loss
	}
}
