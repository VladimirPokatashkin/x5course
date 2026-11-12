package memory

import (
	"context"

	"example.com/order-service/internal/order"
)

type OrderRepository struct {
	items  []order.Order
	nextID int64
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		items:  make([]order.Order, 0),
		nextID: 1,
	}
}

func (r *OrderRepository) Create(_ context.Context, value order.Order) (order.Order, error) {
	value.ID = r.nextID
	r.nextID++
	r.items = append(r.items, value)
	return value, nil
}

func (r *OrderRepository) GetByID(_ context.Context, id int64) (order.Order, error) {
	if id < 1 || id >= r.nextID {
		return order.Order{}, order.ErrOrderNotFound
	}
	return r.items[id-1], nil
}

func (r *OrderRepository) List(_ context.Context) ([]order.Order, error) {
	val := make([]order.Order, len(r.items))
	copy(val, r.items)
	return val, nil
}

func (r *OrderRepository) Update(_ context.Context, value order.Order) (order.Order, error) {
	if value.ID < 1 || value.ID > int64(len(r.items)) {
		return order.Order{}, order.ErrOrderNotFound
	}
	r.items[value.ID-1] = value
	return value, nil
}
