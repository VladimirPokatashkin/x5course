package service_test

import (
	"context"
	"testing"
	"time"

	"example.com/order-service/internal/order"
	service2 "example.com/order-service/internal/service"
	"github.com/stretchr/testify/assert"
)

type fakeRepository struct {
	orders map[int64]order.Order
	nextID int64
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		orders: make(map[int64]order.Order),
		nextID: 1,
	}
}

func (f *fakeRepository) Create(_ context.Context, value order.Order) (order.Order, error) {
	value.ID = f.nextID
	f.nextID++
	f.orders[value.ID] = value
	return value, nil
}

func (f *fakeRepository) GetByID(_ context.Context, id int64) (order.Order, error) {
	value, ok := f.orders[id]
	if !ok {
		return order.Order{}, order.ErrOrderNotFound
	}

	return value, nil
}

func (f *fakeRepository) List(_ context.Context) ([]order.Order, error) {
	val := make([]order.Order, 0, len(f.orders))

	for i := int64(1); i < f.nextID; i++ {
		if o, ok := f.orders[i]; ok {
			val = append(val, o)
		}
	}

	return val, nil
}

func (f *fakeRepository) Update(_ context.Context, value order.Order) (order.Order, error) {
	if _, ok := f.orders[value.ID]; !ok {
		return order.Order{}, order.ErrOrderNotFound
	}
	f.orders[value.ID] = value
	return value, nil
}

var validItems = []order.Item{
	{Name: "plis", Quantity: 12, Price: 120000},
	{Name: "oscilloscope", Quantity: 13, Price: 12000000},
	{Name: "kniga po schemotechnike", Quantity: 14, Price: 12000},
}

var itemsWithEmpty = []order.Item{
	{Name: "plis", Quantity: 12, Price: 120000},
	{Name: "", Quantity: 12, Price: 12000000},
	{Name: "kniga po schemotechnike", Quantity: 12, Price: 12000000},
}

var itemsWithNegativePrice = []order.Item{
	{Name: "plis", Quantity: 12, Price: 120000},
	{Name: "oscilloscope", Quantity: 12, Price: -12000000},
}

var itemsWithZeroPrice = []order.Item{
	{Name: "plis", Quantity: 12, Price: 120000},
	{Name: "oscilloscope", Quantity: 12, Price: 0},
}

var itemsWithNegativeQty = []order.Item{
	{Name: "plis", Quantity: 12, Price: 120000},
	{Name: "oscilloscope", Quantity: -12, Price: 1200000},
}

var itemsWithZeroQty = []order.Item{
	{Name: "plis", Quantity: 12, Price: 120000},
	{Name: "oscilloscope", Quantity: 0, Price: 12},
}

func TestCreateOrder(t *testing.T) {
	cases := []struct {
		name  string
		ctx   context.Context
		input order.CreateOrderInput
		price int64
		err   error
	}{
		{name: "valid", ctx: context.Background(), input: order.CreateOrderInput{Items: validItems}, price: 157608000, err: nil},
		{name: "empty", ctx: context.Background(), input: order.CreateOrderInput{Items: make([]order.Item, 0)}, price: 0, err: order.ErrInvalidOrder},
		{name: "with empty item", ctx: context.Background(), input: order.CreateOrderInput{Items: itemsWithEmpty}, price: 0, err: order.ErrInvalidOrder},
		{name: "with negative price", ctx: context.Background(), input: order.CreateOrderInput{Items: itemsWithNegativePrice}, price: 0, err: order.ErrInvalidOrder},
		{name: "with zero price", ctx: context.Background(), input: order.CreateOrderInput{Items: itemsWithZeroPrice}, price: 0, err: order.ErrInvalidOrder},
		{name: "with negative quantity", ctx: context.Background(), input: order.CreateOrderInput{Items: itemsWithNegativeQty}, price: 0, err: order.ErrInvalidOrder},
		{name: "with zero quantity", ctx: context.Background(), input: order.CreateOrderInput{Items: itemsWithZeroQty}, price: 0, err: order.ErrInvalidOrder},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var service = service2.NewOrderService(newFakeRepository())
			value, err := service.Create(tt.ctx, tt.input)

			if tt.err != nil {
				assert.ErrorIs(t, tt.err, err)
				assert.Equal(t, order.Order{}, value)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.price, value.TotalPrice)
			}
		})
	}
}

