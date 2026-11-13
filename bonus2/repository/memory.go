package repository

import (
	"fmt"
	"maps"

	"example.com/delivery/service"
)

type Memory struct {
	pricers map[string]service.DeliveryPricer
}

func NewMemory(pricers map[string]service.DeliveryPricer) *Memory {
	copyOfPricers := make(map[string]service.DeliveryPricer, len(pricers))
	maps.Copy(copyOfPricers, pricers)
	return &Memory{pricers: copyOfPricers}
}

func Default() *Memory {
	return NewMemory(map[string]service.DeliveryPricer{
		"courier": service.NewCourier(),
		"pickup":  service.NewPickup(),
		"express": service.NewExpress(),
	})
}

func (m *Memory) Find(name string) (service.DeliveryPricer, error) {
	pricer, exists := m.pricers[name]

	if !exists {
		return nil, fmt.Errorf("%w", service.ErrUnknownTariff)
	}

	return pricer, nil
}
