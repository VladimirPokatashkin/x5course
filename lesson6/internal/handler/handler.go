package handler

import (
	"lessson6/internal/model"
)

type ServiceInterface interface {
	Create(customer string) (model.Order, error)
}
type Handler struct {
	orders ServiceInterface
}

func NewHandler(orders ServiceInterface) *Handler {
	return &Handler{orders: orders}
}

func (h *Handler) Create(customer string) (model.Order, error) {
	return h.orders.Create(customer)
}
