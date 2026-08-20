package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
	"time"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// Migrate reads embedded .sql files in order and applies them idempotently.
// It tracks applied migrations in a `schema_migrations` table.
func Migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	applied, err := appliedMigrations(ctx, db)
	if err != nil {
		return err
	}

	entries, err := fs.ReadDir(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		if applied[name] {
			continue
		}
		raw, err := migrationFS.ReadFile(path.Join("migrations", name))
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		stmts := splitSQL(string(raw))
		for i, stmt := range stmts {
			trimmed := strings.TrimSpace(stmt)
			if trimmed == "" {
				continue
			}
			if _, err := db.ExecContext(ctx, trimmed); err != nil {
				return fmt.Errorf("apply %s stmt#%d: %w", name, i, err)
			}
		}
		if _, err := db.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES (?)", name); err != nil {
			return fmt.Errorf("record %s: %w", name, err)
		}
	}
	return nil
}

func appliedMigrations(ctx context.Context, db *sql.DB) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("list migrations: %w", err)
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out[v] = true
	}
	return out, rows.Err()
}

// splitSQL splits a multi-statement script. Comments (lines starting with --)
// are stripped. Statements are separated by semicolons at end-of-line.
func splitSQL(src string) []string {
	lines := strings.Split(src, "\n")
	var buf strings.Builder
	var out []string
	for _, ln := range lines {
		trim := strings.TrimSpace(ln)
		if strings.HasPrefix(trim, "--") {
			continue
		}
		buf.WriteString(ln)
		buf.WriteString("\n")
		if strings.HasSuffix(trim, ";") {
			out = append(out, buf.String())
			buf.Reset()
		}
	}
	rest := strings.TrimSpace(buf.String())
	if rest != "" {
		out = append(out, rest)
	}
	return out
}

var _ = time.Second
