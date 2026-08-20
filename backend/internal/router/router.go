package router

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"

	"github.com/goxm2/sports-league/internal/config"
	"github.com/goxm2/sports-league/internal/handler"
	"github.com/goxm2/sports-league/internal/middleware"
	"github.com/goxm2/sports-league/internal/pkg/pagination"
	"github.com/goxm2/sports-league/internal/pkg/response"
	"github.com/goxm2/sports-league/internal/service"
)

// Deps bundles the services required to build the router.
type Deps struct {
	Cfg       *config.Config
	RDB       *redis.Client
	Auth      *service.AuthService
	Users     *service.UserService
	Roles     *service.RoleService
	Bootstrap *service.BootstrapService
	Seasons   *service.SeasonService
	Teams     *service.TeamService
	Venues    *service.VenueService
	Schedules *service.ScheduleService
	Matches   *service.MatchService
	Disciplines *service.DisciplineService
	Standings *service.StandingService
	Reports   *service.ReportService
}

// perm is a shorthand for RequirePermission.
func perm(code string) func(http.Handler) http.Handler { return middleware.RequirePermission(code) }
func permAny(codes ...string) func(http.Handler) http.Handler {
	return middleware.RequireAnyPermission(codes...)
}

// idem wraps a handler func with the idempotency middleware.
func idem(mw *middleware.IdempotencyMiddleware, h http.HandlerFunc) http.Handler {
	return mw.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("X-Idempotency-Route", r.Method+":"+r.URL.Path)
		h.ServeHTTP(w, r)
	}))
}

// authRoute returns the auth+rbac chain applied to a handler.
func authRoute(authMW *middleware.AuthMiddleware, h http.HandlerFunc, perms ...string) http.Handler {
	chain := http.Handler(h)
	chain = authMW.RequireAuth(chain)
	if len(perms) > 0 {
		chain = permAny(perms...)(chain)
	}
	return chain
}

