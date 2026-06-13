package app

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yasumi/yasumi-project-backend/internal/auth"
	"github.com/yasumi/yasumi-project-backend/internal/config"
	"github.com/yasumi/yasumi-project-backend/internal/httpapi"
	"github.com/yasumi/yasumi-project-backend/internal/repository"
	"github.com/yasumi/yasumi-project-backend/internal/service"
	"github.com/yasumi/yasumi-project-backend/internal/synctoken"
	"github.com/yasumi/yasumi-project-backend/internal/telemetry"
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
	repo := repository.New(a.pool)
	accounts := auth.NewAccountService(auth.NewRepositoryAdapter(repo), a.cfg, auth.SystemClock{})
	metrics := telemetry.NewMetrics()
	return &http.Server{
		Addr: a.cfg.HTTP.Address(),
		Handler: httpapi.NewRouter(
			a.cfg,
			a.logger,
			metrics,
			accounts,
			accounts,
			synctoken.NewHMACIssuer(a.cfg.SyncToken),
			service.NewSyncUploadService(service.NewRepositoryAdapter(repo), service.SystemClock{}),
			NewReadinessChecker(a.pool, a.cfg.PowerSync, metrics),
		),
		ReadHeaderTimeout: a.cfg.HTTP.ReadHeaderTimeout,
	}
}
