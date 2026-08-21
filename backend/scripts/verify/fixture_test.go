package verify

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"

	"github.com/goxm2/sports-league/internal/auth"
	"github.com/goxm2/sports-league/internal/config"
	"github.com/goxm2/sports-league/internal/database"
	"github.com/goxm2/sports-league/internal/service"
)

type fixture struct {
	t           *testing.T
	ctx         context.Context
	db          *sql.DB
	rdb         *redis.Client
	cfg         *config.Config
	deps        *service.Deps
	auth        *service.AuthService
	users       *service.UserService
	roles       *service.RoleService
	seasons     *service.SeasonService
	teams       *service.TeamService
	venues      *service.VenueService
	schedules   *service.ScheduleService
	matches     *service.MatchService
	disciplines *service.DisciplineService
	reports     *service.ReportService
	bootstrap   *service.BootstrapService
}

func newFixture(t *testing.T) *fixture {
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
	resetDatabase(t, db)
	t.Cleanup(func() { _ = db.Close() })

	redisAddr := os.Getenv("GSL_TEST_REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6380"
	}
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatalf("connect Redis test database: %v", err)
	}
	if err := rdb.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("flush Redis test database: %v", err)
	}
	t.Cleanup(func() { _ = rdb.Close() })

	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:          "verification-secret-with-at-least-thirty-two-bytes",
			Issuer:          "go-sports-league-verification",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 24 * time.Hour,
		},
		Admin:     config.AdminConfig{Username: "admin", Password: "Admin123!", Email: "admin@league.test"},
		Server:    config.ServerConfig{AllowOrigins: []string{"*"}},
		RateLimit: config.RateLimitConfig{LoginPerMinuteIP: 1000, LoginPerMinuteUser: 1000, GlobalPerMinute: 1000},
	}
	deps := service.NewDeps(cfg, db, auth.NewManager(cfg.JWT))
	return &fixture{
		t: t, ctx: context.Background(), db: db, rdb: rdb, cfg: cfg, deps: deps,
		auth: service.NewAuthService(deps), users: service.NewUserService(deps), roles: service.NewRoleService(deps),
		seasons: service.NewSeasonService(deps), teams: service.NewTeamService(deps), venues: service.NewVenueService(deps),
		schedules: service.NewScheduleService(deps), matches: service.NewMatchService(deps),
		disciplines: service.NewDisciplineService(deps), reports: service.NewReportService(deps),
		bootstrap: service.NewBootstrapService(deps),
	}
}

func resetDatabase(t *testing.T, db *sql.DB) {
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

func (f *fixture) exec(query string, args ...any) sql.Result {
	f.t.Helper()
	result, err := f.db.ExecContext(f.ctx, query, args...)
	if err != nil {
		f.t.Fatalf("exec %q: %v", query, err)
	}
	return result
}

func (f *fixture) insertID(query string, args ...any) int64 {
	f.t.Helper()
	result := f.exec(query, args...)
	id, err := result.LastInsertId()
	if err != nil {
		f.t.Fatalf("last insert id: %v", err)
	}
	return id
}

func (f *fixture) seedUser(username string, status int, passwordHash string) int64 {
	f.t.Helper()
	return f.insertID(`INSERT INTO users (username,email,password_hash,full_name,status) VALUES (?,?,?,?,?)`,
		username, username+"@league.test", passwordHash, username, status)
}

func (f *fixture) seedSeason(code, status string) int64 {
	f.t.Helper()
	creator := f.seedUser("creator-"+code, 1, "unused")
	return f.insertID(`INSERT INTO seasons (code,name,sport,format,status,current_rule_version,created_by) VALUES (?,?,?,?,?,?,?)`,
		code, "Season "+code, "football", "round_robin", status, 1, creator)
}

func (f *fixture) seedRule(seasonID int64) int64 {
	f.t.Helper()
	return f.insertID(`INSERT INTO scoring_rules (season_id,version,win_points,draw_points,loss_points,tiebreakers,is_active,created_by) VALUES (?,?,?,?,?,?,?,?)`,
		seasonID, 1, 3, 1, 0, `["points","goal_diff","goals_for"]`, 1, 1)
}

func (f *fixture) seedTeam(seasonID int64, code string) int64 {
	f.t.Helper()
	return f.insertID(`INSERT INTO teams (season_id,name,code,status,created_by) VALUES (?,?,?,?,?)`, seasonID, "Team "+code, code, "active", 1)
}

func (f *fixture) seedPlayer(seasonID, teamID int64, name string, number int) int64 {
	f.t.Helper()
	return f.insertID(`INSERT INTO players (season_id,team_id,name,number,status) VALUES (?,?,?,?,?)`, seasonID, teamID, name, number, "active")
}

func (f *fixture) seedVenue(name string) int64 {
	f.t.Helper()
	return f.insertID(`INSERT INTO venues (code,name,address,capacity,sport,status) VALUES (?,?,?,?,?,?)`,
		fmt.Sprintf("V-%d", time.Now().UnixNano()), name, "verification", 100, "football", "available")
}

func (f *fixture) seedMatch(seasonID, homeTeamID, awayTeamID int64, homeScore, awayScore int, confirmStatus string) int64 {
	f.t.Helper()
	venueID := f.seedVenue(fmt.Sprintf("venue-%d-%d", homeTeamID, awayTeamID))
	scheduleID := f.insertID(`INSERT INTO schedules (season_id,round,home_team_id,away_team_id,venue_id,match_date,start_time,status) VALUES (?,?,?,?,?,?,?,?)`,
		seasonID, 1, homeTeamID, awayTeamID, venueID, time.Now().Add(24*time.Hour), "19:30", "published")
	return f.insertID(`INSERT INTO matches (schedule_id,season_id,home_team_id,away_team_id,venue_id,match_date,start_time,home_score,away_score,status,confirm_status) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		scheduleID, seasonID, homeTeamID, awayTeamID, venueID, time.Now(), "19:30", homeScore, awayScore, "completed", confirmStatus)
}

func (f *fixture) count(query string, args ...any) int64 {
	f.t.Helper()
	var count int64
	if err := f.db.QueryRowContext(f.ctx, query, args...).Scan(&count); err != nil {
		f.t.Fatalf("count %q: %v", query, err)
	}
	return count
}
