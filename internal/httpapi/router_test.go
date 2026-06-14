package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/yasumi/yasumi-project-backend/internal/auth"
	"github.com/yasumi/yasumi-project-backend/internal/config"
	"github.com/yasumi/yasumi-project-backend/internal/domain"
	"github.com/yasumi/yasumi-project-backend/internal/service"
	"github.com/yasumi/yasumi-project-backend/internal/synctoken"
	"github.com/yasumi/yasumi-project-backend/internal/telemetry"
)

const (
	testUserID = "00000000-0000-4000-8000-000000000001"
	testToken  = "session-token"
)

func TestHealthzIsPublic(t *testing.T) {
	handler := newTestHandler(Readiness{Database: true, Sync: true})

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

func TestCORSAllowsLocalFrontendPreflight(t *testing.T) {
	handler := newTestHandler(Readiness{Database: true, Sync: true})

	req := httptest.NewRequest(http.MethodOptions, "/v1/session", nil)
	req.Header.Set("Origin", "http://127.0.0.1:5173")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "Authorization")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://127.0.0.1:5173" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
}

func TestReadyzReportsDependencies(t *testing.T) {
	for _, tt := range []struct {
		name       string
		readiness  Readiness
		wantStatus int
		wantBody   string
	}{
		{
			name:       "ready",
			readiness:  Readiness{Database: true, Sync: true},
			wantStatus: http.StatusOK,
			wantBody:   `"status":"ready"`,
		},
		{
			name:       "not ready",
			readiness:  Readiness{Database: true, Sync: false},
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   `"sync":"unavailable"`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			handler := newTestHandler(tt.readiness)
			req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Fatalf("body = %s, want to contain %s", rec.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestSessionRequiresAuthentication(t *testing.T) {
	handler := newTestHandler(Readiness{Database: true, Sync: true})

	req := httptest.NewRequest(http.MethodGet, "/v1/session", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	assertErrorShape(t, rec.Body.Bytes(), "unauthorized")
}

func TestSessionReturnsAuthenticatedUser(t *testing.T) {
	handler := newTestHandler(Readiness{Database: true, Sync: true})

	req := httptest.NewRequest(http.MethodGet, "/v1/session", nil)
	req.Header.Set("Authorization", "Bearer "+testToken)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), testUserID) {
		t.Fatalf("body = %s, want user id %s", rec.Body.String(), testUserID)
	}
	if rec.Header().Get(requestIDHeader) == "" {
		t.Fatal("missing request id response header")
	}
}

func TestRegisterReturnsInitialSession(t *testing.T) {
	handler := newTestHandler(Readiness{Database: true, Sync: true})

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", strings.NewReader(`{
		"username":"yasumi",
		"email":"user@example.com",
		"password":"password123"
	}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "password123") {
		t.Fatal("response leaked raw password")
	}
	if !strings.Contains(rec.Body.String(), `"access_token":"access-token"`) {
		t.Fatalf("body = %s, want access token", rec.Body.String())
	}
}

func TestRegisterDuplicateUsernameUsesStableError(t *testing.T) {
	handler := newTestHandler(Readiness{Database: true, Sync: true})

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", strings.NewReader(`{
		"username":"taken",
		"email":"user@example.com",
		"password":"password123"
	}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusConflict, rec.Body.String())
	}
	assertErrorShape(t, rec.Body.Bytes(), "username_already_taken")
}

