package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/goxm2/sports-league/internal/pkg/errorsx"
)

// DB is the shared *sql.DB handle, embedded by every repository.
type DB struct {
	*sql.DB
}

func New(db *sql.DB) DB { return DB{db} }

// NotFound wraps a sql.ErrNoRows into an AppError.
func NotFound(resource string, id any) error {
	return errorsx.NotFound(fmt.Sprintf("%s %v not found", resource, id))
}

// translateDup returns a Conflict AppError on a MySQL duplicate-key error.
func translateDup(err error, msg string) error {
	if err == nil {
		return nil
	}
	if isDuplicateKey(err) {
		return errorsx.Conflict(msg)
	}
	return errorsx.Wrap(500, 50000, "database error", err)
}

func isDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "Duplicate entry") || strings.Contains(s, "1062")
}

// rollback rolls back the transaction and returns the triggering error.
func rollback(tx *sql.Tx, err error) error {
	_ = tx.Rollback()
	return err
}
