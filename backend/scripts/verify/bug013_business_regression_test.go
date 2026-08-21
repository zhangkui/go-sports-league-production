package verify

import (
	"testing"

	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/errorsx"
)

func TestBug013_BusinessRegression(t *testing.T) {
	runBug013BusinessRegression(t)
}

func runBug013BusinessRegression(t *testing.T) {
	f := newFixture(t)
	seasonID := f.seedSeason("bug013", models.SeasonStatusRegistration)
	teamID := f.seedTeam(seasonID, "BUG013")
	f.seedPlayer(seasonID, teamID, "Existing Player", 9)
	player, createErr := f.teams.CreatePlayer(f.ctx, models.CreatePlayerRequest{TeamID: teamID, Name: "Duplicate Number", Number: 9, Position: "FWD"})
	if !errorsx.IsConflict(createErr) {
		t.Errorf("duplicate shirt number error = %v, want conflict classification", createErr)
	}
	if player != nil {
		t.Errorf("duplicate shirt number returned player: %#v", player)
	}
	if count := f.count(`SELECT COUNT(*) FROM players WHERE team_id=? AND number=9`, teamID); count != 1 {
		t.Errorf("players with duplicate team number = %d, want 1", count)
	}
}
