package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"example.com/order-service/internal/order"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	swagger "github.com/swaggo/http-swagger"
)

type OrderService interface {
	Create(ctx context.Context, input order.CreateOrderInput) (order.Order, error)
	GetByID(ctx context.Context, id int64) (order.Order, error)
	List(ctx context.Context) ([]order.Order, error)
	UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error)
	History(ctx context.Context, id int64) ([]order.StatusChange, error)
}

type Handler struct {
	service OrderService
	logger  *slog.Logger
}

type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

type healthResponse struct {
	Status string `json:"status"`
}

func NewRouter(service OrderService, logger *slog.Logger) http.Handler {
	h := &Handler{service: service, logger: logger}
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(requestLoggingMiddleware(logger))
	router.Use(recoveryMiddleware(logger))

	router.Get("/health", h.health)
	router.Get("/swagger/*", swagger.WrapHandler)
	router.Route("/api/v1/orders", func(r chi.Router) {
		r.Post("/", h.createOrder)
		r.Get("/", h.listOrders)
		r.Get("/{id}", h.getOrderByID)
		r.Patch("/{id}/status", h.updateStatus)
		r.Get("/{id}/history", h.getStatusHistoryByID)
	})

	return router
}

// health godoc
// @Summary check if server is available
// @Description returns availability status
// @Tags health
// @Produce json
// @Success 200 {object} healthResponse
// @Router /health [get]
func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

// createOrder godoc
// @Summary order creation
// @Description creates new order with list of items and calculates the total price
// @Tags orders
// @Accept json
// @Produce json
// @Param input body order.CreateOrderInput true "items to include in the order"
// @Success 201 {object} order.Order
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/orders [post]
func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var input order.CreateOrderInput
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&input); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid JSON body",
			Code:  "INVALID_JSON",
		})
		return
	}

	created, err := h.service.Create(r.Context(), input)
	if errors.Is(err, order.ErrInvalidOrder) {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid order",
			Code:  "INVALID_ORDER",
		})
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "create order failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error: "internal server error",
			Code:  "INTERNAL_SERVER_ERROR",
		})
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

// listOrders godoc
// @Summary list of orders
// @Description returns list of all created orders
// @Tags orders
// @Produce json
// @Success 200 {array} order.Order
// @Failure 500 {object} errorResponse
// @Router /api/v1/orders [get]
func (h *Handler) listOrders(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.List(r.Context())
	if err != nil {
		h.logger.ErrorContext(r.Context(), "list orders failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error: "internal server error",
			Code:  "INTERNAL_SERVER_ERROR",
		})
		return
	}

	writeJSON(w, http.StatusOK, list)
}

// getOrderByID godoc
// @Summary get an order by id
// @Description returns order by its unique identifier
// @Tags orders
// @Produce json
// @Param id path int true "order ID"
// @Success 200 {object} order.Order
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/orders/{id} [get]
func (h *Handler) getOrderByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 1 {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid order ID",
			Code:  "INVALID_ORDER_ID",
		})
		return
	}

	value, err := h.service.GetByID(r.Context(), int64(id))
	if errors.Is(err, order.ErrOrderNotFound) {
		writeJSON(w, http.StatusNotFound, errorResponse{
			Error: "order not found",
			Code:  "ORDER_NOT_FOUND",
		})
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "get order by id failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error: "internal server error",
			Code:  "INTERNAL_SERVER_ERROR",
		})
		return
	}

	writeJSON(w, http.StatusOK, value)
}

// updateOrderStatus godoc
// @Summary update the status of order
// @Description updates order status according to lifecycle rules
// @Tags orders
// @Accept json
// @Produce json
// @Param id path int true "order ID"
// @Param input body order.UpdateStatusInput true "new order status"
// @Success 200 {object} order.Order
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 409 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/orders/{id}/status [patch]
func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 1 {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid order ID",
			Code:  "INVALID_ORDER_ID",
		})
		return
	}

	var input order.UpdateStatusInput
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&input); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid JSON body",
			Code:  "INVALID_JSON",
		})
		return
	}

	value, err := h.service.UpdateStatus(r.Context(), int64(id), input.Status)
	if errors.Is(err, order.ErrOrderNotFound) {
		writeJSON(w, http.StatusNotFound, errorResponse{
			Error: "order not found",
			Code:  "ORDER_NOT_FOUND",
		})
		return
	}
	if errors.Is(err, order.ErrInvalidStatusTransition) {
		writeJSON(w, http.StatusConflict, errorResponse{
			Error: "invalid status transition",
			Code:  "INVALID_STATUS_TRANSITION",
		})
		return
	}
	if errors.Is(err, order.ErrInvalidStatus) {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid status",
			Code:  "INVALID_STATUS",
		})
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "update status of order failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error: "internal server error",
			Code:  "INTERNAL_SERVER_ERROR",
		})
		return
	}

	writeJSON(w, http.StatusOK, value)
}

// getOrderHistory godoc
// @Summary get the history of order status changing
// @Description returns history of order status changing chronological
// @Tags orders
// @Produce json
// @Param id path int true "order ID"
// @Success 200 {array} order.StatusChange
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/orders/{id}/history [get]
func (h *Handler) getStatusHistoryByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 1 {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid order ID",
			Code:  "INVALID_ORDER_ID",
		})
		return
	}

	value, err := h.service.History(r.Context(), int64(id))
	if errors.Is(err, order.ErrOrderNotFound) {
		writeJSON(w, http.StatusNotFound, errorResponse{
			Error: "order not found",
			Code:  "ORDER_NOT_FOUND",
		})
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "get status history by id failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error: "internal server error",
			Code:  "INTERNAL_SERVER_ERROR",
		})
		return
	}

	writeJSON(w, http.StatusOK, value)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
