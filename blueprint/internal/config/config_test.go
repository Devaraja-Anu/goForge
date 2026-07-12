package config

import (
	"testing"

	"github.com/stretchr/testify/require"
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

			require.Equal(t, "********", s.String())
			require.Equal(t, "********", s.LogValue().String())
		})
	}
}

func TestSecretReveal(t *testing.T) {
	s := secret("super-secret")

	require.Equal(t, "super-secret", s.Reveal())
}

func TestGetEnv(t *testing.T) {
	t.Setenv("TEST_ENV", "value")

	require.Equal(t, "value", getEnv("TEST_ENV", "fallback"))
	require.Equal(t, "fallback", getEnv("DOES_NOT_EXIST", "fallback"))
}

func TestValidateSuccess(t *testing.T) {
	cfg := Config{
		DatabaseURL: "postgres://localhost/db",
	}

	require.NoError(t, cfg.validate())
}

func TestValidateMissingDatabaseURL(t *testing.T) {
	cfg := Config{}

	err := cfg.validate()

	require.Error(t, err)
	require.Contains(t, err.Error(), "DATABASE_URL")
}

func TestLoad(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("DATABASE_URL", "postgres://localhost/testdb")
	t.Setenv("PORT", "9000")
	t.Setenv("FRONTEND_URL", "http://localhost:3000")

	cfg, err := Load()

	require.NoError(t, err)

	require.Equal(t, AppEnv("test"), cfg.AppEnv)
	require.Equal(t, "9000", cfg.Port)
	require.Equal(t, "http://localhost:3000", cfg.FrontendURL)
	require.Equal(t, "postgres://localhost/testdb", cfg.DatabaseURL.Reveal())
}

func TestLoadMissingDatabaseURL(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("IP_STRATEGY", "remote_addr")

	_, err := Load()

	require.Error(t, err)
	require.Contains(t, err.Error(), "DATABASE_URL")
}
