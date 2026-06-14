package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("YASUMI_HTTP_PORT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTP.Port != 7659 {
		t.Fatalf("HTTP.Port = %d, want 7659", cfg.HTTP.Port)
	}
	if cfg.Postgres.Database != "yasumi" {
		t.Fatalf("Postgres.Database = %q, want yasumi", cfg.Postgres.Database)
	}
}

func TestValidateRejectsInvalidPort(t *testing.T) {
	cfg := MustLoad()
	cfg.HTTP.Port = 70000

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}

func TestLoadRejectsInvalidInteger(t *testing.T) {
	t.Setenv("YASUMI_HTTP_PORT", "not-a-port")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}

func TestValidateRejectsLocalSecretsOutsideLocal(t *testing.T) {
	cfg := MustLoad()
	cfg.AppEnv = "production"

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}
