package errors

import (
	errorPayment "payment-services/constants/errors/payment"
)

func ErrMapping(err error) bool {
	var (
		GeneralErrors = GeneralErrors
		PaymentErrors = errorPayment.PaymentErrors
	)

	allErrors := make([]error, 0)
	allErrors = append(allErrors, GeneralErrors...)
	allErrors = append(allErrors, PaymentErrors...)

	for _, item := range allErrors {
		if err.Error() == item.Error() {
			return true
		}
	}

	return false
}
