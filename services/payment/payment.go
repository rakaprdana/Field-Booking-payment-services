package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	clients "payment-services/clients/midtrans"
	midtransClient "payment-services/clients/midtrans"
	"payment-services/common/gcs"
	"payment-services/common/utils"
	"payment-services/config"
	"payment-services/constants"
	errPayment "payment-services/constants/errors/payment"
	"payment-services/controllers/kafka"
	"payment-services/domain/dto"
	"payment-services/domain/models"
	"payment-services/repositories"
	"strings"
	"time"

	"gorm.io/gorm"
)

type PaymentService struct {
	repository repositories.IRepositoryRegistry
	gcs        gcs.IGCSClient
	kafka      kafka.IKafkaRegistry
	midtrans   midtransClient.IMidtransClient
}

// Create implements [IPaymentService].
func (p *PaymentService) Create(ctx context.Context, req *dto.PaymentRequest) (*dto.PaymentResponse, error) {
	var (
		txErr, err error
		payment    *models.Payment
		response   *dto.PaymentResponse
		midtrans   *clients.MidtransData
	)

	err = p.repository.GetTx().Transaction(func(tx *gorm.DB) error {
		if !req.ExpiredAt.After(time.Now()) {
			return errPayment.ErrExpiredAtInvalid
		}
		midtrans, txErr = p.midtrans.CreatePaymentLink(req)
		if txErr != nil {
			return txErr
		}

		paymentReq := &dto.PaymentRequest{
			OrderID:     req.OrderID,
			Amount:      req.Amount,
			Description: req.Description,
			ExpiredAt:   req.ExpiredAt,
			PaymentLink: midtrans.RedirectURL,
		}
		payment, txErr = p.repository.GetPayment().Create(ctx, tx, paymentReq)
		if txErr != nil {
			return txErr
		}

		txErr = p.repository.GetPaymentHistory().Create(ctx, tx, &dto.PaymentHistoryRequest{
			PaymentID: payment.ID,
			Status:    payment.Status.GetStatusString(),
		})
		return nil
	})

	if err != nil {
		return nil, err
	}

	response = &dto.PaymentResponse{
		UUID:        payment.UUID,
		OrderID:     payment.OrderID,
		Amount:      payment.Amount,
		Status:      payment.Status.GetStatusString(),
		PaymentLink: payment.PaymentLink,
		Description: payment.Description,
	}

	return response, nil
}

// GetAllWithPagination implements [IPaymentService].
func (p *PaymentService) GetAllWithPagination(ctx context.Context, param *dto.PaymentRequestParam) (*utils.PaginationResult, error) {
	payments, total, err := p.repository.GetPayment().FindAllWithPagination(ctx, param)
	if err != nil {
		return nil, err
	}

	paymentResult := make([]dto.PaymentResponse, 0, len(payments))
	for _, payment := range payments {
		paymentResult = append(paymentResult, dto.PaymentResponse{
			UUID:          payment.UUID,
			TransactionID: payment.TransactionID,
			OrderID:       payment.OrderID,
			Amount:        payment.Amount,
			Status:        payment.Status.GetStatusString(),
			PaymentLink:   payment.PaymentLink,
			InvoiceLink:   payment.InvoiceLink,
			VANumber:      payment.VANumber,
			Bank:          payment.Bank,
			Description:   payment.Description,
			ExpiredAt:     payment.ExpiredAt,
			CreatedAt:     payment.CreatedAt,
			UpdatedAt:     payment.UpdatedAt,
		})
	}

	paginationParam := &utils.PaginationParams{
		Count: int(total),
		Page:  param.Page,
		Limit: param.Limit,
		Data:  paymentResult,
	}

	response := utils.GeneratePagination(*paginationParam)
	return &response, nil
}

