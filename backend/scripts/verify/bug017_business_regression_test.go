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

type bug017Fixture struct {
	t         *testing.T
	ctx       context.Context
	db        *sql.DB
	schedules *service.ScheduleService
}

func newBug017Fixture(t *testing.T) *bug017Fixture {
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
	resetBug017Database(t, db)
	t.Cleanup(func() { _ = db.Close() })

	cfg := &config.Config{JWT: config.JWTConfig{
		Secret:          "bug017-verification-secret-at-least-thirty-two-bytes",
		Issuer:          "go-sports-league-bug017-verification",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
	}}
	deps := service.NewDeps(cfg, db, auth.NewManager(cfg.JWT))
	return &bug017Fixture{
		t:         t,
		ctx:       context.Background(),
		db:        db,
		schedules: service.NewScheduleService(deps),
	}
}

func resetBug017Database(t *testing.T, db *sql.DB) {
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

func (f *bug017Fixture) exec(query string, args ...any) sql.Result {
	f.t.Helper()
	result, err := f.db.ExecContext(f.ctx, query, args...)
	if err != nil {
		f.t.Fatalf("exec %q: %v", query, err)
	}
	return result
}

func (f *bug017Fixture) insertID(query string, args ...any) int64 {
	f.t.Helper()
	result := f.exec(query, args...)
	id, err := result.LastInsertId()
	if err != nil {
		f.t.Fatalf("last insert id: %v", err)
	}
	return id
}

func (f *bug017Fixture) seedSeason(code string) int64 {
	f.t.Helper()
	creatorID := f.insertID(
		`INSERT INTO users (username,email,password_hash,full_name,status) VALUES (?,?,?,?,?)`,
		"creator-"+code,
		"creator-"+code+"@league.test",
		"unused",
		"Creator "+code,
		1,
	)
	return f.insertID(
		`INSERT INTO seasons (code,name,sport,format,status,current_rule_version,created_by) VALUES (?,?,?,?,?,?,?)`,
		code,
		"Season "+code,
		"football",
		"round_robin",
		models.SeasonStatusOpen,
		1,
		creatorID,
	)
}

func (f *bug017Fixture) seedApprovedTeam(seasonID int64, code string) int64 {
	f.t.Helper()
	return f.insertID(
		`INSERT INTO teams (season_id,name,code,status,created_by) VALUES (?,?,?,?,?)`,
		seasonID,
		"Team "+code,
		code,
		"approved",
		1,
	)
}

func (f *bug017Fixture) seedVenue(name string) int64 {
	f.t.Helper()
	venueID := f.insertID(
		`INSERT INTO venues (code,name,address,capacity,sport,status) VALUES (?,?,?,?,?,?)`,
		fmt.Sprintf("BUG017-%d", time.Now().UnixNano()),
		name,
		"verification",
		100,
		"football",
		models.VenueStatusAvailable,
	)
	for weekday := 0; weekday < 7; weekday++ {
		f.exec(
			`INSERT INTO venue_availability (venue_id,weekday,start_time,end_time) VALUES (?,?,?,?)`,
			venueID,
			weekday,
			"00:00",
			"23:59",
		)
	}
	return venueID
}

func verifyBug017RoundRobin(t *testing.T, teamCount int, format string) {
	t.Helper()
	f := newBug017Fixture(t)
	seasonID := f.seedSeason(fmt.Sprintf("bug017-%d-%s", teamCount, format))
	teamIDs := make([]int64, 0, teamCount)
	for index := 0; index < teamCount; index++ {
		teamIDs = append(teamIDs, f.seedApprovedTeam(seasonID, fmt.Sprintf("T%02d", index+1)))
	}
	venueID := f.seedVenue(fmt.Sprintf("BUG-017 %d team venue", teamCount))
	legs := 1
	if format == "double_round" {
		legs = 2
	}
	expectedRounds := (teamCount - 1) * legs
	expectedMatches := teamCount * (teamCount - 1) / 2 * legs

	schedules, conflicts, err := f.schedules.Generate(f.ctx, models.GenerateScheduleRequest{
		SeasonID:    seasonID,
		VenueID:     venueID,
		StartDate:   "2026-09-01",
		GamesPerDay: 1,
		Format:      format,
	}, 1)
	if err != nil {
		t.Fatalf("generate %d-team %s schedule: %v", teamCount, format, err)
	}
	if len(conflicts) != 0 {
		t.Errorf("generated conflicts = %#v, want none", conflicts)
	}
	if len(schedules) != expectedMatches {
		t.Errorf("generated schedule count = %d, want %d", len(schedules), expectedMatches)
	}

	allowedTeams := make(map[int64]struct{}, len(teamIDs))
	for _, teamID := range teamIDs {
		allowedTeams[teamID] = struct{}{}
	}
	pairCounts := make(map[[2]int64]int, expectedMatches/legs)
	roundCounts := make(map[int]int, expectedRounds)
	teamAppearances := make(map[int64]int, teamCount)
	for _, schedule := range schedules {
		if _, ok := allowedTeams[schedule.HomeTeamID]; !ok {
			t.Errorf("unexpected home team %d", schedule.HomeTeamID)
		}
		if _, ok := allowedTeams[schedule.AwayTeamID]; !ok {
			t.Errorf("unexpected away team %d", schedule.AwayTeamID)
		}
		if schedule.HomeTeamID == schedule.AwayTeamID {
			t.Errorf("self pairing for team %d", schedule.HomeTeamID)
		}
		pair := [2]int64{schedule.HomeTeamID, schedule.AwayTeamID}
		if pair[0] > pair[1] {
			pair[0], pair[1] = pair[1], pair[0]
		}
		pairCounts[pair]++
		roundCounts[schedule.Round]++
		teamAppearances[schedule.HomeTeamID]++
		teamAppearances[schedule.AwayTeamID]++
	}

	expectedUniquePairs := teamCount * (teamCount - 1) / 2
	if len(pairCounts) != expectedUniquePairs {
		t.Errorf("unique team pair count = %d, want %d; schedules=%#v", len(pairCounts), expectedUniquePairs, schedules)
	}
	for pair, count := range pairCounts {
		if count != legs {
			t.Errorf("pair %v occurrence count = %d, want %d", pair, count, legs)
		}
	}
	if len(roundCounts) != expectedRounds {
		t.Errorf("round count = %d, want %d; round_counts=%v", len(roundCounts), expectedRounds, roundCounts)
	}
	for round := 1; round <= expectedRounds; round++ {
		if roundCounts[round] != teamCount/2 {
			t.Errorf("round %d match count = %d, want %d", round, roundCounts[round], teamCount/2)
		}
	}
	for _, teamID := range teamIDs {
		expectedAppearances := (teamCount - 1) * legs
		if teamAppearances[teamID] != expectedAppearances {
			t.Errorf("team %d appearance count = %d, want %d", teamID, teamAppearances[teamID], expectedAppearances)
		}
	}

	var persistedMatches int
	if err := f.db.QueryRowContext(f.ctx, `SELECT COUNT(*) FROM schedules WHERE season_id=?`, seasonID).Scan(&persistedMatches); err != nil {
		t.Fatalf("count persisted schedules: %v", err)
	}
	if persistedMatches != expectedMatches {
		t.Errorf("persisted schedule count = %d, want %d", persistedMatches, expectedMatches)
	}
	var persistedRounds int
	if err := f.db.QueryRowContext(f.ctx, `SELECT COUNT(DISTINCT round) FROM schedules WHERE season_id=?`, seasonID).Scan(&persistedRounds); err != nil {
		t.Fatalf("count persisted rounds: %v", err)
	}
	if persistedRounds != expectedRounds {
		t.Errorf("persisted round count = %d, want %d", persistedRounds, expectedRounds)
	}
}

func TestBug017_BusinessRegression(t *testing.T) {
	t.Run("two-team legal schedule remains available", func(t *testing.T) {
		verifyBug017RoundRobin(t, 2, "round_robin")
	})
	t.Run("four-team single round robin keeps six independent pairings", func(t *testing.T) {
		verifyBug017RoundRobin(t, 4, "round_robin")
	})
	t.Run("six-team single round robin keeps fifteen independent pairings", func(t *testing.T) {
		verifyBug017RoundRobin(t, 6, "round_robin")
	})
	t.Run("four-team double round robin keeps both legs independent", func(t *testing.T) {
		verifyBug017RoundRobin(t, 4, "double_round")
	})
}
