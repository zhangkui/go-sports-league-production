package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/goxm2/sports-league/internal/auth"
	"github.com/goxm2/sports-league/internal/config"
	"github.com/goxm2/sports-league/internal/database"
	"github.com/goxm2/sports-league/internal/pkg/logx"
	"github.com/goxm2/sports-league/internal/router"
	"github.com/goxm2/sports-league/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		// logger not yet initialised; print to stderr
		panic(err)
	}
	logx.Init(cfg.Logging.Level, cfg.Logging.Format)
	defer logx.Sync()
	logx.Info("starting go-sports-league", "port", cfg.Server.Port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// database
	db, err := database.NewMySQL(cfg.MySQL)
	if err != nil {
		logx.Fatal("mysql connect failed", "err", err)
	}
	defer db.Close()

	if err := database.Migrate(context.Background(), db); err != nil {
		logx.Fatal("migration failed", "err", err)
	}
	logx.Info("database migrations applied")

	// redis
	rdb, err := database.NewRedis(cfg.Redis)
	if err != nil {
		logx.Warn("redis unavailable; rate limiting & idempotency will degrade to fail-open", "err", err)
		rdb = nil
	}
	if rdb != nil {
		defer rdb.Close()
	}

	// services
	jwtMgr := auth.NewManager(cfg.JWT)
	deps := service.NewDeps(cfg, db, jwtMgr)

	authSvc := service.NewAuthService(deps)
	userSvc := service.NewUserService(deps)
	roleSvc := service.NewRoleService(deps)
	seasonSvc := service.NewSeasonService(deps)
	teamSvc := service.NewTeamService(deps)
	venueSvc := service.NewVenueService(deps)
	schedSvc := service.NewScheduleService(deps)
	matchSvc := service.NewMatchService(deps)
	discSvc := service.NewDisciplineService(deps)
	standSvc := service.NewStandingService(deps)
	repSvc := service.NewReportService(deps)
	bootSvc := service.NewBootstrapService(deps)

	// bootstrap default admin
	if err := bootSvc.EnsureAdmin(ctx); err != nil {
		logx.Warn("bootstrap admin failed", "err", err)
	}
	// background token purger
	go bootSvc.PurgeTokens(context.Background(), 1*time.Hour)

	routerDeps := &router.Deps{
		Cfg: cfg, RDB: rdb,
		Auth: authSvc, Users: userSvc, Roles: roleSvc, Bootstrap: bootSvc,
		Seasons: seasonSvc, Teams: teamSvc, Venues: venueSvc, Schedules: schedSvc,
		Matches: matchSvc, Disciplines: discSvc, Standings: standSvc, Reports: repSvc,
	}
	handler := router.New(routerDeps)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logx.Fatal("server error", "err", err)
		}
	}()
	logx.Info("server listening", "addr", srv.Addr)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	logx.Info("shutdown signal received, draining...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logx.Error("graceful shutdown failed", "err", err)
	}
	cancel()
	logx.Info("server stopped")
}
