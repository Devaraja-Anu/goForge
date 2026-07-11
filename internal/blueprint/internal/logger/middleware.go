package logger

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func Middleware(base *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			start := time.Now()
			requestID := middleware.GetReqID(r.Context())

			reqLogger := base.With(RequestAttrs(
				r.Method, r.URL.Path, requestID, r.RemoteAddr,
			)...)

			ctx := WithContext(r.Context(), reqLogger)

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r.WithContext(ctx))

			reqLogger.Info("request completed",
				"status", ww.Status(),
				LatencyAttr(start),
				"bytes", ww.BytesWritten(),
			)
		})
	}
}