// New builds the configured HTTP router with all middleware and routes.
func New(d *Deps) http.Handler {
	r := mux.NewRouter()
	pagination.SetMuxVarsHook(mux.Vars)

	cors := middleware.CORSMiddleware(d.Cfg.Server.AllowOrigins)
	recovery := middleware.RecoveryMiddleware
	logging := middleware.LoggingMiddleware
	requestID := middleware.RequestIDMiddleware

	authMW := middleware.NewAuthMiddleware(d.Auth)
	loginLimiter := middleware.LoginRateLimit(d.RDB, d.Cfg.RateLimit.LoginPerMinuteIP, d.Cfg.RateLimit.LoginPerMinuteUser)
	idemMW := middleware.NewIdempotencyMiddleware(d.RDB, 24*time.Hour)
	auditMW := middleware.NewAuditMiddleware(d.Reports)

	// health
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		response.Write(w, r, map[string]any{"status": "ok"})
	}).Methods(http.MethodGet)
	r.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		response.Write(w, r, map[string]any{"status": "ready"})
	}).Methods(http.MethodGet)

	authH := handler.NewAuthHandler(d.Auth)
	userH := handler.NewUserHandler(d.Users, d.Roles, d.Auth)
	seasonH := handler.NewSeasonHandler(d.Seasons)
	teamH := handler.NewTeamHandler(d.Teams)
	venueH := handler.NewVenueHandler(d.Venues)
	schedH := handler.NewScheduleHandler(d.Schedules)
	matchH := handler.NewMatchHandler(d.Matches)
	discH := handler.NewDisciplineHandler(d.Disciplines)
	standH := handler.NewStandingHandler(d.Standings)
	repH := handler.NewReportHandler(d.Reports)
	auditH := handler.NewAuditHandler(d.Reports)

	api := r.PathPrefix("/api/v1").Subrouter()

	// --- public auth ---
	api.Handle("/auth/register", idem(idemMW, authH.Register)).Methods(http.MethodPost)
	api.Handle("/auth/login", loginLimiter(idem(idemMW, authH.Login))).Methods(http.MethodPost)
	api.Handle("/auth/refresh", idem(idemMW, authH.Refresh)).Methods(http.MethodPost)
	api.Handle("/auth/logout", http.HandlerFunc(authH.Logout)).Methods(http.MethodPost)
	api.Handle("/me", authMW.RequireAuth(http.HandlerFunc(authH.Me))).Methods(http.MethodGet)
	api.Handle("/me/password", authMW.RequireAuth(idem(idemMW, authH.ChangePassword))).Methods(http.MethodPut)

	// --- users & roles (admin) ---
	users := api.PathPrefix("/users").Subrouter()
	users.Use(authMW.RequireAuth, perm("users:manage"))
	users.HandleFunc("", userH.List).Methods(http.MethodGet)
	users.HandleFunc("/{id}", userH.Get).Methods(http.MethodGet)
	users.Handle("/{id}", idem(idemMW, userH.Update)).Methods(http.MethodPut)
	users.HandleFunc("/{id}/enable", userH.Enable).Methods(http.MethodPut)
	users.Handle("/{id}/roles", idem(idemMW, userH.AssignRoles)).Methods(http.MethodPut)
	users.Handle("/{id}/password", idem(idemMW, userH.ResetPassword)).Methods(http.MethodPost)

	roles := api.PathPrefix("/roles").Subrouter()
	roles.Use(authMW.RequireAuth, perm("users:manage"))
	roles.HandleFunc("", userH.ListRoles).Methods(http.MethodGet)
	roles.Handle("", idem(idemMW, userH.CreateRole)).Methods(http.MethodPost)
	roles.HandleFunc("/{id}", userH.GetRole).Methods(http.MethodGet)
	roles.Handle("/{id}", idem(idemMW, userH.UpdateRole)).Methods(http.MethodPut)
	roles.HandleFunc("/{id}", userH.DeleteRole).Methods(http.MethodDelete)

	perms := api.PathPrefix("/permissions").Subrouter()
	perms.Use(authMW.RequireAuth, perm("users:manage"))
	perms.HandleFunc("", userH.ListPermissions).Methods(http.MethodGet)

	// --- seasons ---
	seasons := api.PathPrefix("/seasons").Subrouter()
	seasons.Use(authMW.RequireAuth)
	seasons.Handle("", dispatchGetPost(seasonH.List, permAny("seasons:manage", "seasons:view")(idem(idemMW, seasonH.Create)))).Methods(http.MethodGet, http.MethodPost)
	seasons.Handle("/{id}", dispatchGetPut(seasonH.Get, permAny("seasons:manage", "seasons:view")(http.HandlerFunc(seasonH.Update)))).Methods(http.MethodGet, http.MethodPut)
	seasons.Handle("/{id}/status", perm("seasons:manage")(http.HandlerFunc(seasonH.ChangeStatus))).Methods(http.MethodPut)
	seasons.Handle("/{id}/rules", dispatchGetPost(seasonH.GetActiveRule, perm("seasons:manage")(http.HandlerFunc(seasonH.SetScoringRule)))).Methods(http.MethodGet, http.MethodPost)
	seasons.HandleFunc("/{id}/rules/versions", seasonH.ListRuleVersions).Methods(http.MethodGet)

	// --- teams ---
	teams := api.PathPrefix("/teams").Subrouter()
	teams.Use(authMW.RequireAuth)
	teams.HandleFunc("", teamH.List).Methods(http.MethodGet)
	teams.Handle("", idem(idemMW, teamH.Create)).Methods(http.MethodPost)
	teams.HandleFunc("/{id}", teamH.Get).Methods(http.MethodGet)
	teams.Handle("/{id}", idem(idemMW, teamH.Update)).Methods(http.MethodPut)
	teams.Handle("/{id}/review", permAny("teams:approve", "seasons:manage")(idem(idemMW, teamH.ReviewRegistration))).Methods(http.MethodPut)
	regR := api.PathPrefix("/team-registrations").Subrouter()
	regR.Use(authMW.RequireAuth, permAny("teams:approve", "seasons:manage"))
	regR.HandleFunc("", teamH.ListRegistrations).Methods(http.MethodGet)

	// --- players ---
	players := api.PathPrefix("/players").Subrouter()
	players.Use(authMW.RequireAuth)
	players.HandleFunc("", teamH.ListPlayers).Methods(http.MethodGet)
	players.Handle("", idem(idemMW, teamH.CreatePlayer)).Methods(http.MethodPost)
	players.HandleFunc("/{id}", teamH.GetPlayer).Methods(http.MethodGet)
	players.Handle("/{id}", idem(idemMW, teamH.UpdatePlayer)).Methods(http.MethodPut)
	players.HandleFunc("/{id}/disciplines", discH.PlayerHistory).Methods(http.MethodGet)

	// --- transfers ---
	transfers := api.PathPrefix("/transfers").Subrouter()
	transfers.Use(authMW.RequireAuth)
	transfers.HandleFunc("", teamH.ListTransfers).Methods(http.MethodGet)
	transfers.Handle("", idem(idemMW, teamH.CreateTransfer)).Methods(http.MethodPost)
	transfers.Handle("/{id}/review", permAny("teams:approve", "seasons:manage")(idem(idemMW, teamH.ReviewTransfer))).Methods(http.MethodPut)

	// --- venues ---
	venues := api.PathPrefix("/venues").Subrouter()
	venues.Use(authMW.RequireAuth)
	venues.HandleFunc("", venueH.List).Methods(http.MethodGet)
	venues.Handle("", idem(idemMW, venueH.Create)).Methods(http.MethodPost)
	venues.HandleFunc("/{id}", venueH.Get).Methods(http.MethodGet)
	venues.Handle("/{id}", idem(idemMW, venueH.Update)).Methods(http.MethodPut)
	venues.HandleFunc("/{id}", venueH.Delete).Methods(http.MethodDelete)
	venues.HandleFunc("/{id}/availability", venueH.ListAvailability).Methods(http.MethodGet)
	venues.Handle("/{id}/availability", idem(idemMW, venueH.SetAvailability)).Methods(http.MethodPut)

	// --- schedules ---
	schedules := api.PathPrefix("/schedules").Subrouter()
	schedules.Use(authMW.RequireAuth)
	schedules.HandleFunc("", schedH.List).Methods(http.MethodGet)
	schedules.Handle("/generate", permAny("schedules:generate", "schedules:manage")(idem(idemMW, schedH.Generate))).Methods(http.MethodPost)
	schedules.HandleFunc("/conflicts", schedH.ListConflicts).Methods(http.MethodGet)
	schedules.HandleFunc("/{id}", schedH.Get).Methods(http.MethodGet)
	schedules.Handle("/{id}", idem(idemMW, schedH.Update)).Methods(http.MethodPut)

	// --- matches ---
	matches := api.PathPrefix("/matches").Subrouter()
	matches.Use(authMW.RequireAuth)
	matches.HandleFunc("", matchH.List).Methods(http.MethodGet)
	matches.HandleFunc("/{id}", matchH.Get).Methods(http.MethodGet)
	matches.Handle("/{id}", permAny("matches:record", "matches:manage")(idem(idemMW, matchH.Record))).Methods(http.MethodPut)
	matches.Handle("/{id}/status", permAny("matches:record", "matches:manage")(http.HandlerFunc(matchH.SetStatus))).Methods(http.MethodPut)
	matches.Handle("/{id}/confirm", permAny("matches:confirm", "matches:manage")(idem(idemMW, matchH.Confirm))).Methods(http.MethodPut)
	matches.HandleFunc("/{id}/dispute", matchH.Dispute).Methods(http.MethodPut)
	matches.HandleFunc("/{id}/events", matchH.ListEvents).Methods(http.MethodGet)
	matches.Handle("/{id}/events", permAny("matches:record", "matches:manage")(idem(idemMW, matchH.CreateEvent))).Methods(http.MethodPost)
	matches.Handle("/{id}/events/{event_id}", permAny("matches:record", "matches:manage")(idem(idemMW, matchH.UpdateEvent))).Methods(http.MethodPut)
	matches.Handle("/{id}/events/{event_id}", permAny("matches:record", "matches:manage")(http.HandlerFunc(matchH.DeleteEvent))).Methods(http.MethodDelete)
	matches.HandleFunc("/{id}/stats", matchH.ListPlayerStats).Methods(http.MethodGet)
	matches.Handle("/{id}/stats", permAny("matches:record", "matches:manage")(idem(idemMW, matchH.UpsertPlayerStat))).Methods(http.MethodPost)

	// --- disciplines ---
	disciplines := api.PathPrefix("/disciplines").Subrouter()
	disciplines.Use(authMW.RequireAuth)
	disciplines.HandleFunc("", discH.List).Methods(http.MethodGet)
	disciplines.Handle("", permAny("disciplines:manage")(idem(idemMW, discH.Create))).Methods(http.MethodPost)
	disciplines.HandleFunc("/{id}", discH.Get).Methods(http.MethodGet)
	disciplines.Handle("/{id}/overturn", permAny("disciplines:manage")(http.HandlerFunc(discH.Overturn))).Methods(http.MethodPut)

	// --- appeals ---
	appeals := api.PathPrefix("/appeals").Subrouter()
	appeals.Use(authMW.RequireAuth)
	appeals.HandleFunc("", discH.ListAppeals).Methods(http.MethodGet)
	appeals.Handle("", idem(idemMW, discH.CreateAppeal)).Methods(http.MethodPost)
	appeals.Handle("/{id}/review", permAny("appeals:manage")(idem(idemMW, discH.ReviewAppeal))).Methods(http.MethodPut)

	// --- standings & reports (view-only mostly) ---
	stand := api.PathPrefix("/standings").Subrouter()
	stand.Use(authMW.OptionalAuth)
	stand.HandleFunc("", standH.List).Methods(http.MethodGet)
	stand.HandleFunc("/snapshots", standH.ListSnapshots).Methods(http.MethodGet)
	stand.HandleFunc("/snapshots/{id}", standH.GetSnapshot).Methods(http.MethodGet)

	reports := api.PathPrefix("/reports").Subrouter()
	reports.Use(authMW.RequireAuth, perm("reports:view"))
	reports.HandleFunc("/season-summary", repH.SeasonSummary).Methods(http.MethodGet)
	reports.HandleFunc("/player-ranking", repH.PlayerRanking).Methods(http.MethodGet)

	// --- audit logs ---
	auditR := api.PathPrefix("/audit-logs").Subrouter()
	auditR.Use(authMW.RequireAuth, permAny("audit_logs:view"))
	auditR.HandleFunc("", auditH.List).Methods(http.MethodGet)
	auditR.HandleFunc("/{id}", auditH.Get).Methods(http.MethodGet)

	return recovery(cors(requestID(logging(auditMW.Wrap(r)))))
}

// dispatchGetPost routes GET to list and POST to create handler.
func dispatchGetPost(list http.HandlerFunc, create http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			list.ServeHTTP(w, r)
		case http.MethodPost:
			create.ServeHTTP(w, r)
		}
	})
}

// dispatchGetPut routes GET to get and PUT to update handler.
func dispatchGetPut(get http.HandlerFunc, update http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			get.ServeHTTP(w, r)
		case http.MethodPut:
			update.ServeHTTP(w, r)
		}
	})
}
