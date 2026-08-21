package verify

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/goxm2/sports-league/internal/auth"
	"github.com/goxm2/sports-league/internal/config"
	"github.com/goxm2/sports-league/internal/database"
	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/service"
)

type bug018Fixture struct {
	t         *testing.T
	ctx       context.Context
	db        *sql.DB
	schedules *service.ScheduleService
}

func newBug018Fixture(t *testing.T) *bug018Fixture {
	t.Helper()
	dsn := os.Getenv("GSL_TEST_MYSQL_DSN")
	if dsn == "" {
		dsn = "league_user:league_pass@tcp(127.0.0.1:13307)/league_db?charset=utf8mb4&parseTime=true&loc=Local&multiStatements=true"
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open MySQL test database: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("connect MySQL test database: %v", err)
	}
	if err := database.Migrate(ctx, db); err != nil {
		t.Fatalf("migrate MySQL test database: %v", err)
	}
	resetBug018Database(t, db)
	t.Cleanup(func() { _ = db.Close() })
	cfg := &config.Config{JWT: config.JWTConfig{
		Secret:          "bug018-verification-secret-at-least-thirty-two-bytes",
		Issuer:          "go-sports-league-bug018-verification",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
	}}
	deps := service.NewDeps(cfg, db, auth.NewManager(cfg.JWT))
	return &bug018Fixture{t: t, ctx: context.Background(), db: db, schedules: service.NewScheduleService(deps)}
}

