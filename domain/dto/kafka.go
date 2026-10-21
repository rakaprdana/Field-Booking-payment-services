package dto

import (
	"time"

	"github.com/google/uuid"
)

type KafkaEvent struct {
	Name string `json:"name"`
}

type KafkaMetaData struct {
	Sender   string `json:"sender"`
	SenderAt string `json:"senderAt"`
}

type KafkaData struct {
	OrderID   uuid.UUID  `json:"orderID"`
	PaymentID uuid.UUID  `json:"paymentID"`
	Status    string     `json:"status"`
	ExpiredAt time.Time  `json:"expiredAt"`
	PaidAt    *time.Time `json:"paidAt"`
}

type KafkaBody struct {
	Type string     `json:"type"`
	Data *KafkaData `json:"data"`
}

type KafkaMessage struct {
	Event    KafkaEvent    `json:"event"`
	Metadata KafkaMetaData `json:"metadata"`
	Body     KafkaBody     `json:"body"`
}
