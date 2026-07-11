package api

import (
	"context"
	"net/http"
	"time"
)

func (app *application) healthCheck(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	checkers := map[string]func(context.Context) error{
		"database": app.db.Ping,
	}

	checks := make(map[string]string, len(checkers))
	allOk := true

	for name, check := range checkers {

		if err := check(ctx); err != nil {
			app.logger.Error("healthcheck failed", "dependency", name, "error", err)
			checks[name] = "down"
			allOk = false
			continue
		} else {
			checks[name] = "ok"
		}
	}

	status := http.StatusOK
	overall := "available"

	if !allOk {
		status = http.StatusServiceUnavailable
		overall = "unavailable"
	}

	writeJSON(w, r, status, envelope{
		"data": map[string]any{
			"status":      overall,
			"environment": string(app.config.AppEnv),
			"checks":      checks,
		},
	})

}
