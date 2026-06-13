package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
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

func (r *Router) requireAuth(next func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		token, ok := bearerToken(req.Header.Get("Authorization"))
		if !ok {
			writeAPIError(w, http.StatusUnauthorized, apiError{
				Code:      string(domain.ErrorUnauthorized),
				Message:   "missing bearer token",
				Fields:    map[string]string{},
				Retryable: false,
			})
			return
		}

		user, err := r.authn.Authenticate(req.Context(), token)
		if err != nil {
			writeAPIError(w, http.StatusUnauthorized, apiError{
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
