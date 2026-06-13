package synctoken

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/yasumi/yasumi-project-backend/internal/config"
)

func TestHMACIssuerIssuesPowerSyncCompatibleJWT(t *testing.T) {
	issuer := NewHMACIssuer(config.SyncTokenConfig{
		Secret: "local-dev-sync-secret-change-me",
		TTL:    15 * time.Minute,
	})
	issuer.now = func() time.Time {
		return time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC)
	}

	token, err := issuer.Issue(context.Background(), "user-01", "device-01", "0.1.0")
	if err != nil {
		t.Fatalf("Issue error = %v", err)
	}

	parts := strings.Split(token.Value, ".")
	if len(parts) != 3 {
		t.Fatalf("token parts = %d, want JWT with three parts", len(parts))
	}

	var header map[string]string
	decodeJWTPart(t, parts[0], &header)
	if header["alg"] != "HS256" || header["kid"] != "local-dev-sync-key" {
		t.Fatalf("header = %+v, want HS256 local key", header)
	}

	var payload map[string]any
	decodeJWTPart(t, parts[1], &payload)
	if payload["sub"] != "user-01" || payload["user_id"] != "user-01" {
		t.Fatalf("payload = %+v, want user subject and claim", payload)
	}
	if payload["device_id"] != "device-01" {
		t.Fatalf("payload device_id = %v", payload["device_id"])
	}
}

func decodeJWTPart(t *testing.T, value string, target any) {
	t.Helper()
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		t.Fatalf("decode JWT part: %v", err)
	}
	if err := json.Unmarshal(decoded, target); err != nil {
		t.Fatalf("unmarshal JWT part: %v", err)
	}
}
