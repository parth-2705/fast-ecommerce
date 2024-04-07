package middleware

import (
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

func AttachErrorAlertMiddleware() gin.HandlerFunc {
	return (sentrygin.New(sentrygin.Options{
		Repanic: true,
	}))
}
