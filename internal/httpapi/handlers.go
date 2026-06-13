package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/yasumi/yasumi-project-backend/internal/domain"
)

func (r *Router) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (r *Router) ready(w http.ResponseWriter, req *http.Request) {
	readiness := r.readiness.Check(req.Context())
	checks := map[string]string{
		"database": checkStatus(readiness.Database),
		"sync":     checkStatus(readiness.Sync),
	}
	if readiness.Database && readiness.Sync {
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "ready",
			"checks": checks,
		})
		return
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"status": "not_ready",
		"checks": checks,
	})
}

func (r *Router) session(w http.ResponseWriter, req *http.Request) {
	user := authenticatedUser(req)
	writeJSON(w, http.StatusOK, map[string]any{
		"user": map[string]string{
			"id":           user.ID,
			"display_name": user.DisplayName,
		},
		"session": map[string]bool{
			"authenticated": true,
		},
	})
}

func (r *Router) syncToken(w http.ResponseWriter, req *http.Request) {
	var body syncTokenRequest
	if req.Body != nil {
		decoder := json.NewDecoder(req.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			writeAPIError(w, http.StatusBadRequest, apiError{
				Code:      string(domain.ErrorValidationFailed),
				Message:   "invalid sync token request",
				Fields:    map[string]string{"body": "invalid_json"},
				Retryable: false,
			})
			return
		}
	}

	body.DeviceID = strings.TrimSpace(body.DeviceID)
	body.ClientVersion = strings.TrimSpace(body.ClientVersion)
	if body.DeviceID == "" {
		writeAPIError(w, http.StatusBadRequest, apiError{
			Code:      string(domain.ErrorValidationFailed),
			Message:   "device_id is required",
			Fields:    map[string]string{"device_id": "required"},
			Retryable: false,
		})
		return
	}

	user := authenticatedUser(req)
	token, err := r.tokens.Issue(req.Context(), user.ID, body.DeviceID, body.ClientVersion)
	if err != nil {
		status, apiErr := domainError(err)
		writeAPIError(w, status, apiErr)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"token":      token.Value,
		"expires_at": token.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		"stream_scope": map[string]string{
			"user_id": token.UserID,
		},
	})
}

type syncTokenRequest struct {
	DeviceID      string `json:"device_id"`
	ClientVersion string `json:"client_version"`
	UserID        string `json:"user_id"`
}

func checkStatus(ok bool) string {
	if ok {
		return "ok"
	}
	return "unavailable"
}
