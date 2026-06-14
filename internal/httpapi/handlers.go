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

func (r *Router) metricsText(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if r.metrics == nil {
		return
	}
	_, _ = w.Write([]byte(r.metrics.PrometheusText()))
}

func (r *Router) register(w http.ResponseWriter, req *http.Request) {
	var body auth.RegisterRequest
	if err := decodeJSON(req, &body); err != nil {
		r.writeAPIError(w, req, http.StatusBadRequest, apiError{
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
		r.writeAPIError(w, req, status, apiErr)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (r *Router) login(w http.ResponseWriter, req *http.Request) {
	var body auth.LoginRequest
	if err := decodeJSON(req, &body); err != nil {
		r.writeAPIError(w, req, http.StatusBadRequest, apiError{
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
		r.writeAPIError(w, req, status, apiErr)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (r *Router) logout(w http.ResponseWriter, req *http.Request) {
	token, ok := bearerToken(req.Header.Get("Authorization"))
	if !ok {
		r.writeAPIError(w, req, http.StatusUnauthorized, apiError{
			Code:      string(domain.ErrorUnauthorized),
			Message:   "missing bearer token",
			Fields:    map[string]string{},
			Retryable: false,
		})
		return
	}
	if err := r.accounts.Logout(req.Context(), token); err != nil {
		status, apiErr := authOrDomainError(err)
		r.writeAPIError(w, req, status, apiErr)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (r *Router) refresh(w http.ResponseWriter, req *http.Request) {
	var body refreshRequest
	if err := decodeJSON(req, &body); err != nil {
		r.writeAPIError(w, req, http.StatusBadRequest, apiError{
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
		r.writeAPIError(w, req, status, apiErr)
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

func (r *Router) updateProfile(w http.ResponseWriter, req *http.Request) {
	var body auth.UpdateProfileRequest
	if err := decodeJSON(req, &body); err != nil {
		r.writeAPIError(w, req, http.StatusBadRequest, apiError{
			Code:      string(domain.ErrorValidationFailed),
			Message:   "invalid profile request",
			Fields:    map[string]string{"body": "invalid_json"},
			Retryable: false,
		})
		return
	}
	user := authenticatedUser(req)
	result, err := r.accounts.UpdateProfile(req.Context(), user.ID, body)
	if err != nil {
		status, apiErr := authOrDomainError(err)
		r.writeAPIError(w, req, status, apiErr)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": result})
}

func (r *Router) changePassword(w http.ResponseWriter, req *http.Request) {
	var body auth.ChangePasswordRequest
	if err := decodeJSON(req, &body); err != nil {
		r.writeAPIError(w, req, http.StatusBadRequest, apiError{
			Code:      string(domain.ErrorValidationFailed),
			Message:   "invalid password request",
			Fields:    map[string]string{"body": "invalid_json"},
			Retryable: false,
		})
		return
	}
	user := authenticatedUser(req)
	if err := r.accounts.ChangePassword(req.Context(), user.ID, body); err != nil {
		status, apiErr := authOrDomainError(err)
		r.writeAPIError(w, req, status, apiErr)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (r *Router) weather(w http.ResponseWriter, req *http.Request) {
	city := strings.TrimSpace(req.URL.Query().Get("city"))
	if city == "" {
		r.writeAPIError(w, req, http.StatusBadRequest, apiError{
			Code:      string(domain.ErrorValidationFailed),
			Message:   "city is required",
			Fields:    map[string]string{"weather_city": "required"},
			Retryable: false,
		})
		return
	}
	if len(city) > 120 {
		r.writeAPIError(w, req, http.StatusBadRequest, apiError{
			Code:      string(domain.ErrorValidationFailed),
			Message:   "city is too long",
			Fields:    map[string]string{"weather_city": "too_long"},
			Retryable: false,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"city":        city,
		"summary":     weatherSummary(city),
		"temperature": weatherTemperature(city),
		"unit":        "C",
	})
}

func (r *Router) syncToken(w http.ResponseWriter, req *http.Request) {
	var body syncTokenRequest
	if req.Body != nil {
		decoder := json.NewDecoder(req.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			r.writeAPIError(w, req, http.StatusBadRequest, apiError{
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
		r.writeAPIError(w, req, http.StatusBadRequest, apiError{
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
		r.writeAPIError(w, req, status, apiErr)
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
		r.writeAPIError(w, req, http.StatusServiceUnavailable, apiError{
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
		r.writeAPIError(w, req, http.StatusBadRequest, apiError{
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
		r.writeAPIError(w, req, status, apiErr)
		r.recordSyncUploadResult(req, "rejected")
		return
	}

	resultLabel := "accepted"
	if result.DuplicateOf != nil {
		resultLabel = "duplicate"
	}
	r.recordSyncUploadResult(req, resultLabel)
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

func weatherSummary(city string) string {
	switch strings.ToLower(city) {
	case "tokyo", "東京":
		return "Partly cloudy"
	case "shanghai", "上海":
		return "Humid"
	case "london":
		return "Light rain"
	case "new york", "new york city":
		return "Clear"
	default:
		return "Mild"
	}
}

func weatherTemperature(city string) int {
	switch strings.ToLower(city) {
	case "tokyo", "東京":
		return 24
	case "shanghai", "上海":
		return 27
	case "london":
		return 17
	case "new york", "new york city":
		return 22
	default:
		return 21
	}
}

func (r *Router) recordSyncUploadResult(req *http.Request, result string) {
	if r.metrics == nil {
		return
	}
	r.metrics.Inc("yasumi_sync_upload_results_total", map[string]string{
		"route":  routeLabel(req),
		"result": result,
	})
}