func TestLoginInvalidCredentialsAreGeneric(t *testing.T) {
	handler := newTestHandler(Readiness{Database: true, Sync: true})

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(`{
		"identifier":"bad",
		"password":"wrong-password"
	}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	assertErrorShape(t, rec.Body.Bytes(), "invalid_credentials")
}

func TestRefreshReturnsRotatedSession(t *testing.T) {
	handler := newTestHandler(Readiness{Database: true, Sync: true})

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", strings.NewReader(`{"refresh_token":"refresh-token"}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"refresh_token":"new-refresh-token"`) {
		t.Fatalf("body = %s, want rotated refresh token", rec.Body.String())
	}
}

func TestLogoutRevokesCurrentSession(t *testing.T) {
	handler := newTestHandler(Readiness{Database: true, Sync: true})

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+testToken)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
}

func TestSyncTokenRequiresAuthentication(t *testing.T) {
	handler := newTestHandler(Readiness{Database: true, Sync: true})

	req := httptest.NewRequest(http.MethodPost, "/v1/sync/token", strings.NewReader(`{"device_id":"device-01"}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	assertErrorShape(t, rec.Body.Bytes(), "unauthorized")
}

func TestSyncTokenIgnoresClientSuppliedUserID(t *testing.T) {
	handler := newTestHandler(Readiness{Database: true, Sync: true})

	req := httptest.NewRequest(http.MethodPost, "/v1/sync/token", strings.NewReader(`{
		"device_id":"device-01",
		"client_version":"0.1.0",
		"user_id":"00000000-0000-4000-8000-999999999999"
	}`))
	req.Header.Set("Authorization", "Bearer "+testToken)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"user_id":"`+testUserID+`"`) {
		t.Fatalf("body = %s, want authenticated user scope", rec.Body.String())
	}
}

func TestSyncTokenValidationErrorShape(t *testing.T) {
	handler := newTestHandler(Readiness{Database: true, Sync: true})

	req := httptest.NewRequest(http.MethodPost, "/v1/sync/token", strings.NewReader(`{"device_id":"  "}`))
	req.Header.Set("Authorization", "Bearer "+testToken)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	assertErrorShape(t, rec.Body.Bytes(), "validation_failed")
}

func TestNoBusinessCRUDRoutes(t *testing.T) {
	handler := newTestHandler(Readiness{Database: true, Sync: true})

	for _, path := range []string{
		"/v1/items",
		"/v1/areas",
		"/v1/recurring-task-templates",
		"/v1/operation-history",
		"/v1/user-settings",
	} {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		req.Header.Set("Authorization", "Bearer "+testToken)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s status = %d, want %d", path, rec.Code, http.StatusNotFound)
		}
	}
}

func TestSyncUploadRequiresAuthentication(t *testing.T) {
	handler := newTestHandler(Readiness{Database: true, Sync: true})

	req := httptest.NewRequest(http.MethodPost, "/v1/sync/upload", strings.NewReader(`{"device_id":"device-01","mutations":[]}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	assertErrorShape(t, rec.Body.Bytes(), "unauthorized")
}

func TestSyncUploadReturnsAcceptedResult(t *testing.T) {
	handler := newTestHandler(Readiness{Database: true, Sync: true})

	req := httptest.NewRequest(http.MethodPost, "/v1/sync/upload", strings.NewReader(`{
		"client_batch_id":"batch-01",
		"device_id":"device-01",
		"mutations":[]
	}`))
	req.Header.Set("Authorization", "Bearer "+testToken)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusAccepted, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"client_batch_id":"batch-01"`) {
		t.Fatalf("body = %s, want accepted batch", rec.Body.String())
	}
}

func TestSyncUploadUsesStableErrorShape(t *testing.T) {
	handler := newTestHandler(Readiness{Database: true, Sync: true})

	req := httptest.NewRequest(http.MethodPost, "/v1/sync/upload", strings.NewReader(`{"device_id":"forbidden"}`))
	req.Header.Set("Authorization", "Bearer "+testToken)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	assertErrorShape(t, rec.Body.Bytes(), "forbidden")
}

func TestMetricsExposeRequestFailureAuthAndSyncCounters(t *testing.T) {
	handler := newTestHandler(Readiness{Database: true, Sync: true})

	req := httptest.NewRequest(http.MethodGet, "/v1/session", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	req = httptest.NewRequest(http.MethodPost, "/v1/sync/upload", strings.NewReader(`{
		"client_batch_id":"batch-01",
		"device_id":"device-01",
		"mutations":[]
	}`))
	req.Header.Set("Authorization", "Bearer "+testToken)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`yasumi_auth_failures_total{reason="missing_bearer",route="/v1/session"} 1`,
		`yasumi_http_request_failures_total{method="GET",route="/v1/session",status="401"} 1`,
		`yasumi_sync_upload_results_total{result="accepted",route="/v1/sync/upload"} 1`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("metrics = %s, want to contain %s", body, want)
		}
	}
}

func TestRequestLoggingDoesNotLeakSensitiveInputs(t *testing.T) {
	var logs bytes.Buffer
	handler := newTestHandlerWithLogger(
		Readiness{Database: true, Sync: true},
		slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelInfo})),
	)
	secretBody := `{"identifier":"user@example.com","password":"super-secret-password"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(secretBody))
	req.Header.Set("Authorization", "Bearer raw-access-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	output := logs.String()
	for _, forbidden := range []string{"super-secret-password", "raw-access-token", secretBody, "Authorization"} {
		if strings.Contains(output, forbidden) {
			t.Fatalf("log leaked %q: %s", forbidden, output)
		}
	}
	if !strings.Contains(output, `"/v1/auth/login"`) {
		t.Fatalf("log = %s, want route label", output)
	}
}

func newTestHandler(readiness Readiness) http.Handler {
	return newTestHandlerWithLogger(readiness, telemetry.NewLogger(config.LogConfig{Level: "error", Format: "text"}))
}

func newTestHandlerWithLogger(readiness Readiness, logger *slog.Logger) http.Handler {
	cfg := config.MustLoad()
	cfg.HTTP.RequestTimeout = time.Second
	return NewRouter(
		cfg,
		logger,
		telemetry.NewMetrics(),
		fakeAuthenticator{},
		fakeAccountService{},
		fakeTokenIssuer{},
		fakeSyncUploadAcceptor{},
		staticReadiness{readiness: readiness},
	)
}

type fakeAuthenticator struct{}

func (fakeAuthenticator) Authenticate(_ context.Context, token string) (auth.User, error) {
	if token != testToken {
		return auth.User{}, errors.New("bad token")
	}
	return auth.User{ID: testUserID, DisplayName: "Yasumi User"}, nil
}

type fakeTokenIssuer struct{}

func (fakeTokenIssuer) Issue(_ context.Context, userID, deviceID, clientVersion string) (synctoken.Token, error) {
	if deviceID == "" {
		return synctoken.Token{}, errors.New("missing device")
	}
	return synctoken.Token{
		Value:     "sync-token",
		ExpiresAt: time.Date(2026, 6, 13, 10, 30, 0, 0, time.UTC),
		UserID:    userID,
	}, nil
}

type fakeAccountService struct{}

func (fakeAccountService) Register(_ context.Context, req auth.RegisterRequest) (auth.AuthResponse, error) {
	if req.Username == "taken" {
		return auth.AuthResponse{}, &domain.Error{
			Code:      domain.ErrorUsernameAlreadyTaken,
			Message:   "username is already taken",
			Fields:    map[domain.FieldKey]string{domain.FieldUsername: "already_taken"},
			Retryable: false,
		}
	}
	displayName := "Test User"
	return auth.AuthResponse{
		User: auth.AccountUserDTO{
			ID:          testUserID,
			Username:    req.Username,
			Email:       req.Email,
			DisplayName: &displayName,
		},
		Session: auth.AuthSessionDTO{
			Authenticated: true,
			AccessToken:   "access-token",
			RefreshToken:  "refresh-token",
			ExpiresAt:     "2026-06-13T10:30:00Z",
		},
	}, nil
}

func (fakeAccountService) Login(_ context.Context, req auth.LoginRequest) (auth.AuthResponse, error) {
	if req.Identifier == "bad" {
		return auth.AuthResponse{}, &domain.Error{
			Code:      domain.ErrorInvalidCredentials,
			Message:   "invalid credentials",
			Fields:    map[domain.FieldKey]string{},
			Retryable: false,
		}
	}
	displayName := "Test User"
	return auth.AuthResponse{
		User: auth.AccountUserDTO{
			ID:          testUserID,
			Username:    req.Identifier,
			Email:       "user@example.com",
			DisplayName: &displayName,
		},
		Session: auth.AuthSessionDTO{
			Authenticated: true,
			AccessToken:   "access-token",
			RefreshToken:  "refresh-token",
			ExpiresAt:     "2026-06-13T10:30:00Z",
		},
	}, nil
}

func (fakeAccountService) Refresh(_ context.Context, refreshToken string) (auth.AuthResponse, error) {
	if refreshToken == "expired" {
		return auth.AuthResponse{}, &domain.Error{
			Code:      domain.ErrorSessionExpired,
			Message:   "session expired",
			Fields:    map[domain.FieldKey]string{},
			Retryable: false,
		}
	}
	displayName := "Test User"
	return auth.AuthResponse{
		User: auth.AccountUserDTO{
			ID:          testUserID,
			Username:    "yasumi",
			Email:       "user@example.com",
			DisplayName: &displayName,
		},
		Session: auth.AuthSessionDTO{
			Authenticated: true,
			AccessToken:   "new-access-token",
			RefreshToken:  "new-refresh-token",
			ExpiresAt:     "2026-06-13T10:30:00Z",
		},
	}, nil
}

func (fakeAccountService) Logout(_ context.Context, token string) error {
	if token != testToken {
		return auth.ErrUnauthenticated
	}
	return nil
}

type fakeSyncUploadAcceptor struct{}

func (fakeSyncUploadAcceptor) AcceptUpload(_ context.Context, userID string, upload service.SyncUpload) (service.SyncUploadResult, error) {
	if userID != testUserID {
		return service.SyncUploadResult{}, errors.New("wrong user")
	}
	if upload.DeviceID == "forbidden" {
		return service.SyncUploadResult{}, &domain.Error{
			Code:      domain.ErrorForbidden,
			Message:   "row is owned by another user",
			Fields:    map[domain.FieldKey]string{domain.FieldUserID: "owner_mismatch"},
			Retryable: false,
		}
	}
	return service.SyncUploadResult{
		ClientBatchID: upload.ClientBatchID,
		Accepted:      []service.AcceptedWrite{},
	}, nil
}

type staticReadiness struct {
	readiness Readiness
}

func (s staticReadiness) Check(context.Context) Readiness {
	return s.readiness
}

func assertErrorShape(t *testing.T, body []byte, code string) {
	t.Helper()
	var got apiError
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal error body: %v; body=%s", err, string(body))
	}
	if got.Code != code {
		t.Fatalf("code = %q, want %q", got.Code, code)
	}
	if got.Message == "" {
		t.Fatal("message is empty")
	}
	if got.Fields == nil {
		t.Fatal("fields map is nil")
	}
}
