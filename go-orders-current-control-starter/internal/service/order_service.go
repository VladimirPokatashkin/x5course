package service

import (
	"context"
	"math"
	"time"

	"example.com/order-service/internal/order"
)

type OrderService struct {
	repository order.Repository
}

func NewOrderService(repository order.Repository) *OrderService {
	return &OrderService{repository: repository}
}

func (s *OrderService) Create(ctx context.Context, input order.CreateOrderInput) (order.Order, error) {
	if len(input.Items) == 0 {
		return order.Order{}, order.ErrInvalidOrder
	}

	value := order.Order{
		Items:      make([]order.Item, 0),
		History:    make([]order.StatusChange, 0),
		TotalPrice: 0,
	}

	for _, item := range input.Items {
		if item.Quantity < 1 || item.Name == "" || item.Price < 1 {
			return order.Order{}, order.ErrInvalidOrder
		}

		value.Items = append(value.Items, item)

		if item.Price > (math.MaxInt64-value.TotalPrice)/int64(item.Quantity) {
			return order.Order{}, order.ErrInvalidOrder
		}

		value.TotalPrice += item.Price * int64(item.Quantity)
	}

	value.CreatedAt = time.Now()
	value.Status = order.StatusNew
	value.History = append(value.History, order.StatusChange{
		Status:    order.StatusNew,
		ChangedAt: value.CreatedAt,
	})

	created, err := s.repository.Create(ctx, value)
	if err != nil {
		return order.Order{}, err
	}

	return created, nil
}

func (s *OrderService) GetByID(ctx context.Context, id int64) (order.Order, error) {
	value, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return order.Order{}, err
	}

	return value, nil
}

func (s *OrderService) List(ctx context.Context) ([]order.Order, error) {
	return s.repository.List(ctx)
}

func (s *OrderService) UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error) {
	value, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return order.Order{}, err
	}

	if err = validateStatus(status); err != nil {
		return order.Order{}, err
	}

	if err = validateStatusTransition(value.Status, status); err != nil {
		return order.Order{}, err
	}

	value.Status = status
	value.History = append(value.History, order.StatusChange{
		Status:    status,
		ChangedAt: time.Now(),
	})
	return s.repository.Update(ctx, value)
}

func (s *OrderService) History(ctx context.Context, id int64) ([]order.StatusChange, error) {
	value, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return []order.StatusChange{}, err
	}

	result := make([]order.StatusChange, len(value.History))
	copy(result, value.History)

	return result, nil
}

func validateStatus(status order.Status) error {
	if status == order.StatusNew || status == order.StatusCancelled || status == order.StatusCompleted || status == order.StatusProcessing {
		return nil
	}
	return order.ErrInvalidStatus
}

func validateStatusTransition(cur, next order.Status) error {
	if cur == order.StatusNew && next != order.StatusCancelled && next != order.StatusProcessing {
		return order.ErrInvalidStatusTransition
	}

	if cur == order.StatusProcessing && next != order.StatusCompleted && next != order.StatusCancelled {
		return order.ErrInvalidStatusTransition
	}

	if cur == order.StatusCompleted || cur == order.StatusCancelled {
		return order.ErrInvalidStatusTransition
	}

	return nil
}
