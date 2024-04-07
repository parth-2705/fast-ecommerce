package controllers

import (
	"hermes/admin/services/auth"
	"hermes/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetAllReturns(c *gin.Context) {
	sellerID, err := auth.GetSellerIdFromSession(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get seller ID from session " + err.Error()})
		return
	}
	backShips, err := models.GetCompleteBackwardShipments(sellerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to get returns " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"returns": backShips})
}