func TestGetByID(t *testing.T) {
	var service = service2.NewOrderService(newFakeRepository())
	service.Create(context.Background(), order.CreateOrderInput{Items: validItems})
	service.Create(context.Background(), order.CreateOrderInput{Items: validItems})

	cases := []struct {
		name          string
		ctx           context.Context
		id            int64
		expectedItems []order.Item
		expectedPrice int64
		err           error
	}{
		{"valid", context.Background(), 1, validItems, 157608000, nil},
		{"second valid", context.Background(), 2, validItems, 157608000, nil},
		{"missing order", context.Background(), 12, nil, 0, order.ErrOrderNotFound},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			value, err := service.GetByID(tt.ctx, tt.id)
			if tt.err != nil {
				assert.ErrorIs(t, tt.err, err)
				assert.Equal(t, order.Order{}, value)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPrice, value.TotalPrice)
				assert.Equal(t, tt.expectedItems, value.Items)
				assert.Equal(t, tt.id, value.ID)
			}
		})
	}
}

func TestList(t *testing.T) {
	var service = service2.NewOrderService(newFakeRepository())
	service.Create(context.Background(), order.CreateOrderInput{Items: validItems})
	service.Create(context.Background(), order.CreateOrderInput{Items: validItems})
	service.Create(context.Background(), order.CreateOrderInput{Items: validItems})

	cases := []struct {
		name        string
		ctx         context.Context
		expectedLen int
		err         error
	}{
		{"valid", context.Background(), 3, nil},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			value, err := service.List(tt.ctx)
			if tt.err != nil {
				assert.ErrorIs(t, tt.err, err)
				assert.Equal(t, tt.expectedLen, len(value))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLen, len(value))
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	cases := []struct {
		name   string
		ctx    context.Context
		id     int64
		status order.Status
		err    error
	}{
		{"valid", context.Background(), 1, order.StatusCancelled, nil},
		{"missing order", context.Background(), 12, order.StatusProcessing, order.ErrOrderNotFound},
		{"invalid status", context.Background(), 1, "schemotechnika", order.ErrInvalidStatus},
		{"invalid transition (into new)", context.Background(), 1, order.StatusNew, order.ErrInvalidStatusTransition},
		{"invalid transition (from canceled)", context.Background(), 3, order.StatusProcessing, order.ErrInvalidStatusTransition},
		{"invalid transition (from completed)", context.Background(), 2, order.StatusProcessing, order.ErrInvalidStatusTransition},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var service = service2.NewOrderService(newFakeRepository())
			service.Create(context.Background(), order.CreateOrderInput{Items: validItems})
			service.Create(context.Background(), order.CreateOrderInput{Items: validItems})
			service.Create(context.Background(), order.CreateOrderInput{Items: validItems})
			service.UpdateStatus(context.Background(), 2, order.StatusProcessing)
			service.UpdateStatus(context.Background(), 2, order.StatusCompleted)
			service.UpdateStatus(context.Background(), 3, order.StatusCancelled)

			value, err := service.UpdateStatus(tt.ctx, tt.id, tt.status)
			if tt.err != nil {
				assert.ErrorIs(t, err, tt.err)
				assert.Equal(t, order.Order{}, value)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.status, value.Status)

				value, _ := service.GetByID(tt.ctx, tt.id)

				assert.Equal(t, tt.status, value.Status)
			}
		})
	}
}

func TestHistory(t *testing.T) {
	var service = service2.NewOrderService(newFakeRepository())
	service.Create(context.Background(), order.CreateOrderInput{Items: validItems})
	service.UpdateStatus(context.Background(), 1, order.StatusProcessing)
	service.UpdateStatus(context.Background(), 1, order.StatusCompleted)

	cases := []struct {
		name   string
		ctx    context.Context
		id     int64
		expect []order.Status
		err    error
	}{
		{"valid", context.Background(), 1, []order.Status{order.StatusNew, order.StatusProcessing, order.StatusCompleted}, nil},
		{"missing order", context.Background(), 12, []order.Status{}, order.ErrOrderNotFound},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			value, err := service.History(tt.ctx, tt.id)
			if tt.err != nil {
				assert.ErrorIs(t, err, tt.err)
				assert.Equal(t, []order.StatusChange{}, value)
			} else {
				assert.NoError(t, err)

				statuses := make([]order.Status, len(value))
				dates := make([]time.Time, len(value))
				for i, v := range value {
					statuses[i] = v.Status

					if i > 0 && dates[i-1].After(v.ChangedAt) {
						t.Fatal("invalid status change date")
					}

					dates[i] = v.ChangedAt
				}

				assert.Equal(t, tt.expect, statuses)
			}
		})
	}
}
