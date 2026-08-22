package verify

import (
	"database/sql"
	"testing"
	"time"

	"github.com/goxm2/sports-league/internal/models"
)

func TestBug003_BusinessRegression(t *testing.T) {
	f := newFixture(t)
	seasonID := f.seedSeason("BUG003", "active")
	teamID := f.seedTeam(seasonID, "BUG003-T")
	playerID := f.seedPlayer(seasonID, teamID, "Appeal Player", 31)
	issuerID := f.seedUser("bug003-issuer", 1, "unused")
	disciplineID := f.insertID(`INSERT INTO disciplines
		(player_id,team_id,season_id,punishment,reason,severity,suspend_games,fine_amount,status,issued_by)
		VALUES (?,?,?,?,?,?,?,?,?,?)`,
		playerID, teamID, seasonID, models.PunishSuspension, "appeal accepted", models.SeveritySevere,
		3, 0, models.DisciplineStatusActive, issuerID)
	f.exec(`INSERT INTO suspensions
		(discipline_id,player_id,total_games,served_games,start_date,end_date,status)
		VALUES (?,?,?,?,?,?,?)`,
		disciplineID, playerID, 3, 0, time.Now(), nil, models.SuspensionStatusActive)

	overturned, err := f.disciplines.Overturn(f.ctx, disciplineID)
	if err != nil {
		t.Fatalf("overturn discipline: %v", err)
	}
	if overturned.Status != models.DisciplineStatusOverturned {
		t.Errorf("overturn returned status %q, want %q", overturned.Status, models.DisciplineStatusOverturned)
	}

	var disciplineStatus string
	if err := f.db.QueryRowContext(f.ctx, `SELECT status FROM disciplines WHERE id=?`, disciplineID).Scan(&disciplineStatus); err != nil {
		t.Fatalf("read discipline status: %v", err)
	}
	if disciplineStatus != models.DisciplineStatusOverturned {
		t.Errorf("stored discipline status %q, want %q", disciplineStatus, models.DisciplineStatusOverturned)
	}

	var suspensionStatus string
	var endDate sql.NullTime
	if err := f.db.QueryRowContext(f.ctx, `SELECT status,end_date FROM suspensions WHERE discipline_id=?`, disciplineID).Scan(&suspensionStatus, &endDate); err != nil {
		t.Fatalf("read suspension lifecycle: %v", err)
	}
	if suspensionStatus != models.SuspensionStatusServed {
		t.Errorf("suspension status %q after overturned discipline, want %q", suspensionStatus, models.SuspensionStatusServed)
	}
	if !endDate.Valid {
		t.Errorf("suspension end_date is NULL after overturn; suspension lifecycle was not closed")
	}
}
