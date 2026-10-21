package controllers

import (
	controllers "payment-services/controllers/http/payment"
	"payment-services/services"
)

type Registry struct {
	service services.IServiceRegistry
}

// GetPayment implements [IControllerRegistry].
func (r *Registry) GetPayment() controllers.IPaymentController {
	return controllers.NewPaymentController(r.service)
}

type IControllerRegistry interface {
	GetPayment() controllers.IPaymentController
}

func NewControllerRegistry(service services.IServiceRegistry) IControllerRegistry {
	return &Registry{service: service}
}
