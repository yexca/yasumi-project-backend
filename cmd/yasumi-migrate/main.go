package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/yasumi/yasumi-project-backend/internal/config"
	"github.com/yasumi/yasumi-project-backend/internal/migrations"
	"github.com/yasumi/yasumi-project-backend/internal/repository"
	"github.com/yasumi/yasumi-project-backend/internal/telemetry"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuration failed", "error", err)
		os.Exit(1)
	}

	logger := telemetry.NewLogger(cfg.Log)
	slog.SetDefault(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := repository.OpenPool(ctx, cfg.Postgres)
	if err != nil {
		logger.Error("postgres connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := migrations.Apply(ctx, pool); err != nil {
		logger.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	logger.Info("migrations applied")
}
