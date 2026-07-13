package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/devaraja-anu/blueprint/internal/api"
	"github.com/devaraja-anu/blueprint/internal/config"
	"github.com/devaraja-anu/blueprint/internal/db"
	"github.com/devaraja-anu/blueprint/internal/db/database"
	"github.com/devaraja-anu/blueprint/internal/logger"
)

func main() {

	loggerInstance := logger.New()
	slog.SetDefault(loggerInstance)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL.Reveal(), cfg.DBMaxConns, cfg.DBMinConns)
	if err != nil {
		slog.Error("Unable to intialize DB", "error", err)
		os.Exit(1)
	}

	defer pool.Close()

	app := api.NewApplication(cfg, loggerInstance, pool, db.New(pool))

	if err := app.Serve(); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}

}
