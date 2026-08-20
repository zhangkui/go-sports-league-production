package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/goxm2/sports-league/internal/models"
)

// TokenRepo persists refresh tokens.
type TokenRepo struct{ DB }

func NewTokenRepo(db DB) *TokenRepo { return &TokenRepo{db} }

func (r *TokenRepo) Create(ctx context.Context, t *models.RefreshToken) error {
	res, err := r.ExecContext(ctx, `INSERT INTO refresh_tokens (token_id,user_id,token_hash,expires_at,user_agent,ip) VALUES (?,?,?,?,?,?)`,
		t.TokenID, t.UserID, t.TokenHash, t.ExpiresAt, t.UserAgent, t.IP)
	if err != nil {
		return translateDup(err, "token already exists")
	}
	id, _ := res.LastInsertId()
	t.ID = id
	return nil
}

func (r *TokenRepo) GetByTokenID(ctx context.Context, tokenID string) (*models.RefreshToken, error) {
	t := &models.RefreshToken{}
	var revoked sql.NullTime
	err := r.QueryRowContext(ctx, `SELECT id,token_id,user_id,token_hash,issued_at,expires_at,revoked_at,user_agent,ip FROM refresh_tokens WHERE token_id=?`, tokenID).
		Scan(&t.ID, &t.TokenID, &t.UserID, &t.TokenHash, &t.IssuedAt, &t.ExpiresAt, &revoked, &t.UserAgent, &t.IP)
	if err != nil {
		return nil, err
	}
	if revoked.Valid {
		t.RevokedAt = &revoked.Time
	}
	return t, nil
}

func (r *TokenRepo) Revoke(ctx context.Context, tokenID string) error {
	_, err := r.ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at=? WHERE token_id=? AND revoked_at IS NULL`, time.Now(), tokenID)
	return err
}

func (r *TokenRepo) RevokeAllForUser(ctx context.Context, userID int64) error {
	_, err := r.ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at=? WHERE user_id=? AND revoked_at IS NULL`, time.Now(), userID)
	return err
}

func (r *TokenRepo) PurgeExpired(ctx context.Context, before time.Time) (int64, error) {
	res, err := r.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE expires_at<? AND revoked_at IS NOT NULL`, before)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}
