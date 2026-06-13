package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/yasumi/yasumi-project-backend/internal/auth"
	"github.com/yasumi/yasumi-project-backend/internal/config"
	"github.com/yasumi/yasumi-project-backend/internal/synctoken"
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
	authn     auth.Authenticator
	tokens    synctoken.Issuer
	readiness ReadinessChecker
	mux       *http.ServeMux
}

func NewRouter(
	cfg config.Config,
	logger *slog.Logger,
	authn auth.Authenticator,
	tokens synctoken.Issuer,
	readiness ReadinessChecker,
) http.Handler {
	router := &Router{
		cfg:       cfg,
		logger:    logger,
		authn:     authn,
		tokens:    tokens,
		readiness: readiness,
		mux:       http.NewServeMux(),
	}
	router.routes()
	return router.middleware(router.mux)
}

func (r *Router) routes() {
	r.mux.HandleFunc("GET /healthz", r.health)
	r.mux.HandleFunc("GET /readyz", r.ready)
	r.mux.HandleFunc("GET /v1/session", r.requireAuth(r.session))
	r.mux.HandleFunc("POST /v1/sync/token", r.requireAuth(r.syncToken))
}

func (r *Router) middleware(next http.Handler) http.Handler {
	return requestIDMiddleware(timeoutMiddleware(r.cfg.HTTP.RequestTimeout, next))
}
