package controllers

import (
	"hermes/services/Sentry"

	"github.com/gin-gonic/gin"
)

func JSONErrorResponse(c *gin.Context, statusCode int, err error) {
	Sentry.SendErrorToSentry(c, err, nil)
	c.JSON(statusCode, gin.H{
		"error": err.Error(),
	})
}
