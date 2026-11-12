package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func requestLoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				status := wrapped.Status()
				if status == 0 {
					status = http.StatusOK
				}
				logger.InfoContext(
					r.Context(),
					"http request",
					"request_id", middleware.GetReqID(r.Context()),
					"method", r.Method,
					"path", r.URL.Path,
					"status", status,
					"duration", time.Since(start).Milliseconds(),
				)
			}()

			next.ServeHTTP(wrapped, r)
		})
	}
}

func recoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("recovered from panic", slog.Any("error", rec))
					writeJSON(w, http.StatusInternalServerError, errorResponse{
						Error: "internal server error",
						Code:  "INTERNAL_SERVER_ERROR",
					})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
