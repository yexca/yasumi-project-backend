package app

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yasumi/yasumi-project-backend/internal/config"
	"github.com/yasumi/yasumi-project-backend/internal/httpapi"
	"github.com/yasumi/yasumi-project-backend/internal/telemetry"
)

type ReadinessChecker struct {
	pool       *pgxpool.Pool
	syncURL    string
	httpClient *http.Client
	metrics    *telemetry.Metrics
}

func NewReadinessChecker(pool *pgxpool.Pool, cfg config.PowerSyncConfig, metrics *telemetry.Metrics) *ReadinessChecker {
	return &ReadinessChecker{
		pool:    pool,
		syncURL: cfg.URL,
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
		},
		metrics: metrics,
	}
}

func (c *ReadinessChecker) Check(ctx context.Context) httpapi.Readiness {
	readiness := httpapi.Readiness{
		Database: c.databaseReady(ctx),
		Sync:     c.syncReady(ctx),
	}
	c.metrics.Set("yasumi_dependency_ready", map[string]string{"dependency": "database"}, boolGauge(readiness.Database))
	c.metrics.Set("yasumi_dependency_ready", map[string]string{"dependency": "sync"}, boolGauge(readiness.Sync))
	return readiness
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

func boolGauge(ok bool) float64 {
	if ok {
		return 1
	}
	return 0
}
