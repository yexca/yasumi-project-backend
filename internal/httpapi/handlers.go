package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/yasumi/yasumi-project-backend/internal/auth"
	"github.com/yasumi/yasumi-project-backend/internal/domain"
	"github.com/yasumi/yasumi-project-backend/internal/service"
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

func (r *Router) register(w http.ResponseWriter, req *http.Request) {
	var body auth.RegisterRequest
	if err := decodeJSON(req, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, apiError{
			Code:      string(domain.ErrorValidationFailed),
			Message:   "invalid register request",
			Fields:    map[string]string{"body": "invalid_json"},
			Retryable: false,
		})
		return
	}
	result, err := r.accounts.Register(req.Context(), body)
	if err != nil {
		status, apiErr := authOrDomainError(err)
		writeAPIError(w, status, apiErr)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (r *Router) login(w http.ResponseWriter, req *http.Request) {
	var body auth.LoginRequest
	if err := decodeJSON(req, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, apiError{
			Code:      string(domain.ErrorValidationFailed),
			Message:   "invalid login request",
			Fields:    map[string]string{"body": "invalid_json"},
			Retryable: false,
		})
		return
	}
	result, err := r.accounts.Login(req.Context(), body)
	if err != nil {
		status, apiErr := authOrDomainError(err)
		writeAPIError(w, status, apiErr)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (r *Router) logout(w http.ResponseWriter, req *http.Request) {
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
	if err := r.accounts.Logout(req.Context(), token); err != nil {
		status, apiErr := authOrDomainError(err)
		writeAPIError(w, status, apiErr)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (r *Router) refresh(w http.ResponseWriter, req *http.Request) {
	var body refreshRequest
	if err := decodeJSON(req, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, apiError{
			Code:      string(domain.ErrorValidationFailed),
			Message:   "invalid refresh request",
			Fields:    map[string]string{"body": "invalid_json"},
			Retryable: false,
		})
		return
	}
	result, err := r.accounts.Refresh(req.Context(), body.RefreshToken)
	if err != nil {
		status, apiErr := authOrDomainError(err)
		writeAPIError(w, status, apiErr)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (r *Router) session(w http.ResponseWriter, req *http.Request) {
	user := authenticatedUser(req)
	displayName := user.DisplayName
	writeJSON(w, http.StatusOK, map[string]any{
		"user": map[string]any{
			"id":           user.ID,
			"username":     user.Username,
			"email":        user.Email,
			"display_name": displayName,
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

func (r *Router) syncUpload(w http.ResponseWriter, req *http.Request) {
	if r.sync == nil {
		writeAPIError(w, http.StatusServiceUnavailable, apiError{
			Code:      string(domain.ErrorServiceUnavailable),
			Message:   "sync upload service is unavailable",
			Fields:    map[string]string{},
			Retryable: true,
		})
		return
	}

	var upload service.SyncUpload
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&upload); err != nil {
		writeAPIError(w, http.StatusBadRequest, apiError{
			Code:      string(domain.ErrorValidationFailed),
			Message:   "invalid sync upload request",
			Fields:    map[string]string{"body": "invalid_json"},
			Retryable: false,
		})
		return
	}

	user := authenticatedUser(req)
	result, err := r.sync.AcceptUpload(req.Context(), user.ID, upload)
	if err != nil {
		status, apiErr := domainError(err)
		writeAPIError(w, status, apiErr)
		return
	}

	writeJSON(w, http.StatusAccepted, result)
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type syncTokenRequest struct {
	DeviceID      string `json:"device_id"`
	ClientVersion string `json:"client_version"`
	UserID        string `json:"user_id"`
}

func decodeJSON(req *http.Request, out any) error {
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(out)
}

func checkStatus(ok bool) string {
	if ok {
		return "ok"
	}
	return "unavailable"
}
