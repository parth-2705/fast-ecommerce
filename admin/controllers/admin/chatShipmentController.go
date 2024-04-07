package controllers

import (
	"fmt"
	"hermes/models"
	"hermes/services/shiprocket"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type UpdateAddressChatRequestBody struct {
	Phone   string `json:"phone"`
	AWB     string `json:"awb"`
	Address string `json:"address"`
}

func IsPhoneNumberTheOrderNumberUser(phone string, awb string) error {

	shippings, err := models.GetShipmentsByAWB(awb)
	if err != nil {
		return err
	}

	for _, shipping := range shippings {
		order, err := models.GetOrder(shipping.Id)
		if err != nil {
			return err
		}

		// order.Address.Phone container phone number with +91 but phone only contains 91
		if !strings.Contains(order.Address.Phone, phone) {
			return fmt.Errorf("got %s", phone)
		}
	}

	return nil
}

func ReschduleShipment(c *gin.Context) {

	requestBody := make(map[string]string, 0)

	err := c.BindJSON(&requestBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid json " + err.Error()})
		return
	}

	if len(requestBody["phone"]) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone number is empty"})
		return
	}

	if len(requestBody["awb"]) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "AWB is empty"})
		return
	}

	if len(requestBody["date"]) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reschedule date is empty"})
		return
	}

	err = IsPhoneNumberTheOrderNumberUser(requestBody["phone"], requestBody["awb"])
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order Phone number and whatsapp phone number mismatch: " + err.Error()})
		return
	}

	err = shiprocket.ShiprocketReAttempt(requestBody["awb"], requestBody["date"])
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
}

func UpdateShipmentAddress(c *gin.Context) {

	var requestBody UpdateAddressChatRequestBody

	err := c.BindJSON(&requestBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid json " + err.Error()})
		return
	}

	if len(requestBody.Phone) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone number is empty"})
		return
	}

	if len(requestBody.AWB) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "AWB is empty"})
		return
	}

	if len(requestBody.Address) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "HouseArea is empty"})
		return
	}

	err = IsPhoneNumberTheOrderNumberUser(requestBody.Phone, requestBody.AWB)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order Phone number and whatsapp phone number mismatch: " + err.Error()})
		return
	}

	err = shiprocket.ShiprocketUpdateAddressOnReAttempt(requestBody.AWB, requestBody.Address)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
}

func CancelShipment(c *gin.Context) {

	requestBody := make(map[string]string, 0)

	err := c.BindJSON(&requestBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid json " + err.Error()})
		return
	}

	if len(requestBody["phone"]) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone number is empty"})
		return
	}

	if len(requestBody["awb"]) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "AWB is empty"})
		return
	}

	err = IsPhoneNumberTheOrderNumberUser(requestBody["phone"], requestBody["awb"])
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order Phone number and whatsapp phone number mismatch: " + err.Error()})
		return
	}

	err = shiprocket.ShiprocketReturnShipment(requestBody["awb"])
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
}

func ReportShipmentAsFakeAttempt(c *gin.Context) {

	requestBody := make(map[string]string, 0)

	err := c.BindJSON(&requestBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid json " + err.Error()})
		return
	}

	if len(requestBody["phone"]) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone number is empty"})
		return
	}

	if len(requestBody["awb"]) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "AWB is empty"})
		return
	}

	if len(requestBody["proof_image"]) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No proof image found"})
		return
	}

	err = IsPhoneNumberTheOrderNumberUser(requestBody["phone"], requestBody["awb"])
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order Phone number and whatsapp phone number mismatch: " + err.Error()})
		return
	}

	err = shiprocket.ShiprocketFakeAttempt(requestBody["awb"], requestBody["proof_image"])
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
}
