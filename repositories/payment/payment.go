package repositories

import (
	"errors"
	errWrap "payment-services/common/errors"
	"payment-services/constants"
	constErr "payment-services/constants/errors"
	constPaymentErr "payment-services/constants/errors/payment"

	"context"
	"fmt"
	"payment-services/domain/dto"
	"payment-services/domain/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentRepository struct {
	db *gorm.DB
}

// Create implements [IPaymentRepository].
func (p *PaymentRepository) Create(ctx context.Context,
	tx *gorm.DB,
	req *dto.PaymentRequest,
) (*models.Payment, error) {
	status := constants.Initial
	orderID := uuid.MustParse(req.OrderID)
	payment := models.Payment{
		UUID:        uuid.New(),
		OrderID:     orderID,
		Amount:      req.Amount,
		PaymentLink: req.PaymentLink,
		ExpiredAt:   &req.ExpiredAt,
		Description: &req.Description,
		Status:      &status,
	}

	err := tx.WithContext(ctx).
		Create(&payment).
		Error
	if err != nil {
		return nil, errWrap.WrapError(constErr.ErrSQLError)
	}
	return &payment, nil
}

// FindAllWithPagination implements [IPaymentRepository].
func (p *PaymentRepository) FindAllWithPagination(ctx context.Context, param *dto.PaymentRequestParam) ([]models.Payment, int64, error) {
	var (
		payments []models.Payment
		sort     string
		total    int64
	)
	order := "desc"
	if param.SortOrder != nil && *param.SortOrder != "" {
		order = *param.SortOrder
	}

	if param.SortColumn != nil && *param.SortColumn != "" {
		sort = fmt.Sprintf("%s %s", *param.SortColumn, order)
	} else {
		sort = "created_at desc"
	}

	limit := param.Limit
	offset := (param.Page - 1) * limit
	err := p.db.
		WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order(sort).
		Find(&payments).
		Error
	if err != nil {
		return nil, 0, errWrap.WrapError(constErr.ErrSQLError)
	}

	err = p.db.
		WithContext(ctx).
		Model(&models.Payment{}).
		Count(&total).
		Error
	if err != nil {
		return nil, 0, errWrap.WrapError(constErr.ErrSQLError)
	}
	return payments, total, nil

}

// FindByOrderID implements [IPaymentRepository].
func (p *PaymentRepository) FindByOrderID(ctx context.Context, orderID string) (*models.Payment, error) {
	var payment models.Payment
	err := p.db.
		WithContext(ctx).
		Where("order_id = ?", orderID).
		First(&payment).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errWrap.WrapError(constPaymentErr.ErrPaymentNotFound)
		}
		return nil, errWrap.WrapError(constErr.ErrSQLError)
	}
	return &payment, nil
}

// FindByUUID implements [IPaymentRepository].
func (p *PaymentRepository) FindByUUID(ctx context.Context, uuid string) (*models.Payment, error) {
	var payment models.Payment
	err := p.db.
		WithContext(ctx).
		Where("uuid = ?", uuid).
		First(&payment).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errWrap.WrapError(constPaymentErr.ErrPaymentNotFound)
		}
		return nil, errWrap.WrapError(constErr.ErrSQLError)
	}
	return &payment, nil
}

// Update implements [IPaymentRepository].
func (p *PaymentRepository) Update(ctx context.Context, tx *gorm.DB, orderID string, req *dto.UpdatePaymentRequest) (*models.Payment, error) {
	payment := models.Payment{
		Status:        req.Status,
		TransactionID: req.TransactionID,
		InvoiceLink:   req.InvoiceLink,
		PaidAt:        req.PaidAt,
		VANumber:      req.VANumber,
		Bank:          req.Bank,
		Acquirer:      req.Acquirer,
	}

	err := tx.WithContext(ctx).
		Where("order_id = ?", orderID).
		Updates(&payment).
		Error

	if err != nil {
		return nil, errWrap.WrapError(constErr.ErrSQLError)
	}
	return &payment, nil
}

type IPaymentRepository interface {
	FindAllWithPagination(context.Context, *dto.PaymentRequestParam) ([]models.Payment, int64, error)
	FindByUUID(context.Context, string) (*models.Payment, error)
	FindByOrderID(context.Context, string) (*models.Payment, error)
	Create(context.Context, *gorm.DB, *dto.PaymentRequest) (*models.Payment, error)
	Update(context.Context, *gorm.DB, string, *dto.UpdatePaymentRequest) (*models.Payment, error)
}

func NewPaymentRepository(db *gorm.DB) IPaymentRepository {
	return &PaymentRepository{
		db: db,
	}
}
