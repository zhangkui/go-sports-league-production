package verify

import (
	"context"
	"database/sql"
	"errors"
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

type bug019Fixture struct {
	t         *testing.T
	ctx       context.Context
	db        *sql.DB
	schedules *service.ScheduleService
}

func newBug019Fixture(t *testing.T) *bug019Fixture {
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
	resetBug019Database(t, db)
	t.Cleanup(func() { _ = db.Close() })
	cfg := &config.Config{JWT: config.JWTConfig{
		Secret:          "bug019-verification-secret-at-least-thirty-two-bytes",
		Issuer:          "go-sports-league-bug019-verification",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
	}}
	deps := service.NewDeps(cfg, db, auth.NewManager(cfg.JWT))
	return &bug019Fixture{t: t, ctx: context.Background(), db: db, schedules: service.NewScheduleService(deps)}
}

func resetBug019Database(t *testing.T, db *sql.DB) {
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

func (f *bug019Fixture) exec(query string, args ...any) sql.Result {
	f.t.Helper()
	result, err := f.db.ExecContext(f.ctx, query, args...)
	if err != nil {
		f.t.Fatalf("exec %q: %v", query, err)
	}
	return result
}

func (f *bug019Fixture) insertID(query string, args ...any) int64 {
	f.t.Helper()
	result := f.exec(query, args...)
	id, err := result.LastInsertId()
	if err != nil {
		f.t.Fatalf("last insert id: %v", err)
	}
	return id
}

func (f *bug019Fixture) seedSchedule() (int64, int64) {
	f.t.Helper()
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	userID := f.insertID(
		`INSERT INTO users (username,email,password_hash,full_name,status) VALUES (?,?,?,?,?)`,
		"bug019-creator-"+suffix,
		"bug019-"+suffix+"@league.test",
		"unused",
		"BUG-019 Creator",
		1,
	)
	seasonID := f.insertID(
		`INSERT INTO seasons (code,name,sport,format,status,current_rule_version,created_by) VALUES (?,?,?,?,?,?,?)`,
		"BUG019-"+suffix,
		"BUG-019 Season",
		"football",
		"round_robin",
		models.SeasonStatusOpen,
		1,
		userID,
	)
	homeID := f.insertID(
		`INSERT INTO teams (season_id,name,code,status,created_by) VALUES (?,?,?,?,?)`,
		seasonID, "BUG-019 Home", "B019H-"+suffix, "approved", userID,
	)
	awayID := f.insertID(
		`INSERT INTO teams (season_id,name,code,status,created_by) VALUES (?,?,?,?,?)`,
		seasonID, "BUG-019 Away", "B019A-"+suffix, "approved", userID,
	)
	venueID := f.insertID(
		`INSERT INTO venues (code,name,address,capacity,sport,status) VALUES (?,?,?,?,?,?)`,
		"B019V1-"+suffix, "BUG-019 Original Venue", "verification", 100, "football", models.VenueStatusAvailable,
	)
	alternateVenueID := f.insertID(
		`INSERT INTO venues (code,name,address,capacity,sport,status) VALUES (?,?,?,?,?,?)`,
		"B019V2-"+suffix, "BUG-019 Alternate Venue", "verification", 100, "football", models.VenueStatusAvailable,
	)
	scheduleID := f.insertID(
		`INSERT INTO schedules (season_id,round,home_team_id,away_team_id,venue_id,match_date,start_time,status) VALUES (?,?,?,?,?,?,?,?)`,
		seasonID, 1, homeID, awayID, venueID, "2026-09-01", "19:00", models.ScheduleStatusScheduled,
	)
	return scheduleID, alternateVenueID
}

func TestBug019_BusinessRegression(t *testing.T) {
	f := newBug019Fixture(t)
	scheduleID, alternateVenueID := f.seedSchedule()

	var beforeVenueID int64
	var beforeDate, beforeTime, beforeStatus, beforeUpdatedAt string
	if err := f.db.QueryRowContext(
		f.ctx,
		`SELECT venue_id, DATE_FORMAT(match_date,'%Y-%m-%d'), TIME_FORMAT(start_time,'%H:%i:%s'), status, DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s.%f') FROM schedules WHERE id=?`,
		scheduleID,
	).Scan(&beforeVenueID, &beforeDate, &beforeTime, &beforeStatus, &beforeUpdatedAt); err != nil {
		t.Fatalf("read original schedule state: %v", err)
	}
	var beforeCount int
	if err := f.db.QueryRowContext(f.ctx, `SELECT COUNT(*) FROM schedules`).Scan(&beforeCount); err != nil {
		t.Fatalf("count schedules before cancelled update: %v", err)
	}

	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	matchDate := "2026-09-02"
	startTime := "20:30"
	status := models.ScheduleStatusPostponed
	updated, updateErr := f.schedules.Update(cancelledCtx, scheduleID, models.UpdateScheduleRequest{
		VenueID: &alternateVenueID, MatchDate: &matchDate, StartTime: &startTime, Status: &status,
	})
	if !errors.Is(updateErr, context.Canceled) {
		t.Errorf("cancelled schedule update error = %v, want context.Canceled; result=%#v", updateErr, updated)
	}
	if updated != nil {
		t.Errorf("cancelled schedule update returned success object: %#v", updated)
	}

	var afterVenueID int64
	var afterDate, afterTime, afterStatus, afterUpdatedAt string
	if err := f.db.QueryRowContext(
		f.ctx,
		`SELECT venue_id, DATE_FORMAT(match_date,'%Y-%m-%d'), TIME_FORMAT(start_time,'%H:%i:%s'), status, DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s.%f') FROM schedules WHERE id=?`,
		scheduleID,
	).Scan(&afterVenueID, &afterDate, &afterTime, &afterStatus, &afterUpdatedAt); err != nil {
		t.Fatalf("read schedule after cancelled update: %v", err)
	}
	if afterVenueID != beforeVenueID {
		t.Errorf("cancelled request persisted venue_id = %d, want %d", afterVenueID, beforeVenueID)
	}
	if afterDate != beforeDate {
		t.Errorf("cancelled request persisted match_date = %q, want %q", afterDate, beforeDate)
	}
	if afterTime != beforeTime {
		t.Errorf("cancelled request persisted start_time = %q, want %q", afterTime, beforeTime)
	}
	if afterStatus != beforeStatus {
		t.Errorf("cancelled request persisted schedule status %q, want %q", afterStatus, beforeStatus)
	}
	if afterUpdatedAt != beforeUpdatedAt {
		t.Errorf("cancelled request changed updated_at = %q, want %q", afterUpdatedAt, beforeUpdatedAt)
	}
	var afterCount int
	if err := f.db.QueryRowContext(f.ctx, `SELECT COUNT(*) FROM schedules`).Scan(&afterCount); err != nil {
		t.Fatalf("count schedules after cancelled update: %v", err)
	}
	if afterCount != beforeCount {
		t.Errorf("cancelled update changed schedule row count from %d to %d", beforeCount, afterCount)
	}
}
