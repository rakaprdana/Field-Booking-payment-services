package repositories

import (
	repoPayment "payment-services/repositories/payment"
	repoPaymentHistory "payment-services/repositories/paymenthistory"

	"gorm.io/gorm"
)

type Registry struct {
	db *gorm.DB
}

// GetTx implements [IRepositoryRegistry].
func (r *Registry) GetTx() *gorm.DB {
	return r.db
}

// GetPayment implements [IRepositoryRegistry].
func (r *Registry) GetPayment() repoPayment.IPaymentRepository {
	return repoPayment.NewPaymentRepository(r.db)
}

// GetPaymentHistory implements [IRepositoryRegistry].
func (r *Registry) GetPaymentHistory() repoPaymentHistory.IPaymentHistoryRepository {
	return repoPaymentHistory.NewPaymentHistoryRepository(r.db)
}

type IRepositoryRegistry interface {
	GetPayment() repoPayment.IPaymentRepository
	GetPaymentHistory() repoPaymentHistory.IPaymentHistoryRepository
	GetTx() *gorm.DB
}

func NewRepositoryRegsitry(db *gorm.DB) IRepositoryRegistry {
	return &Registry{db: db}
}
