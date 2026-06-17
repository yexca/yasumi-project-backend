package config

import "testing"

func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("YASUMI_SYNC_TOKEN_SECRET", "test-sync-secret")
}

func TestLoadDefaults(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("YASUMI_HTTP_PORT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTP.Port != 7659 {
		t.Fatalf("HTTP.Port = %d, want 7659", cfg.HTTP.Port)
	}
	if len(cfg.HTTP.AllowedOrigins) == 0 {
		t.Fatal("HTTP.AllowedOrigins is empty")
	}
	if cfg.Postgres.Database != "yasumi" {
		t.Fatalf("Postgres.Database = %q, want yasumi", cfg.Postgres.Database)
	}
	if cfg.PowerSync.PublicURL != "" {
		t.Fatalf("PowerSync.PublicURL = %q, want empty", cfg.PowerSync.PublicURL)
	}
}

func TestLoadParsesAllowedOrigins(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("YASUMI_HTTP_ALLOWED_ORIGINS", " http://127.0.0.1:5173, ,http://localhost:4175 ")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got, want := len(cfg.HTTP.AllowedOrigins), 2; got != want {
		t.Fatalf("len(AllowedOrigins) = %d, want %d", got, want)
	}
	if cfg.HTTP.AllowedOrigins[0] != "http://127.0.0.1:5173" {
		t.Fatalf("AllowedOrigins[0] = %q", cfg.HTTP.AllowedOrigins[0])
	}
}

func TestValidateRejectsInvalidPort(t *testing.T) {
	setRequiredEnv(t)
	cfg := MustLoad()
	cfg.HTTP.Port = 70000

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}

func TestLoadRejectsInvalidInteger(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("YASUMI_HTTP_PORT", "not-a-port")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}

func TestLoadRejectsMissingSyncTokenSecret(t *testing.T) {
	t.Setenv("YASUMI_SYNC_TOKEN_SECRET", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}
