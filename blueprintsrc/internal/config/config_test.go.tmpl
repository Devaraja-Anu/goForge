package config

import (
	"strings"
	"testing"
)

func TestSecretMasking(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"standard", "super-secret"},
		{"empty", ""},
		{"special chars", "!@#$%^"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := secret(tt.input)

			if got := s.String(); got != "********" {
				t.Errorf("String() = %q, want %q", got, "********")
			}
			if got := s.LogValue().String(); got != "********" {
				t.Errorf("LogValue().String() = %q, want %q", got, "********")
			}
		})
	}
}

func TestSecretReveal(t *testing.T) {
	s := secret("super-secret")

	if got := s.Reveal(); got != "super-secret" {
		t.Errorf("Reveal() = %q, want %q", got, "super-secret")
	}
}

func TestGetEnv(t *testing.T) {
	t.Setenv("TEST_ENV", "value")

	if got := getEnv("TEST_ENV", "fallback"); got != "value" {
		t.Errorf("got %q, want %q", got, "value")
	}
	if got := getEnv("DOES_NOT_EXIST", "fallback"); got != "fallback" {
		t.Errorf("got %q, want %q", got, "fallback")
	}
}

func TestValidateSuccess(t *testing.T) {
	cfg := Config{
		DatabaseURL: "postgres://localhost/db",
		DBMaxConns:  25, // matches config.Load()'s DB_MAX_CONNS default
		DBMinConns:  5,  // matches config.Load()'s DB_MIN_CONNS default
	}

	if err := cfg.validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateMissingDatabaseURL(t *testing.T) {
	cfg := Config{}

	err := cfg.validate()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Errorf("error = %q, want it to mention DATABASE_URL", err.Error())
	}
}

func TestValidateDBPoolBounds(t *testing.T) {
	tests := []struct {
		name       string
		dbMaxConns int32
		dbMinConns int32
		wantErr    bool
	}{
		{"valid", 25, 5, false},
		{"min equals max", 10, 10, false},
		{"min zero is allowed", 25, 0, false},
		{"max zero is rejected", 0, 0, true},
		{"max negative is rejected", -1, 0, true},
		{"min negative is rejected", 25, -1, true},
		{"min greater than max is rejected", 5, 25, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				DatabaseURL: "postgres://localhost/db",
				DBMaxConns:  tt.dbMaxConns,
				DBMinConns:  tt.dbMinConns,
			}

			err := cfg.validate()
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("DATABASE_URL", "postgres://localhost/testdb")
	t.Setenv("PORT", "9000")
	t.Setenv("FRONTEND_URL", "http://localhost:3000")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.AppEnv != AppEnv("test") {
		t.Errorf("AppEnv = %q, want %q", cfg.AppEnv, "test")
	}
	if cfg.Port != "9000" {
		t.Errorf("Port = %q, want %q", cfg.Port, "9000")
	}
	if cfg.FrontendURL != "http://localhost:3000" {
		t.Errorf("FrontendURL = %q, want %q", cfg.FrontendURL, "http://localhost:3000")
	}
	if cfg.DatabaseURL.Reveal() != "postgres://localhost/testdb" {
		t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL.Reveal(), "postgres://localhost/testdb")
	}
	if cfg.DBMaxConns != 25 {
		t.Errorf("DBMaxConns = %d, want %d (default)", cfg.DBMaxConns, 25)
	}
	if cfg.DBMinConns != 5 {
		t.Errorf("DBMinConns = %d, want %d (default)", cfg.DBMinConns, 5)
	}
}

func TestLoadMissingDatabaseURL(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("DATABASE_URL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Errorf("error = %q, want it to mention DATABASE_URL", err.Error())
	}
}
