package service

import (
	"database/sql"

	"github.com/goxm2/sports-league/internal/pkg/errorsx"
)

// rollback rolls back a transaction and returns the triggering error
// wrapped in an internal AppError.
func rollback(tx *sql.Tx, err error) error {
	_ = tx.Rollback()
	if err == nil {
		return errorsx.Internal("transaction failed")
	}
	if ae, ok := err.(*errorsx.AppError); ok {
		return ae
	}
	return errorsx.Wrap(500, 50000, "database error", err)
}
