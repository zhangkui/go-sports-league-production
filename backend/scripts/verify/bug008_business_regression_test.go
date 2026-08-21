package verify

import (
	"testing"

	"github.com/goxm2/sports-league/internal/models"
)

func TestBug008_BusinessRegression(t *testing.T) {
	fixture := newFixture(t)
	seasonID := fixture.seedSeason("bug008", models.SeasonStatusDraft)
	fixture.seedRule(seasonID)

	changed, err := fixture.seasons.ChangeStatus(fixture.ctx, seasonID, models.SeasonStatusCompleted)
	if err == nil {
		t.Errorf("draft season illegally transitioned directly to completed: %#v", changed)
	}

	status, ruleVersion := loadSeasonLifecycle(t, fixture, seasonID)
	if status != models.SeasonStatusDraft || ruleVersion != 1 {
		t.Errorf("illegal transition changed persisted state to status=%s rule_version=%d", status, ruleVersion)
	}

	changed, err = fixture.seasons.ChangeStatus(fixture.ctx, seasonID, models.SeasonStatusRegistration)
	if err != nil {
		t.Fatalf("valid draft to registration transition: %v", err)
	}
	if changed.Status != models.SeasonStatusRegistration || changed.CurrentRuleVersion != 1 {
		t.Errorf("valid transition response = status=%s rule_version=%d, want registration/1", changed.Status, changed.CurrentRuleVersion)
	}

	status, ruleVersion = loadSeasonLifecycle(t, fixture, seasonID)
	if status != models.SeasonStatusRegistration || ruleVersion != 1 {
		t.Errorf("valid transition persisted state = status=%s rule_version=%d, want registration/1", status, ruleVersion)
	}
}

func loadSeasonLifecycle(t *testing.T, fixture *fixture, seasonID int64) (string, int) {
	t.Helper()
	var status string
	var ruleVersion int
	if err := fixture.db.QueryRowContext(fixture.ctx,
		"SELECT status,current_rule_version FROM seasons WHERE id=?", seasonID,
	).Scan(&status, &ruleVersion); err != nil {
		t.Fatalf("load season lifecycle: %v", err)
	}
	return status, ruleVersion
}
