package routes

import (
	"payment-services/clients"
	"payment-services/constants"
	controllers "payment-services/controllers/http"
	"payment-services/middlewares"

	"github.com/gin-gonic/gin"
)

type PaymentRoute struct {
	controller controllers.IControllerRegistry
	client     clients.IClientRegistry
	group      *gin.RouterGroup
}

// Run implements [IPaymentRoute].
func (p *PaymentRoute) Run() {
	group := p.group.Group("/payment")
	group.POST("/webhook", p.controller.GetPayment().Webhook)
	group.Use(middlewares.Authenticated())
	group.GET("", middlewares.CheckRole([]string{
		constants.Admin,
		constants.Customer,
	}, p.client), p.controller.GetPayment().GetAllWithPagination)
	group.GET("/:uuid", middlewares.CheckRole([]string{
		constants.Admin,
		constants.Customer,
	}, p.client), p.controller.GetPayment().GetByUUID)
	group.POST("", middlewares.CheckRole([]string{
		constants.Customer,
	}, p.client), p.controller.GetPayment().Create)
}

type IPaymentRoute interface {
	Run()
}

func NewPaymentRoute(
	controller controllers.IControllerRegistry,
	group *gin.RouterGroup,
	client clients.IClientRegistry,
) IPaymentRoute {
	return &PaymentRoute{
		controller: controller,
		group:      group,
		client:     client,
	}
}
