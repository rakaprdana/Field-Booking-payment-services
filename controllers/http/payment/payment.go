package controllers

import (
	"net/http"
	errValidate "payment-services/common/errors"
	response "payment-services/common/reponses"
	"payment-services/domain/dto"
	"payment-services/services"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type PaymentController struct {
	service services.IServiceRegistry
}

// Create implements [IPaymentController].
func (p *PaymentController) Create(ctx *gin.Context) {
	var req dto.PaymentRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  ctx,
		})
		return
	}

	validate := validator.New()
	if err = validate.Struct(req); err != nil {
		errMessage := http.StatusText(http.StatusUnprocessableEntity)
		errResponse := errValidate.ErrValidationResponse(err)
		response.HttpResponse(response.ParamHTTPResp{
			Code:    http.StatusBadRequest,
			Err:     err,
			Message: &errMessage,
			Data:    errResponse,
			Gin:     ctx,
		})
		return
	}
	result, err := p.service.GetPayment().Create(ctx, &req)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  ctx,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  ctx,
	})
}

// GetAllWithPagination implements [IPaymentController].
func (p *PaymentController) GetAllWithPagination(ctx *gin.Context) {
	var param dto.PaymentRequestParam
	err := ctx.ShouldBindQuery(&param)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  ctx,
		})
		return
	}

	validate := validator.New()
	if err = validate.Struct(param); err != nil {
		errMessage := http.StatusText(http.StatusUnprocessableEntity)
		errResponse := errValidate.ErrValidationResponse(err)
		response.HttpResponse(response.ParamHTTPResp{
			Code:    http.StatusBadRequest,
			Err:     err,
			Message: &errMessage,
			Data:    errResponse,
			Gin:     ctx,
		})
		return
	}
	result, err := p.service.GetPayment().GetAllWithPagination(ctx, &param)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  ctx,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  ctx,
	})
}

// GetByUUID implements [IPaymentController].
func (p *PaymentController) GetByUUID(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	result, err := p.service.GetPayment().GetByUUID(ctx, uuid)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  ctx,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  ctx,
	})
}

// Webhook implements [IPaymentController].
func (p *PaymentController) Webhook(ctx *gin.Context) {
	var req dto.Webhook

	err := ctx.ShouldBindBodyWithJSON(&req)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  ctx,
		})
		return
	}

	err = p.service.GetPayment().Webhook(ctx, &req)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  ctx,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Gin:  ctx,
	})
}

type IPaymentController interface {
	GetAllWithPagination(*gin.Context)
	GetByUUID(*gin.Context)
	Create(*gin.Context)
	Webhook(*gin.Context)
}

func NewPaymentController(service services.IServiceRegistry) IPaymentController {
	return &PaymentController{
		service: service,
	}
}
