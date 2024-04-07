package network

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func MobileRequest(c *gin.Context) bool {
	return strings.Contains(c.Request.Header.Get("Accept"), "application/json")
}