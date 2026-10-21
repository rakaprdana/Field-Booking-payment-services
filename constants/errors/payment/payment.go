package error

import "errors"

var (
	ErrPaymentNotFound  = errors.New("Payment not found")
	ErrExpiredAtInvalid = errors.New("Expired time must be greater than current time ")
)

var PaymentErrors = []error{
	ErrPaymentNotFound,
	ErrExpiredAtInvalid,
}
