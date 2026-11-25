package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"example.com/lesson9practice/internal/model"
	"example.com/lesson9practice/internal/repository"
	"example.com/lesson9practice/internal/service"
)

type Orders interface {
	Create(string) (model.Order, error)
	ByID(int) (model.Order, error)
	List(options model.ListOptions) ([]model.Order, error)
}
type API struct {
	orders Orders
	logger *slog.Logger
}

func New(orders Orders, logger *slog.Logger) *API { return &API{orders: orders, logger: logger} }

// status godoc
// @Summary check if server is available (like "health")
// @Tags status
// @Produce  json
// @Success 200 {object} StatusResponse
// @Router /status [get]
func (a *API) status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, StatusResponse{Status: "ok"})
}

func (a *API) createOrder(w http.ResponseWriter, r *http.Request) {
	// Указатель позволяет отличить null от объекта. Теги DTO сами не валидируют запрос.
	var in *CreateOrderRequest
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, err.status, err.message, err.code)
		return
	}
	if in == nil {
		writeError(w, 400, "body must be a JSON object", "INVALID_JSON")
		return
	}
	order, err := a.orders.Create(in.Customer)
	if errors.Is(err, service.ErrEmptyCustomer) {
		writeError(w, 400, "customer is required", "VALIDATION_ERROR")
		return
	}
	if err != nil {
		a.internalError(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/v1/orders/"+strconv.Itoa(order.ID))
	writeJSON(w, 201, orderResponse(order))
}

// getOrder godoc
// @Summary get an order by id
// @Tags orders
// @Produce json
// @Param id path int true "order id"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /api/v1/orders/{id} [get]
func (a *API) getOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeError(w, 400, "id must be a positive integer", "INVALID_ID")
		return
	}
	order, err := a.orders.ByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, 404, "order not found", "ORDER_NOT_FOUND")
		return
	}
	if err != nil {
		a.internalError(w, r, err)
		return
	}
	writeJSON(w, 200, orderResponse(order))
}

func (a *API) internalError(w http.ResponseWriter, r *http.Request, err error) {
	a.logger.ErrorContext(r.Context(), "order operation failed", "error", err)
	writeError(w, 500, "internal server error", "INTERNAL_ERROR")
}

// listOrders godoc
// @Summary list of orders
// @Description list of orders sorted by id in ascending order
// @Tags orders
// @Produce json
// @Param limit query int false "page size" minimum(1) maximum(100) default(20)
// @Param offset query int false "offset" minimum(0) default(0)
// @Success 200 {array} OrderResponse
// @Failure 400 {object} ErrorResponse "INVALID_QUERY"
// @Failure 500 {object} ErrorResponse "INTERNAL_ERROR"
// @Failure 405 {object} ErrorResponse "METHOD_NOT_ALLOWED"
// @Router /api/v1/orders [get]
func (a *API) listOrders(w http.ResponseWriter, r *http.Request) {
	q, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		writeError(w, 400, "invalid query", "INVALID_QUERY")
		return
	}

	options, ok := parseListOptions(q)
	if !ok {
		writeError(w, 400, "invalid query", "INVALID_QUERY")
		return
	}

	list, err := a.orders.List(options)
	if err != nil {
		a.internalError(w, r, err)
		return
	}

	res := make([]OrderResponse, 0)
	for _, order := range list {
		res = append(res, orderResponse(order))
	}

	writeJSON(w, 200, res)
}

func parseListOptions(values url.Values) (model.ListOptions, bool) {
	options := model.ListOptions{Limit: 20}

	for key, items := range values {
		if len(items) != 1 {
			return options, false
		}

		value := items[0]
		number, err := strconv.Atoi(value)
		if err != nil {
			return options, false
		}

		for _, c := range value {
			if c < '0' || c > '9' {
				return options, false
			}
		}

		switch key {
		case "limit":
			if number < 1 || number > 100 {
				return options, false
			}
			options.Limit = number
		case "offset":
			options.Offset = number
		default:
			return options, false
		}
	}
	return options, true
}
