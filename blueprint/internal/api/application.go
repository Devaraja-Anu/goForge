package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/devaraja-anu/blueprint/internal/config"
	"github.com/devaraja-anu/blueprint/internal/db"
)

type application struct {
	config   config.Config
	logger   *slog.Logger
	db       *pgxpool.Pool
	wg       sync.WaitGroup
	queries  *db.Queries
	validate *validator.Validate
}

func NewApplication(cfg config.Config, log *slog.Logger, database *pgxpool.Pool,
	queries *db.Queries,
) *application {

	v := validator.New()

	// Use JSON tag names in validation errors instead of struct field names.
	// Without this, errors reference "DisplayName" instead of "display_name".
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		tag := f.Tag.Get("json")
		if tag == "" || tag == "-" {
			return f.Name
		}
		return strings.Split(tag, ",")[0]
	})

	return &application{
		config:   cfg,
		logger:   log,
		db:       database,
		validate: v,
		queries:  queries,
	}
}

func (app *application) Serve() error {

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", app.config.Port),
		Handler:           app.routes(),
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	shutdownError := make(chan error, 1)

	go func() {

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(quit)
		s := <-quit

		app.logger.Info("shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.logger.Info("completing background tasks", "address", srv.Addr)
		app.wg.Wait()
		shutdownError <- nil
	}()

	app.logger.Info("starting server", "addr", srv.Addr)

	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	if err := <-shutdownError; err != nil {
		return err
	}

	app.logger.Info("server stopped", "addr", srv.Addr)
	return nil
}
