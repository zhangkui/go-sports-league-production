package verify

import (
	"testing"

	"github.com/goxm2/sports-league/internal/models"
)

func TestBug002_BusinessRegression(t *testing.T) {
	runBug002BusinessRegression(t)
}

func runBug002BusinessRegression(t *testing.T) {
	f := newFixture(t)
	season, err := f.seasons.Create(f.ctx, models.CreateSeasonRequest{Code: "BUG002", Name: "Rule Link Season", Sport: "football", Format: "round_robin"}, 1)
	if err != nil {
		t.Fatalf("create season: %v", err)
	}
	if season.CurrentRuleVersion != 1 {
		t.Errorf("returned current rule version = %d, want 1", season.CurrentRuleVersion)
	}
	var persisted int
	if err := f.db.QueryRow(`SELECT current_rule_version FROM seasons WHERE id=?`, season.ID).Scan(&persisted); err != nil {
		t.Fatalf("read persisted rule pointer: %v", err)
	}
	if persisted != 1 {
		t.Errorf("persisted current rule version = %d, want 1", persisted)
	}
	rule, err := f.seasons.GetActiveScoringRule(f.ctx, season.ID)
	if err != nil {
		t.Fatalf("load active rule: %v", err)
	}
	if rule == nil || rule.Version != 1 || !rule.IsActive {
		t.Errorf("created season active rule = %#v, want active version 1", rule)
	}
	if versions := f.count(`SELECT COUNT(*) FROM scoring_rule_versions WHERE season_id=? AND version=1`, season.ID); versions != 1 {
		t.Errorf("scoring rule snapshot count = %d, want 1", versions)
	}
}
