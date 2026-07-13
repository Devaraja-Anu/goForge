package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
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
	DBMaxConns        int32
	DBMinConns        int32
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

	trustProxyHeaders, err := strconv.ParseBool(getEnv("TRUST_PROXY_HEADERS", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid TRUST_PROXY_HEADERS: %w", err)
	}

	cfg := Config{
		AppEnv:            AppEnv(envApp),
		DatabaseURL:       secret(os.Getenv("DATABASE_URL")),
		FrontendURL:       getEnv("FRONTEND_URL", "http://localhost:3000"),
		Port:              getEnv("PORT", "8080"),
		DBMaxConns:        getEnvInt("DB_MAX_CONNS", 25),
		DBMinConns:        getEnvInt("DB_MIN_CONNS", 5),
		TrustProxyHeaders: trustProxyHeaders,
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

	if c.DBMinConns > c.DBMaxConns {
		return fmt.Errorf("minimum DB connections cannot be greater than max connections")
	}

	if c.DBMinConns < 0 {
		return fmt.Errorf("minimum DB connections should not be less than 0")
	}

	if c.DBMaxConns <= 0 {
		return fmt.Errorf("max DB connections should be greater than 0")
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

func getEnvInt(key string, fallback int32) int32 {

	v := os.Getenv(key)
	if v == "" {
		return fallback
	}

	val, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return fallback
	}

	return int32(val)
}
