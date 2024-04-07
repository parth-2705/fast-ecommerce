package controllers

import (
	"fmt"
	"hermes/models"
	"hermes/utils/data"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateDefaultAddress(c *gin.Context) {

	phone := c.Param("phone")
	user, err := models.GetUser(phone)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	newAddress := models.Address{}
	err = c.Bind(&newAddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid body" + err.Error()})
		return
	}

	newAddress.ID = data.GetUUIDString("address")
	newAddress.UserID = user.ID

	fmt.Println("This is address:", newAddress)

	err = newAddress.SaveToDB()
	if err != nil {
		c.AbortWithStatus(500)
		return
	}
	c.JSON(http.StatusAccepted, newAddress)
}

func GetDefaultAddress(c *gin.Context) {
	phone := c.Param("phone")

	user, err := models.GetUser(phone)
	if err != nil {
		fmt.Printf("err get default address: %v\n", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	}
	address := user.GetDefaultAddress()
	c.JSON(http.StatusOK, address)
}
