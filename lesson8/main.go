package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
		writeJSON(w, r,
			http.StatusRequestEntityTooLarge,
			map[string]string{"error": "request body is too large."})
	} else {
		writeJSON(w, r,
			http.StatusBadRequest,
			map[string]string{"error": "invalid request body. expected json with the only field \"customer\" of type string."})
	}
}

func newRouter(store *Store) http.Handler {
	r := chi.NewRouter()
	r.Use(requestId, accessLog, recovery)
	r.Get("/health", health)
	r.Get("/orders/{id}", getOrder(store))
	r.Post("/orders", createOrder(store))
	return r
}

func health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, 200, map[string]string{"status": "ok"})
}

func createOrder(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Customer string `json:"customer"`
		}

		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/json" {
			writeJSON(w, r, http.StatusUnsupportedMediaType, map[string]string{"error": "Content-Type must be application/json."})
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1024)

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			handleError(w, r, err)
			return
		}

		if err := decoder.Decode(&struct{}{}); err != io.EOF {
			handleError(w, r, err)
			return
		}

		customer := strings.TrimSpace(input.Customer)
		if customer == "" {
			writeJSON(w, r, http.StatusBadRequest, map[string]string{"error": "customer must be non-empty."})
			return
		}

		order := store.Create(customer)
		w.Header().Set("Location", fmt.Sprintf("/orders/%d", order.ID))
		writeJSON(w, r, http.StatusCreated, order)
	}
}

func writeJSON(w http.ResponseWriter, r *http.Request, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		slog.Error("encode response",
			"error", err,
			"request_id", r.Context().Value(requestKey{}).(string),
		)
		status = http.StatusInternalServerError
		data = []byte(`{"error":"internal server error"}`)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(append(data, '\n')); err != nil {
		log.Printf("write response: %v", err)
	}
}

func getOrder(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil || id < 1 {
			writeJSON(w, r,
				http.StatusBadRequest,
				map[string]string{"error": "invalid order id (must be positive integer)."})
			return
		}

		order, ok := store.ByID(id)
		if !ok {
			writeJSON(w, r,
				http.StatusNotFound,
				map[string]string{"error": "order not found."})
			return
		}

		writeJSON(w, r, http.StatusOK, order)
	}
}

func main() {
	server := &http.Server{
		Addr: "127.0.0.1:8080", Handler: newRouter(NewStore()),
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       10 * time.Second, WriteTimeout: 10 * time.Second,
		IdleTimeout: 60 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
