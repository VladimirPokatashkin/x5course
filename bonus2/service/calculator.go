package service

import (
	"fmt"

	"example.com/delivery/model"
)

type TariffRepository interface {
	Find(name string) (DeliveryPricer, error)
}

type Calculator struct {
	repository TariffRepository
}

func NewCalculator(repository TariffRepository) *Calculator {
	return &Calculator{repository: repository}
}

func (c *Calculator) Calculate(order model.Order) (int64, error) {
	tariff, err := c.repository.Find(order.Tariff)
	if err != nil {
		return 0, &OrderPriceError{
			OrderID: order.ID,
			Tariff:  order.Tariff,
			Err:     fmt.Errorf("could not find tariff: %w", err),
		}
	}

	price, err := tariff.Price(order.WeightGrams, order.DistanceKM)
	if err != nil {
		return 0, &OrderPriceError{
			OrderID: order.ID,
			Tariff:  order.Tariff,
			Err:     fmt.Errorf("invalid order: %w", err),
		}
	}

	return price, nil
}
