package service

import (
	"context"
	"database/sql"

	"github.com/goxm2/sports-league/internal/auth"
	"github.com/goxm2/sports-league/internal/config"
	"github.com/goxm2/sports-league/internal/repository"
)

// Deps bundles all repositories and shared dependencies for the service layer.
type Deps struct {
	Cfg       *config.Config
	DB        *sql.DB
	JWT       *auth.Manager
	Users     *repository.UserRepo
	Tokens    *repository.TokenRepo
	Roles     *repository.RoleRepo
	Seasons   *repository.SeasonRepo
	Teams     *repository.TeamRepo
	Venues    *repository.VenueRepo
	Schedules *repository.ScheduleRepo
	Matches   *repository.MatchRepo
	Disciplines *repository.DisciplineRepo
	Standings *repository.StandingRepo
	Audit     *repository.AuditRepo
	Reports   *repository.ReportRepo
}

// NewDeps wires repositories into a Deps bundle.
func NewDeps(cfg *config.Config, db *sql.DB, jwtMgr *auth.Manager) *Deps {
	rdb := repository.New(db)
	return &Deps{
		Cfg:         cfg,
		DB:          db,
		JWT:         jwtMgr,
		Users:       repository.NewUserRepo(rdb),
		Tokens:      repository.NewTokenRepo(rdb),
		Roles:       repository.NewRoleRepo(rdb),
		Seasons:     repository.NewSeasonRepo(rdb),
		Teams:       repository.NewTeamRepo(rdb),
		Venues:      repository.NewVenueRepo(rdb),
		Schedules:   repository.NewScheduleRepo(rdb),
		Matches:     repository.NewMatchRepo(rdb),
		Disciplines: repository.NewDisciplineRepo(rdb),
		Standings:   repository.NewStandingRepo(rdb),
		Audit:       repository.NewAuditRepo(rdb),
		Reports:     repository.NewReportRepo(rdb),
	}
}

// ctxKey is an unexported type for context values.
type ctxKey int

const (
	ctxKeyUser ctxKey = iota
)

// CtxWithUser stores the authenticated user id/username in context.
type CtxUser struct {
	UserID   int64
	Username string
	Roles    []string
}

func withUser(ctx context.Context, u CtxUser) context.Context {
	return context.WithValue(ctx, ctxKeyUser, u)
}

// UserFromCtx retrieves the authenticated user from context, if present.
func UserFromCtx(ctx context.Context) (CtxUser, bool) {
	u, ok := ctx.Value(ctxKeyUser).(CtxUser)
	return u, ok
}
