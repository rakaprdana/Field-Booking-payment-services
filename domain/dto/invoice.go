package dto

type InvoiceRequest struct {
	InvoiceNumber string      `json:"invoiceNumber"`
	Data          InvoiceData `json:"data"`
}

type InvoiceData struct {
	PaymentDetail InvoicePaymenrDetail `json:"paymentDetail"`
	Items         []InvoiceItem        `json:"items"`
	Total         string               `json:"total"`
}

type InvoicePaymenrDetail struct {
	BankName      string `json:"bankName"`
	PaymentMethod string `json:"paymentMethod"`
	VANumber      string `json:"vaNumber"`
	IsPaid        bool   `json:"isPaid"`
}

type InvoiceItem struct {
	Description string `json:"description"`
	Price       string `json:"price"`
}
