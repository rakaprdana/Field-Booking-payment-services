package repositories

import (
	"context"
	errWrap "payment-services/common/errors"
	constErr "payment-services/constants/errors"
	"payment-services/domain/dto"
	"payment-services/domain/models"

	"gorm.io/gorm"
)

type PaymentHistoryRepository struct {
	db *gorm.DB
}

// Create implements [IPaymentHistoryRepository].
func (p *PaymentHistoryRepository) Create(
	ctx context.Context,
	tx *gorm.DB,
	req *dto.PaymentHistoryRequest,
) error {
	paymentHistory := models.PaymentHistory{
		PaymentID: req.PaymentID,
		Status:    req.Status,
	}

	err := tx.WithContext(ctx).
		Create(&paymentHistory).
		Error
	if err != nil {
		return errWrap.WrapError(constErr.ErrSQLError)
	}
	return nil
}

type IPaymentHistoryRepository interface {
	Create(context.Context, *gorm.DB, *dto.PaymentHistoryRequest) error
}

func NewPaymentHistoryRepository(
	db *gorm.DB,
) IPaymentHistoryRepository {
	return &PaymentHistoryRepository{db: db}
}
