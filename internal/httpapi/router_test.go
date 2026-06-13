package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yasumi/yasumi-project-backend/internal/config"
	"github.com/yasumi/yasumi-project-backend/internal/telemetry"
)

func TestHealthz(t *testing.T) {
	handler := NewRouter(config.MustLoad(), telemetry.NewLogger(config.LogConfig{
		Level:  "error",
		Format: "text",
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if contentType := rec.Header().Get("Content-Type"); contentType != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want application/json; charset=utf-8", contentType)
	}
}

func TestNoBusinessCRUDRoutes(t *testing.T) {
	handler := NewRouter(config.MustLoad(), telemetry.NewLogger(config.LogConfig{
		Level:  "error",
		Format: "text",
	}))

	for _, path := range []string{"/v1/items", "/v1/areas", "/v1/recurring_task_templates", "/v1/operation_history", "/v1/user_settings"} {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s status = %d, want %d", path, rec.Code, http.StatusNotFound)
		}
	}
}
