package middlewares

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"payment-services/clients"
	response "payment-services/common/reponses"
	"payment-services/config"
	"payment-services/constants"
	constErr "payment-services/constants/errors"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func HandlePanic() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorf("Recoverd from panic %v", r)
				ctx.JSON(http.StatusInternalServerError, response.Response{
					Status:  constants.Error,
					Message: constErr.ErrInternalServerError.Error(),
				})
				ctx.Abort()
			}
		}()
		ctx.Next()
	}
}

func RateLimiter(lmt *limiter.Limiter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		err := tollbooth.LimitByRequest(lmt, ctx.Writer, ctx.Request)
		if err != nil {
			ctx.JSON(http.StatusTooManyRequests, response.Response{
				Status:  constants.Error,
				Message: constErr.ErrTooManyRequests.Error(),
			})
			ctx.Abort()
		}
		ctx.Next()
	}
}

func responseUnauthorized(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusUnauthorized, response.Response{
		Status:  constants.Error,
		Message: message,
	})
	ctx.Abort()
}

func validateAPIKey(ctx *gin.Context) error {
	apiKey := ctx.GetHeader(constants.XApiKey)
	requestAt := ctx.GetHeader(constants.XRequestAt)
	serivceName := ctx.GetHeader(constants.XServiceName)
	signatureKey := config.Config.SignatureKey

	validateKey := fmt.Sprintf("%s:%s:%s", serivceName, signatureKey, requestAt)

	hash := sha256.New()
	hash.Write([]byte(validateKey))

	resulHash := hex.EncodeToString(hash.Sum(nil))
	if apiKey != resulHash {
		return constErr.ErrUnauthorized
	}

	return nil
}

func contains(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}

	return false
}

func CheckRole(roles []string, client clients.IClientRegistry) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, err := client.GetUser().GetUserByToken(ctx.Request.Context())
		if err != nil {
			responseUnauthorized(ctx, constErr.ErrUnauthorized.Error())
			return
		}

		if !contains(roles, user.Role) {
			responseUnauthorized(ctx, constErr.ErrUnauthorized.Error())
			return
		}
		ctx.Next()
	}
}

func Authenticated() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var err error
		token := ctx.GetHeader(constants.Authorization)
		if token == "" {
			responseUnauthorized(ctx, constErr.ErrUnauthorized.Error())
			return
		}

		err = validateAPIKey(ctx)
		if err != nil {
			responseUnauthorized(ctx, err.Error())
			return
		}
		cleanToken := strings.TrimSpace(strings.TrimPrefix(token, "Bearer"))
		newCtx := context.WithValue(ctx.Request.Context(), constants.Token, cleanToken)
		ctx.Request = ctx.Request.WithContext(newCtx)
		ctx.Next()
	}
}
