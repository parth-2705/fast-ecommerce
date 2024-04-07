package controllers

import (
	"hermes/models"
	"hermes/services/shiprocket"
	"hermes/utils/data"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetListOfSellers(c *gin.Context) {
	sellers, err := models.GetSellersList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sellers": sellers,
	})
}

func GetSellerByID(c *gin.Context) {
	sellerID := c.Param("sellerID")
	seller, err := models.GetSellerByID(sellerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"seller": seller,
	})
}

func CreateSeller(c *gin.Context) {
	newSeller := models.Seller{}
	err := c.BindJSON(&newSeller)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
		return
	}

	newSeller.Phone = "+91" + newSeller.Phone
	newSeller.ID = data.GetUUIDStringWithoutPrefix()

	err = shiprocket.CreateShipRocketPickUpAddress(newSeller)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to generate warehouse details due to " + err.Error()})
		return
	}

	newSeller.ProfileCompleted = true
	err = newSeller.Create()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to create seller " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "Successfully created seller", "seller": newSeller})
}
