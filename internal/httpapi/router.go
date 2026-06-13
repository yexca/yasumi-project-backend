package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/yasumi/yasumi-project-backend/internal/config"
)

type Router struct {
	cfg    config.Config
	logger *slog.Logger
	mux    *http.ServeMux
}

func NewRouter(cfg config.Config, logger *slog.Logger) http.Handler {
	router := &Router{
		cfg:    cfg,
		logger: logger,
		mux:    http.NewServeMux(),
	}
	router.routes()
	return router
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func (r *Router) routes() {
	r.mux.HandleFunc("GET /healthz", r.health)
	r.mux.HandleFunc("GET /readyz", r.ready)
	r.mux.HandleFunc("GET /v1", r.v1)
}

func (r *Router) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (r *Router) ready(w http.ResponseWriter, _ *http.Request) {
	// Phase 01 does not connect to PostgreSQL or PowerSync yet.
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (r *Router) v1(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"name": "yasumi-api",
		"env":  r.cfg.AppEnv,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
