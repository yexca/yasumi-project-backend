package app

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yasumi/yasumi-project-backend/internal/config"
	"github.com/yasumi/yasumi-project-backend/internal/httpapi"
)

type ReadinessChecker struct {
	pool       *pgxpool.Pool
	syncURL    string
	httpClient *http.Client
}

func NewReadinessChecker(pool *pgxpool.Pool, cfg config.PowerSyncConfig) *ReadinessChecker {
	return &ReadinessChecker{
		pool:    pool,
		syncURL: cfg.URL,
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

func (c *ReadinessChecker) Check(ctx context.Context) httpapi.Readiness {
	return httpapi.Readiness{
		Database: c.databaseReady(ctx),
		Sync:     c.syncReady(ctx),
	}
}

func (c *ReadinessChecker) databaseReady(ctx context.Context) bool {
	if c.pool == nil {
		return false
	}
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return c.pool.Ping(pingCtx) == nil
}

func (c *ReadinessChecker) syncReady(ctx context.Context) bool {
	if c.syncURL == "" {
		return false
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.syncURL, nil)
	if err != nil {
		return false
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusInternalServerError
}
