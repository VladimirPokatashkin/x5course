package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type requestKey struct{}

var requestCounter atomic.Uint64

func requestId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := fmt.Sprintf("request%d", requestCounter.Add(1))
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), requestKey{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			status := wrapped.Status()
			if status == 0 {
				status = http.StatusOK
			}
			slog.Info(
				"http request",
				"request_id", r.Context().Value(requestKey{}).(string),
				"method", r.Method,
				"path", r.URL.Path,
				"status", status,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		}()
		next.ServeHTTP(wrapped, r)
	})
}

func recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				writeJSON(w, r, http.StatusInternalServerError, map[string]string{"error": "internal server error."})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
