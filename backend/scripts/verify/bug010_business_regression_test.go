package verify

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/goxm2/sports-league/internal/database"
	"github.com/goxm2/sports-league/internal/handler"
	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/errorsx"
	"github.com/goxm2/sports-league/internal/repository"
	"github.com/goxm2/sports-league/internal/service"
)

func TestBug010_BusinessRegression(t *testing.T) {
	db := openBug010DB(t)
	repo := repository.NewSeasonRepo(repository.New(db))
	svc := service.NewSeasonService(&service.Deps{DB: db, Seasons: repo})
	h := handler.NewSeasonHandler(svc)

	unique := time.Now().UnixNano()
	creatorID := insertBug010(t, db,
		`INSERT INTO users (username,email,password_hash,full_name,status) VALUES (?,?,?,?,?)`,
		fmt.Sprintf("bug010-%d", unique), fmt.Sprintf("bug010-%d@league.test", unique), "unused", "BUG-010", 1)
	missingID := insertBug010Season(t, db, creatorID, "missing")
	normalID := insertBug010Season(t, db, creatorID, "normal")
	ruleID := insertBug010(t, db,
		`INSERT INTO scoring_rules (season_id,version,win_points,draw_points,loss_points,tiebreakers,is_active,created_by) VALUES (?,?,?,?,?,?,?,?)`,
		normalID, 1, 3, 1, 0, "points,goal_diff,goals_for", 1, creatorID)
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM scoring_rules WHERE season_id IN (?,?)`, missingID, normalID)
		_, _ = db.Exec(`DELETE FROM scoring_rule_versions WHERE season_id IN (?,?)`, missingID, normalID)
		_, _ = db.Exec(`DELETE FROM seasons WHERE id IN (?,?)`, missingID, normalID)
		_, _ = db.Exec(`DELETE FROM users WHERE id=?`, creatorID)
		_ = db.Close()
	})

	rule, err := svc.GetActiveScoringRule(context.Background(), normalID)
	if err != nil {
		t.Fatalf("normal active-rule lookup failed: %v", err)
	}
	if rule == nil || rule.ID != ruleID || rule.SeasonID != normalID || rule.Version != 1 || !rule.IsActive {
		t.Errorf("normal active rule = %#v, want id=%d season=%d version=1 active=true", rule, ruleID, normalID)
	} else if rule.WinPoints != 3 || rule.DrawPoints != 1 || rule.LossPoints != 0 {
		t.Errorf("normal active-rule points = %d/%d/%d, want 3/1/0", rule.WinPoints, rule.DrawPoints, rule.LossPoints)
	}

	before := countBug010(t, db, `SELECT COUNT(*) FROM scoring_rules WHERE season_id IN (?,?)`, missingID, normalID)

	rule, err = svc.GetActiveScoringRule(context.Background(), missingID)
	if err == nil {
		t.Errorf("season without an active scoring rule returned success: %#v", rule)
	} else if !errorsx.IsNotFound(err) {
		t.Errorf("missing active-rule error = %T %v, want not-found", err, err)
	}
	if rule != nil {
		t.Errorf("season without an active scoring rule returned fabricated object: %#v", rule)
	}

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	rule, err = svc.GetActiveScoringRule(cancelled, normalID)
	if err == nil {
		t.Errorf("cancelled active-rule query returned success: %#v", rule)
	} else if errorsx.IsNotFound(err) {
		t.Errorf("cancelled active-rule query was misclassified as not found: %v", err)
	}
	if rule != nil {
		t.Errorf("cancelled active-rule query returned fabricated object: %#v", rule)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/seasons/%d/rules", missingID), nil)
	req.SetPathValue("id", fmt.Sprint(missingID))
	recorder := httptest.NewRecorder()
	h.GetActiveRule(recorder, req)
	if recorder.Code != http.StatusNotFound {
		t.Errorf("missing active-rule HTTP status = %d, want 404; body=%s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		Code int             `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode missing active-rule response: %v", err)
	}
	if response.Code != 40400 {
		t.Errorf("missing active-rule business code = %d, want 40400", response.Code)
	}
	if len(response.Data) != 0 && string(response.Data) != "null" {
		t.Errorf("missing active-rule response exposed fabricated data: %s", response.Data)
	}

	after := countBug010(t, db, `SELECT COUNT(*) FROM scoring_rules WHERE season_id IN (?,?)`, missingID, normalID)
	if after != before {
		t.Errorf("active-rule reads changed persisted rule count from %d to %d", before, after)
	}
}

func openBug010DB(t *testing.T) *sql.DB {
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
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		t.Fatalf("connect MySQL test database: %v", err)
	}
	if err := database.Migrate(ctx, db); err != nil {
		_ = db.Close()
		t.Fatalf("migrate MySQL test database: %v", err)
	}
	return db
}

func insertBug010Season(t *testing.T, db *sql.DB, creatorID int64, suffix string) int64 {
	t.Helper()
	return insertBug010(t, db,
		`INSERT INTO seasons (code,name,sport,format,status,current_rule_version,created_by) VALUES (?,?,?,?,?,?,?)`,
		fmt.Sprintf("B010-%s-%d", suffix, time.Now().UnixNano()), "BUG-010 "+suffix,
		"football", "round_robin", models.SeasonStatusDraft, 1, creatorID)
}

func insertBug010(t *testing.T, db *sql.DB, query string, args ...any) int64 {
	t.Helper()
	result, err := db.Exec(query, args...)
	if err != nil {
		t.Fatalf("prepare BUG-010 fixture: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("read BUG-010 fixture id: %v", err)
	}
	return id
}

func countBug010(t *testing.T, db *sql.DB, query string, args ...any) int64 {
	t.Helper()
	var count int64
	if err := db.QueryRow(query, args...).Scan(&count); err != nil {
		t.Fatalf("count BUG-010 rows: %v", err)
	}
	return count
}
