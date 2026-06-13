package app

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yasumi/yasumi-project-backend/internal/auth"
	"github.com/yasumi/yasumi-project-backend/internal/config"
	"github.com/yasumi/yasumi-project-backend/internal/httpapi"
	"github.com/yasumi/yasumi-project-backend/internal/synctoken"
)

type App struct {
	cfg    config.Config
	logger *slog.Logger
	pool   *pgxpool.Pool
}

func New(cfg config.Config, logger *slog.Logger, pool *pgxpool.Pool) *App {
	return &App{
		cfg:    cfg,
		logger: logger,
		pool:   pool,
	}
}

func (a *App) HTTPServer() *http.Server {
	return &http.Server{
		Addr: a.cfg.HTTP.Address(),
		Handler: httpapi.NewRouter(
			a.cfg,
			a.logger,
			auth.NewDevBearerAuthenticator(a.cfg.Auth),
			synctoken.NewHMACIssuer(a.cfg.SyncToken),
			NewReadinessChecker(a.pool, a.cfg.PowerSync),
		),
		ReadHeaderTimeout: a.cfg.HTTP.ReadHeaderTimeout,
	}
}
