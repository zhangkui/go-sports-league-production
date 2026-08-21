package verify

import (
	"testing"

	"github.com/goxm2/sports-league/internal/models"
)

func TestBug011_BusinessRegression(t *testing.T) {
	runBug011BusinessRegression(t)
}

func runBug011BusinessRegression(t *testing.T) {
	f := newFixture(t)
	seasonID := f.seedSeason("bug011", models.SeasonStatusRegistration)
	team, createErr := f.teams.Create(f.ctx, models.CreateTeamRequest{Code: "BUG011", Name: "Registration Team", SeasonID: seasonID}, 1)
	if createErr != nil {
		t.Fatalf("create team: %v", createErr)
	}
	if team == nil || team.ID == 0 {
		t.Fatalf("team creation returned no persisted team: %#v", team)
	}
	if registrations := f.count(`SELECT COUNT(*) FROM team_registrations WHERE team_id=? AND season_id=? AND status='pending'`, team.ID, seasonID); registrations != 1 {
		t.Errorf("initial pending registration count = %d, want 1 for created team %d", registrations, team.ID)
	}
	if orphaned := f.count(`SELECT COUNT(*) FROM team_registrations WHERE team_id=0`); orphaned != 0 {
		t.Errorf("found %d orphaned initial registrations with team_id=0", orphaned)
	}
}
