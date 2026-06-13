package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yasumi/yasumi-project-backend/internal/app"
	"github.com/yasumi/yasumi-project-backend/internal/config"
	"github.com/yasumi/yasumi-project-backend/internal/repository"
	"github.com/yasumi/yasumi-project-backend/internal/telemetry"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuration failed", "error", err)
		os.Exit(1)
	}

	logger := telemetry.NewLogger(cfg.Log)
	slog.SetDefault(logger)

	pool, err := repository.OpenPool(ctx, cfg.Postgres)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	application := app.New(cfg, logger, pool)
	server := application.HTTPServer()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting yasumi api", "addr", server.Addr, "env", cfg.AppEnv)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
		defer cancel()

		logger.Info("shutting down yasumi api")
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
			os.Exit(1)
		}
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("yasumi api stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}

	if err := waitForServerStop(errCh, cfg.HTTP.ShutdownTimeout); err != nil {
		logger.Error("server stop wait failed", "error", err)
		os.Exit(1)
	}

	logger.Info("yasumi api stopped")
}

func waitForServerStop(errCh <-chan error, timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-timer.C:
		return nil
	}
}
