package api

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/devaraja-anu/blueprint/internal/logger"
)

func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.FromContext(r.Context()).Error("panic recovered",
					"error", err,
					"path", r.URL.Path,
					"stack", string(debug.Stack()),
				)
				writeJSON(w, r, http.StatusInternalServerError, errorPayload{
					Message: "the server encountered an error and could not process your request", Code: "internal_error",
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func timeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, timeout, `{"error":{"code":"timeout","message":"Request timed out"}}`)
	}
}
