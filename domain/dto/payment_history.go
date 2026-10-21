package dto

import "payment-services/constants"

type PaymentHistoryRequest struct {
	PaymentID uint                          `json:"paymentID"`
	Status    constants.PaymentStatusString `json:"status"`
}
