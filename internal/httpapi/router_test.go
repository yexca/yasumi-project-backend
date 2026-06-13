package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/yasumi/yasumi-project-backend/internal/auth"
	"github.com/yasumi/yasumi-project-backend/internal/config"
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

func newTestHandler(readiness Readiness) http.Handler {
	cfg := config.MustLoad()
	cfg.HTTP.RequestTimeout = time.Second
	return NewRouter(
		cfg,
		telemetry.NewLogger(config.LogConfig{Level: "error", Format: "text"}),
		fakeAuthenticator{},
		fakeTokenIssuer{},
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