// GetByUUID implements [IPaymentService].
func (p *PaymentService) GetByUUID(ctx context.Context, uuid string) (*dto.PaymentResponse, error) {
	payment, err := p.repository.GetPayment().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	paymentResult := dto.PaymentResponse{
		UUID:          payment.UUID,
		TransactionID: payment.TransactionID,
		OrderID:       payment.OrderID,
		Amount:        payment.Amount,
		Status:        payment.Status.GetStatusString(),
		PaymentLink:   payment.PaymentLink,
		InvoiceLink:   payment.InvoiceLink,
		VANumber:      payment.VANumber,
		Bank:          payment.Bank,
		Description:   payment.Description,
		ExpiredAt:     payment.ExpiredAt,
		CreatedAt:     payment.CreatedAt,
		UpdatedAt:     payment.UpdatedAt,
	}

	return &paymentResult, nil
}

// Webhook implements [IPaymentService].
func (p *PaymentService) Webhook(ctx context.Context, req *dto.Webhook) error {
	var (
		txErr, err         error
		paymentAfterUpdate *models.Payment
		paidAt             *time.Time
		invoiceLink        string
		pdf                []byte
	)

	err = p.repository.GetTx().Transaction(func(tx *gorm.DB) error {
		_, txErr = p.repository.GetPayment().FindByOrderID(ctx, req.OrderID.String())
		if txErr != nil {
			return txErr
		}

		if req.TransactionStatus == constants.SettlementString {
			now := time.Now()
			paidAt = &now
		}

		status := req.TransactionStatus.GetStatusInt()
		vaNumber := req.VANumbers[0].VaNumber
		bank := req.VANumbers[0].Bank
		_, txErr = p.repository.GetPayment().Update(ctx, tx, req.OrderID.String(), &dto.UpdatePaymentRequest{
			TransactionID: &req.TransactionID,
			Status:        &status,
			PaidAt:        paidAt,
			VANumber:      &vaNumber,
			Bank:          &bank,
			Acquirer:      req.Acquirer,
		})

		if txErr != nil {
			return txErr
		}

		paymentAfterUpdate, txErr = p.repository.GetPayment().FindByOrderID(ctx, req.OrderID.String())
		if txErr != nil {
			return txErr
		}

		txErr = p.repository.GetPaymentHistory().Create(ctx, tx, &dto.PaymentHistoryRequest{
			PaymentID: paymentAfterUpdate.ID,
			Status:    paymentAfterUpdate.Status.GetStatusString(),
		})

		if req.TransactionStatus == constants.SettlementString {
			paidDay := paidAt.Format("02")
			paidMonth := p.convertToIndonesianMonth(paidAt.Format("January"))
			paidYear := paidAt.Format("2026")
			invoiceNumber := fmt.Sprintf("INV/%s/ORD/%d", time.Now().Format(time.DateOnly), p.randomNumber())
			total := utils.RupiahFormarter(&paymentAfterUpdate.Amount)
			invoiceRequest := &dto.InvoiceRequest{
				InvoiceNumber: invoiceNumber,
				Data: dto.InvoiceData{
					PaymentDetail: dto.InvoicePaymentDetail{
						PaymentMethod: req.PaymentType,
						BankName:      string(*paymentAfterUpdate.Bank),
						VANumber:      *paymentAfterUpdate.VANumber,
						Date:          fmt.Sprintf("%s %s %s", paidDay, paidMonth, paidYear),
						IsPaid:        true,
					},
					Items: []dto.InvoiceItem{
						{
							Description: *paymentAfterUpdate.Description,
							Price:       total,
						},
					},
					Total: total,
				},
			}
			pdf, txErr = p.generatePDF(invoiceRequest)
			if txErr != nil {
				return txErr
			}

			_, txErr := p.uploadToGCS(ctx, invoiceNumber, pdf)
			if txErr != nil {
				return txErr
			}

			_, txErr = p.repository.GetPayment().Update(ctx, tx, req.OrderID.String(), &dto.UpdatePaymentRequest{
				InvoiceLink: &invoiceLink,
			})
			if txErr != nil {
				return txErr
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = p.produceToKafka(req, paymentAfterUpdate, paidAt)
	if err != nil {
		return err
	}

	return nil
}

type IPaymentService interface {
	GetAllWithPagination(context.Context, *dto.PaymentRequestParam) (*utils.PaginationResult, error)
	GetByUUID(context.Context, string) (*dto.PaymentResponse, error)
	Create(context.Context, *dto.PaymentRequest) (*dto.PaymentResponse, error)
	Webhook(context.Context, *dto.Webhook) error
}

func NewPaymentService(
	repository repositories.IRepositoryRegistry,
	gcs gcs.IGCSClient,
	kafka kafka.IKafkaRegistry,
	midtrans midtransClient.IMidtransClient,
) IPaymentService {
	return &PaymentService{
		repository: repository,
		gcs:        gcs,
		kafka:      kafka,
		midtrans:   midtrans,
	}
}

func (p *PaymentService) convertToIndonesianMonth(englishMonth string) string {
	montMap := map[string]string{
		"January":   "Januari",
		"February":  "Februari",
		"March":     "Maret",
		"April":     "April",
		"May":       "Mei",
		"June":      "Juni",
		"July":      "Juli",
		"August":    "Agustus",
		"September": "September",
		"October":   "Oktober",
		"November":  "November",
		"December":  "Desember",
	}

	indonesianMonth, ok := montMap[englishMonth]
	if !ok {
		return errors.New("Month not found").Error()
	}

	return indonesianMonth
}

func (p *PaymentService) generatePDF(req *dto.InvoiceRequest) ([]byte, error) {
	htmlTemplatePath := "/template/invoice.html"
	htmlTemplate, err := os.ReadFile(htmlTemplatePath)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	jsonData, _ := json.Marshal(req)
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return nil, err
	}

	pdf, err := utils.GeneratePDFFromHTML(string(htmlTemplate), data)
	if err != nil {
		return nil, err
	}

	return pdf, nil
}

func (p *PaymentService) uploadToGCS(
	ctx context.Context,
	invoiceNumber string,
	pdf []byte,
) (string, error) {
	invoiceNumberReplace := strings.ToLower(strings.ReplaceAll(invoiceNumber, "/", "-"))
	filename := fmt.Sprintf("%s.pdf", invoiceNumberReplace)
	url, err := p.gcs.UploadFile(ctx, filename, pdf)

	if err != nil {
		return "", err
	}
	return url, nil
}

func (p *PaymentService) randomNumber() int {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	number := random.Intn(900000) * 100000
	return number
}

func (p *PaymentService) mapTransactionStatusToEvent(
	status constants.PaymentStatusString,
) string {
	var paymentStatus string
	switch status {
	case constants.PendingString:
		paymentStatus = strings.ToUpper(constants.PendingString.String())
	case constants.SettlementString:
		paymentStatus = strings.ToUpper(constants.SettlementString.String())
	case constants.ExpireString:
		paymentStatus = strings.ToUpper(constants.ExpireString.String())
	}
	return paymentStatus
}

func (p *PaymentService) produceToKafka(
	req *dto.Webhook,
	payment *models.Payment,
	paidAt *time.Time,
) error {
	event := dto.KafkaEvent{
		Name: p.mapTransactionStatusToEvent(req.TransactionStatus),
	}

	metadata := dto.KafkaMetaData{
		Sender:    "payment-service",
		SendingAt: time.Now().Format(time.RFC3339),
	}
	body := dto.KafkaBody{
		Type: "JSON",
		Data: &dto.KafkaData{
			OrderID:   payment.OrderID,
			PaymentID: payment.UUID,
			Status:    req.TransactionStatus.String(),
			PaidAt:    paidAt,
			ExpiredAt: *payment.ExpiredAt,
		},
	}

	kafkaMessage := dto.KafkaMessage{
		Event:    event,
		Metadata: metadata,
		Body:     body,
	}

	topic := config.Config.Kafka.Topic
	kafkaMessageJSON, _ := json.Marshal(kafkaMessage)
	err := p.kafka.GetKafkaProducer().ProduceMessage(topic, kafkaMessageJSON)
	if err != nil {
		return err
	}

	return nil
}
