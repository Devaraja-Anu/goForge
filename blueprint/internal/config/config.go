package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type AppEnv string

type secret string

type Config struct {
	DatabaseURL       secret
	Port              string
	AppEnv            AppEnv
	FrontendURL       string
	TrustProxyHeaders bool
}

func (s secret) LogValue() slog.Value {
	return slog.StringValue("********")
}

func (s secret) String() string {
	return "********"
}

// config.go
func Load() (Config, error) {

	// APP_ENV is read before godotenv.Load() intentionally —
	// it must come from the real environment, never from a .env file.
	envApp := getEnv("APP_ENV", "development")

	if envApp != "production" && envApp != "test" {
		if err := godotenv.Load(); err != nil {
			slog.Warn("no .env file found, relying on environment variables")
		}
	}

	cfg := Config{
		AppEnv:            AppEnv(envApp),
		DatabaseURL:       secret(os.Getenv("DATABASE_URL")),
		FrontendURL:       getEnv("FRONTEND_URL", "http://localhost:3000"),
		Port:              getEnv("PORT", "8080"),
		TrustProxyHeaders: getEnv("TRUST_PROXY_HEADERS", "false") == "true",
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (s secret) Reveal() string {
	return string(s)
}

func (c Config) validate() error {
	var missing []string

	if c.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required env variables: %s", strings.Join(missing, ","))
	}

	return nil
}

// getEnv reads an environment variable, returning fallback if it's unset or empty.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
