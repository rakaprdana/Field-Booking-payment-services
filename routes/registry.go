package routes

import (
	client "payment-services/clients"
	controllers "payment-services/controllers/http"
	routes "payment-services/routes/payment"

	"github.com/gin-gonic/gin"
)

type Registry struct {
	controller controllers.IControllerRegistry
	group      *gin.RouterGroup
	client     client.IClientRegistry
}

// Serve implements [IRouteRegistry].
func (r *Registry) Serve() {
	r.PaymentRoute().Run()
}

type IRouteRegistry interface {
	Serve()
}

func NewRouteRegistry(
	controller controllers.IControllerRegistry,
	group *gin.RouterGroup,
	client client.IClientRegistry,
) IRouteRegistry {
	return &Registry{
		controller: controller,
		group:      group,
		client:     client,
	}
}

func (r *Registry) PaymentRoute() routes.IPaymentRoute {
	return routes.NewPaymentRoute(r.controller, r.group, r.client)
}
