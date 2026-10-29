package handler

import (
	"lessson6/internal/model"
	"lessson6/internal/service"
)

type Handler struct {
	orders *service.Service
}

func NewHandler(orders *service.Service) *Handler {
	return &Handler{orders: orders}
}

func (h *Handler) Create(customer string) (model.Order, error) {
	return h.orders.Create(customer)
}
