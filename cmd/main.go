package cmd

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"payment-services/clients"
	midtransClient "payment-services/clients/midtrans"
	"payment-services/common/gcs"
	response "payment-services/common/reponses"
	"payment-services/config"
	"payment-services/constants"
	controller "payment-services/controllers/http"
	kafkaClient "payment-services/controllers/kafka"
	"payment-services/domain/models"
	"payment-services/middlewares"
	"payment-services/repositories"
	"payment-services/routes"
	"payment-services/services"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var command = &cobra.Command{
	Use:   "serve",
	Short: "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
		_ = godotenv.Load()
		config.Init()
		db, err := config.InitDatabase()
		if err != nil {
			panic(err)
		}

		loc, err := time.LoadLocation("Asia/Jakarta")
		if err != nil {
			panic(err)
		}
		time.Local = loc

		err = db.AutoMigrate(
			&models.Payment{},
			&models.PaymentHistory{},
		)

		if err != nil {
			panic(err)
		}

		// Initialize GCS client
		gcs := initGCS()
		// Initialize Kafka for broker message
		kafka := kafkaClient.NewKafkaRegistry(config.Config.Kafka.Brokers)
		// Initialize Midtrans for payment gateway
		midtrans := midtransClient.NewMidtransClient(
			config.Config.Midtrans.ServerKey,
			config.Config.Midtrans.IsProduction,
		)

		client := clients.NewClientRegistry()
		repository := repositories.NewRepositoryRegsitry(db)
		service := services.NewServiceRegistry(repository, gcs, kafka, midtrans)
		controller := controller.NewControllerRegistry(service)

		router := gin.Default()
		router.Use(middlewares.HandlePanic())
		router.NoRoute(func(ctx *gin.Context) {
			ctx.JSON(http.StatusNotFound, response.Response{
				Status:  constants.Error,
				Message: fmt.Sprintf("Path %s", http.StatusNotFound),
			})
		})

		router.GET("/", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, response.Response{
				Status:  constants.Success,
				Message: "Welcome to payment service",
			})
		})
		router.Use(func(ctx *gin.Context) {
			ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			ctx.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-service-name, x-api-key, x-request-at")
			ctx.Next()
		})
		lmt := tollbooth.NewLimiter(
			config.Config.RateLimiterMaxRequest,
			&limiter.ExpirableOptions{
				DefaultExpirationTTL: time.Duration(config.Config.RateLimiterTimeSecond) * time.Second,
			},
		)
		router.Use(middlewares.RateLimiter(lmt))

		group := router.Group("/api/v1")
		route := routes.NewRouteRegistry(controller, group, client)
		route.Serve()

		port := fmt.Sprintf(":%d", config.Config.Port)
		router.Run(port)
	},
}

func Run() {
	err := command.Execute()

	if err != nil {
		panic(err)
	}
}

func initGCS() gcs.IGCSClient {
	decode, err := base64.StdEncoding.DecodeString(config.Config.GCSPrivateKey)
	if err != nil {
		panic(err)
	}
	stringPrivateKey := string(decode)
	gcsServiceAccount := gcs.ServiceAccountKeyJSON{
		Type:                    config.Config.GCSType,
		ProjectID:               config.Config.GCSProjectID,
		PrivateKeyID:            config.Config.GCSProjectID,
		PrivateKey:              stringPrivateKey,
		ClientEmail:             config.Config.GCSClientEmail,
		ClientID:                config.Config.GCSClientID,
		AuthURI:                 config.Config.GCSAuthURI,
		TokenURI:                config.Config.GCSTokenURI,
		AuthProviderX509CertURL: config.Config.GCSAuthProviderX509CertURL,
		ClientX509CertURL:       config.Config.GCSClientX509CertURL,
		UniverseDomain:          config.Config.GCSUniverseDomain,
	}
	gcsClient := gcs.NewGCSClient(
		gcsServiceAccount,
		config.Config.GCSBucketName,
	)
	return gcsClient
}
