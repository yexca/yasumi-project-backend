package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yasumi/yasumi-project-backend/internal/auth"
	"github.com/yasumi/yasumi-project-backend/internal/domain"
)

type contextKey string

const (
	requestIDContextKey contextKey = "request_id"
	userContextKey      contextKey = "user"

	requestIDHeader = "X-Request-ID"
)

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestID := strings.TrimSpace(req.Header.Get(requestIDHeader))
		if requestID == "" {
			requestID = newRequestID()
		}
		w.Header().Set(requestIDHeader, requestID)
		ctx := context.WithValue(req.Context(), requestIDContextKey, requestID)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func timeoutMiddleware(timeout time.Duration, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), timeout)
		defer cancel()
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func (r *Router) observabilityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, req)

		route := routeLabel(req)
		status := strconv.Itoa(rec.status)
		if r.metrics != nil {
			r.metrics.Inc("yasumi_http_requests_total", map[string]string{
				"method": req.Method,
				"route":  route,
				"status": status,
			})
			if rec.status >= http.StatusBadRequest {
				r.metrics.Inc("yasumi_http_request_failures_total", map[string]string{
					"method": req.Method,
					"route":  route,
					"status": status,
				})
			}
		}
		if r.logger != nil {
			r.logger.Info("http request completed",
				"request_id", requestID(req.Context()),
				"method", req.Method,
				"route", route,
				"status", rec.status,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		}
	})
}

func (r *Router) requireAuth(next func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		token, ok := bearerToken(req.Header.Get("Authorization"))
		if !ok {
			r.recordAuthFailure(req, "missing_bearer")
			r.writeAPIError(w, req, http.StatusUnauthorized, apiError{
				Code:      string(domain.ErrorUnauthorized),
				Message:   "missing bearer token",
				Fields:    map[string]string{},
				Retryable: false,
			})
			return
		}

		user, err := r.authn.Authenticate(req.Context(), token)
		if err != nil {
			if status, apiErr := authOrDomainError(err); apiErr.Code == string(domain.ErrorAccountDisabled) {
				r.recordAuthFailure(req, apiErr.Code)
				r.writeAPIError(w, req, status, apiErr)
				return
			}
			r.recordAuthFailure(req, "invalid_bearer")
			r.writeAPIError(w, req, http.StatusUnauthorized, apiError{
				Code:      string(domain.ErrorUnauthorized),
				Message:   "invalid bearer token",
				Fields:    map[string]string{},
				Retryable: false,
			})
			return
		}

		ctx := context.WithValue(req.Context(), userContextKey, user)
		next(w, req.WithContext(ctx))
	}
}

func (r *Router) recordAuthFailure(req *http.Request, reason string) {
	if r.metrics == nil {
		return
	}
	r.metrics.Inc("yasumi_auth_failures_total", map[string]string{
		"route":  routeLabel(req),
		"reason": reason,
	})
}

func authenticatedUser(req *http.Request) auth.User {
	user, _ := req.Context().Value(userContextKey).(auth.User)
	return user
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return token, token != ""
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(b[:])
}

func requestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDContextKey).(string)
	return id
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func routeLabel(req *http.Request) string {
	path := req.URL.Path
	switch path {
	case "/healthz", "/readyz", "/metrics":
		return path
	case "/v1/auth/register", "/v1/auth/login", "/v1/auth/logout", "/v1/auth/refresh", "/v1/session", "/v1/sync/token", "/v1/sync/upload":
		return path
	default:
		return "unmatched"
	}
}
