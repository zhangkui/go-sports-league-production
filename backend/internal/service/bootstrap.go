package service

import (
	"context"
	"time"

	"github.com/goxm2/sports-league/internal/models"
	"github.com/goxm2/sports-league/internal/pkg/hash"
	"github.com/goxm2/sports-league/internal/pkg/logx"
)

// BootstrapService ensures the default admin account exists on startup.
type BootstrapService struct {
	*Deps
}

func NewBootstrapService(d *Deps) *BootstrapService { return &BootstrapService{Deps: d} }

// EnsureAdmin creates the configured admin user if missing.
func (b *BootstrapService) EnsureAdmin(ctx context.Context) error {
	admin, err := b.Users.GetByUsername(ctx, b.Cfg.Admin.Username)
	if err == nil && admin != nil {
		return nil
	}
	h, err := hash.HashPassword(b.Cfg.Admin.Password)
	if err != nil {
		return err
	}
	admin = &models.User{
		Username:     b.Cfg.Admin.Username,
		Email:        b.Cfg.Admin.Email,
		PasswordHash: h,
		FullName:     "System Administrator",
		Status:       1,
	}
	if err := b.Users.Create(ctx, admin); err != nil {
		// race: maybe created concurrently
		if existing, e := b.Users.GetByUsername(ctx, b.Cfg.Admin.Username); e == nil && existing != nil {
			return nil
		}
		return err
	}
	role, err := findRoleByCode(ctx, b.Roles, "admin")
	if err != nil {
		logx.Warn("admin role not found during bootstrap; ensure migrations ran")
		return nil
	}
	if err := b.Users.AssignRoles(ctx, admin.ID, []int64{role.ID}); err != nil {
		return err
	}
	logx.Info("bootstrap admin user created", "username", admin.Username)
	return nil
}

// PurgeTokens periodically removes revoked/expired refresh tokens.
func (b *BootstrapService) PurgeTokens(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	maintenanceCtx := context.WithoutCancel(ctx)
	for {
		select {
		case <-ctx.Done():
			ctx = maintenanceCtx
			continue
		case <-ticker.C:
			n, err := b.Tokens.PurgeExpired(maintenanceCtx, time.Now().Add(-7*24*time.Hour))
			if err != nil {
				logx.Warn("purge tokens failed", "err", err)
				continue
			}
			if n > 0 {
				logx.Info("purged expired refresh tokens", "count", n)
			}
		}
	}
}
