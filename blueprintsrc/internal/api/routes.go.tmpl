package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/devaraja-anu/blueprint/internal/logger"
)

func (app *application) routes() http.Handler {

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(recoverer)
	r.Use(logger.Middleware(app.logger))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{app.config.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(timeoutMiddleware(10 * time.Second))

	r.Get("/healthcheck", app.healthCheck)

	return r
}
