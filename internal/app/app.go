package app

import (
	"log/slog"
	"net/http"

	"github.com/yasumi/yasumi-project-backend/internal/config"
	"github.com/yasumi/yasumi-project-backend/internal/httpapi"
)

type App struct {
	cfg    config.Config
	logger *slog.Logger
}

func New(cfg config.Config, logger *slog.Logger) *App {
	return &App{
		cfg:    cfg,
		logger: logger,
	}
}

func (a *App) HTTPServer() *http.Server {
	return &http.Server{
		Addr:              a.cfg.HTTP.Address(),
		Handler:           httpapi.NewRouter(a.cfg, a.logger),
		ReadHeaderTimeout: a.cfg.HTTP.ReadHeaderTimeout,
	}
}