func resetBug018Database(t *testing.T, db *sql.DB) {
	t.Helper()
	tables := []string{
		"audit_logs", "standings_snapshots", "standings", "appeals", "suspensions", "disciplines",
		"player_match_stats", "match_events", "matches", "schedule_conflicts", "schedules", "venue_availability",
		"venues", "transfers", "players", "team_registrations", "teams", "scoring_rules", "seasons",
		"role_permissions", "user_roles", "permissions", "roles", "refresh_tokens", "users", "schema_migrations",
	}
	if _, err := db.Exec("SET FOREIGN_KEY_CHECKS=0"); err != nil {
		t.Fatalf("disable foreign key checks: %v", err)
	}
	for _, table := range tables {
		if _, err := db.Exec("TRUNCATE TABLE " + table); err != nil {
			t.Fatalf("truncate %s: %v", table, err)
		}
	}
	if _, err := db.Exec("SET FOREIGN_KEY_CHECKS=1"); err != nil {
		t.Fatalf("enable foreign key checks: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := database.Migrate(ctx, db); err != nil {
		t.Fatalf("reseed migrations: %v", err)
	}
}

func (f *bug018Fixture) exec(query string, args ...any) sql.Result {
	f.t.Helper()
	result, err := f.db.ExecContext(f.ctx, query, args...)
	if err != nil {
		f.t.Fatalf("exec %q: %v", query, err)
	}
	return result
}

func (f *bug018Fixture) insertID(query string, args ...any) int64 {
	f.t.Helper()
	result := f.exec(query, args...)
	id, err := result.LastInsertId()
	if err != nil {
		f.t.Fatalf("last insert id: %v", err)
	}
	return id
}

func (f *bug018Fixture) seedSchedule(code string) (int64, int64) {
	f.t.Helper()
	userID := f.insertID(
		`INSERT INTO users (username,email,password_hash,full_name,status) VALUES (?,?,?,?,?)`,
		"creator-"+code,
		"creator-"+code+"@league.test",
		"unused",
		"Creator "+code,
		1,
	)
	seasonID := f.insertID(
		`INSERT INTO seasons (code,name,sport,format,status,current_rule_version,created_by) VALUES (?,?,?,?,?,?,?)`,
		code,
		"Season "+code,
		"football",
		"round_robin",
		models.SeasonStatusOpen,
		1,
		userID,
	)
	homeID := f.insertID(
		`INSERT INTO teams (season_id,name,code,status,created_by) VALUES (?,?,?,?,?)`,
		seasonID,
		"Home "+code,
		"H"+code,
		"approved",
		userID,
	)
	awayID := f.insertID(
		`INSERT INTO teams (season_id,name,code,status,created_by) VALUES (?,?,?,?,?)`,
		seasonID,
		"Away "+code,
		"A"+code,
		"approved",
		userID,
	)
	venueID := f.insertID(
		`INSERT INTO venues (code,name,address,capacity,sport,status) VALUES (?,?,?,?,?,?)`,
		fmt.Sprintf("BUG018-%d", time.Now().UnixNano()),
		"Venue "+code,
		"verification",
		100,
		"football",
		models.VenueStatusAvailable,
	)
	scheduleID := f.insertID(
		`INSERT INTO schedules (season_id,round,home_team_id,away_team_id,venue_id,match_date,start_time,status) VALUES (?,?,?,?,?,?,?,?)`,
		seasonID,
		1,
		homeID,
		awayID,
		venueID,
		"2026-09-01",
		"19:00",
		models.ScheduleStatusScheduled,
	)
	return seasonID, scheduleID
}

func TestBug018_BusinessRegression(t *testing.T) {
	f := newBug018Fixture(t)
	targetSeasonID, targetScheduleID := f.seedSchedule("bug018-target")
	otherSeasonID, otherScheduleID := f.seedSchedule("bug018-other")
	f.exec(
		`INSERT INTO schedule_conflicts (schedule_id,conflict_type,description) VALUES (?,?,?)`,
		targetScheduleID,
		models.ConflictTypeVenue,
		"venue already booked",
	)
	f.exec(
		`INSERT INTO schedule_conflicts (schedule_id,conflict_type,description) VALUES (?,?,NULL)`,
		targetScheduleID,
		models.ConflictTypeTeam,
	)
	f.exec(
		`INSERT INTO schedule_conflicts (schedule_id,conflict_type,description) VALUES (?,?,NULL)`,
		targetScheduleID,
		models.ConflictTypeRound,
	)
	f.exec(
		`INSERT INTO schedule_conflicts (schedule_id,conflict_type,description) VALUES (?,?,NULL)`,
		otherScheduleID,
		models.ConflictTypeWindow,
	)

	conflicts, err := f.schedules.ListConflicts(f.ctx, targetSeasonID)
	if err != nil {
		t.Fatalf("list target-season conflicts: %v", err)
	}
	if len(conflicts) != 3 {
		t.Errorf("visible target-season conflict count = %d, want 3; conflicts=%#v", len(conflicts), conflicts)
	}
	types := make(map[string]models.ScheduleConflict, len(conflicts))
	for _, conflict := range conflicts {
		if conflict.ScheduleID != targetScheduleID {
			t.Errorf("target-season list leaked schedule %d, want only %d", conflict.ScheduleID, targetScheduleID)
		}
		types[conflict.ConflictType] = conflict
	}
	if conflict, ok := types[models.ConflictTypeVenue]; !ok {
		t.Errorf("described venue conflict missing; conflicts=%#v", conflicts)
	} else if conflict.Description != "venue already booked" {
		t.Errorf("venue conflict description = %q, want %q", conflict.Description, "venue already booked")
	}
	for _, conflictType := range []string{models.ConflictTypeTeam, models.ConflictTypeRound} {
		conflict, ok := types[conflictType]
		if !ok {
			t.Errorf("NULL-description %s conflict missing; conflicts=%#v", conflictType, conflicts)
			continue
		}
		if conflict.Description != "" {
			t.Errorf("NULL-description %s conflict hydrated as %q, want empty string", conflictType, conflict.Description)
		}
	}
	if _, ok := types[models.ConflictTypeWindow]; ok {
		t.Errorf("other-season window conflict leaked into target-season list")
	}

	otherConflicts, err := f.schedules.ListConflicts(f.ctx, otherSeasonID)
	if err != nil {
		t.Fatalf("list other-season conflicts: %v", err)
	}
	if len(otherConflicts) != 1 {
		t.Errorf("other-season conflict count = %d, want 1; conflicts=%#v", len(otherConflicts), otherConflicts)
	} else if otherConflicts[0].ConflictType != models.ConflictTypeWindow {
		t.Errorf("other-season conflict type = %q, want %q", otherConflicts[0].ConflictType, models.ConflictTypeWindow)
	}

	var storedCount int
	if err := f.db.QueryRowContext(
		f.ctx,
		`SELECT COUNT(*) FROM schedule_conflicts WHERE schedule_id=?`,
		targetScheduleID,
	).Scan(&storedCount); err != nil {
		t.Fatalf("count stored target conflicts: %v", err)
	}
	if storedCount != 3 {
		t.Errorf("stored target conflict count = %d, want 3", storedCount)
	}
	var nullDescriptionCount int
	if err := f.db.QueryRowContext(
		f.ctx,
		`SELECT COUNT(*) FROM schedule_conflicts WHERE schedule_id=? AND description IS NULL`,
		targetScheduleID,
	).Scan(&nullDescriptionCount); err != nil {
		t.Fatalf("count NULL-description target conflicts: %v", err)
	}
	if nullDescriptionCount != 2 {
		t.Errorf("stored NULL-description conflict count = %d, want 2", nullDescriptionCount)
	}
}
