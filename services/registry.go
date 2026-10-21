package services

import (
	midtransClient "payment-services/clients/midtrans"
	"payment-services/common/gcs"
	"payment-services/controllers/kafka"
	"payment-services/repositories"
	servicePayment "payment-services/services/payment"
)

type Registry struct {
	repository repositories.IRepositoryRegistry
	gcs        gcs.IGCSClient
	kafka      kafka.IKafkaRegistry
	midtrans   midtransClient.IMidtransClient
}

// GetPayment implements [IServiceRegistry].
func (r *Registry) GetPayment() servicePayment.IPaymentService {
	return servicePayment.NewPaymentService(
		r.repository,
		r.gcs,
		r.kafka,
		r.midtrans,
	)
}

type IServiceRegistry interface {
	GetPayment() servicePayment.IPaymentService
}

func NewServiceRegistry(
	repository repositories.IRepositoryRegistry,
	gcs gcs.IGCSClient,
	kafka kafka.IKafkaRegistry,
	midtrans midtransClient.IMidtransClient,
) IServiceRegistry {
	return &Registry{
		repository: repository,
		gcs:        gcs,
		kafka:      kafka,
		midtrans:   midtrans,
	}
}
