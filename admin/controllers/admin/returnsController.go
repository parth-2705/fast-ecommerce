package controllers

import (
	"hermes/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetAllReturns(c *gin.Context) {
	backShips, err := models.GetCompleteBackwardShipments("")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to get returns " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"returns": backShips})
}
