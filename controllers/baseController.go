package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func BaseRoute(context *gin.Context) {
	context.HTML(http.StatusOK, "basePage", gin.H{})
}

func ProductRoute(context *gin.Context) {
	context.HTML(http.StatusOK, "productPage", gin.H{})
}

func TestRoute(context *gin.Context) {
	context.HTML(http.StatusOK, "testing", gin.H{})
}