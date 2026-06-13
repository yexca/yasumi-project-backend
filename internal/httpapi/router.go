package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/yasumi/yasumi-project-backend/internal/auth"
	"github.com/yasumi/yasumi-project-backend/internal/config"
	"github.com/yasumi/yasumi-project-backend/internal/service"
	"github.com/yasumi/yasumi-project-backend/internal/synctoken"
	"github.com/yasumi/yasumi-project-backend/internal/telemetry"
)

type ReadinessChecker interface {
	Check(ctx context.Context) Readiness
}

type Readiness struct {
	Database bool
	Sync     bool
}

type Router struct {
	cfg       config.Config
	logger    *slog.Logger
	metrics   *telemetry.Metrics
	authn     auth.Authenticator
	accounts  AccountService
	tokens    synctoken.Issuer
	sync      SyncUploadAcceptor
	readiness ReadinessChecker
	mux       *http.ServeMux
}

type SyncUploadAcceptor interface {
	AcceptUpload(ctx context.Context, userID string, upload service.SyncUpload) (service.SyncUploadResult, error)
}

type AccountService interface {
	Register(ctx context.Context, req auth.RegisterRequest) (auth.AuthResponse, error)
	Login(ctx context.Context, req auth.LoginRequest) (auth.AuthResponse, error)
	Refresh(ctx context.Context, refreshToken string) (auth.AuthResponse, error)
	Logout(ctx context.Context, accessToken string) error
}

func NewRouter(
	cfg config.Config,
	logger *slog.Logger,
	metrics *telemetry.Metrics,
	authn auth.Authenticator,
	accounts AccountService,
	tokens synctoken.Issuer,
	sync SyncUploadAcceptor,
	readiness ReadinessChecker,
) http.Handler {
	router := &Router{
		cfg:       cfg,
		logger:    logger,
		metrics:   metrics,
		authn:     authn,
		accounts:  accounts,
		tokens:    tokens,
		sync:      sync,
		readiness: readiness,
		mux:       http.NewServeMux(),
	}
	router.routes()
	return router.middleware(router.mux)
}

func (r *Router) routes() {
	r.mux.HandleFunc("GET /healthz", r.health)
	r.mux.HandleFunc("GET /readyz", r.ready)
	r.mux.HandleFunc("GET /metrics", r.metricsText)
	r.mux.HandleFunc("POST /v1/auth/register", r.register)
	r.mux.HandleFunc("POST /v1/auth/login", r.login)
	r.mux.HandleFunc("POST /v1/auth/logout", r.requireAuth(r.logout))
	r.mux.HandleFunc("POST /v1/auth/refresh", r.refresh)
	r.mux.HandleFunc("GET /v1/session", r.requireAuth(r.session))
	r.mux.HandleFunc("POST /v1/sync/token", r.requireAuth(r.syncToken))
	r.mux.HandleFunc("POST /v1/sync/upload", r.requireAuth(r.syncUpload))
}

func (r *Router) middleware(next http.Handler) http.Handler {
	return requestIDMiddleware(timeoutMiddleware(r.cfg.HTTP.RequestTimeout, r.observabilityMiddleware(next)))
}
