package model

import "errors"

type Order struct {
	ID       int
	Customer string
	Status   string
}

var ErrEmptyCustomer = errors.New("имя клиента не задано")
var ErrNotFound = errors.New("заказ не найден")
