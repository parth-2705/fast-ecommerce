package controllers

import (
	"fmt"
	"hermes/db"
	"hermes/models"
	"hermes/utils/amplitude"
	"hermes/utils/data"
	"hermes/utils/network"
	"hermes/utils/tmpl"
	"net/http"

	"github.com/gin-gonic/gin"
)

func getAddressIDFromURL(c *gin.Context) (addressID string) {
	addressID = c.Param("addressID")
	fmt.Printf("addressID: %v\n", addressID)
	return
}

func getProductIdFromURL(c *gin.Context) (productID string) {
	productID = c.Param("productID")
	fmt.Printf("productID: %v\n", productID)
	return productID
}

func AddressGetHandler(c *gin.Context) {

	user, err := getUserObjectFromSession(c)
	if err != nil {
		c.AbortWithStatus(401)
		return
	}

	variantID := getVariantIDFromURL(c)
	if variantID == "" {
		c.AbortWithStatus(400)
		return
	}

	// fmt.Printf("producId: %v\n", producId)

	trackingMap := amplitude.GetTrackingMap(c)
	tmpl.TrackAmplitudeEvent("Address form opened", trackingMap)

	c.HTML(http.StatusOK, "address", gin.H{
		"username":   user.Name,
		"user_phone": user.Phone,
	})
}

func AddressPostHandler(c *gin.Context) {
	dealID := c.Query("deal")
	user, err := getUserObjectFromSession(c)
	if err != nil {
		c.AbortWithStatus(401)
		return
	}

	newAddress := models.Address{}
	err = c.Bind(&newAddress)
	if err != nil {
		fmt.Println(err)
		return
	}

	newAddress.ID = data.GetUUIDString("address")
	newAddress.UserID = user.ID

	err = newAddress.SaveToDB()
	if err != nil {
		c.AbortWithStatus(500)
		return
	}

	variantID := getVariantIDFromURL(c)
	if variantID == "" {
		c.AbortWithStatus(400)
		return
	}

	if len(dealID) == 0 {
		c.Redirect(http.StatusFound, fmt.Sprintf("/order/summary/%s?address=%s", variantID, newAddress.ID))
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/order/summary/%s?address=%s&deal=%s", variantID, newAddress.ID, dealID))
}

func AddressDeleteHandler(c *gin.Context) {
	user, err := getUserObjectFromSession(c)
	if err != nil {
		c.AbortWithStatus(401)
		return
	}
	addressId := c.Query("id")
	err = models.DeleteAddress(addressId, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to delete address " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "Successfully deleted address"})
}

func ManageAddresses(c *gin.Context) {
	user, err := getUserObjectFromSession(c)
	if err != nil {
		c.AbortWithStatus(401)
		return
	}
	defaultAddress := user.GetDefaultAddress()
	otherAddresses, err := user.GetNonDefaultAddresses()
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to get addresses for user"})
		return
	}
	if network.MobileRequest(c) {
		c.JSON(200, gin.H{"defaultAddress": defaultAddress, "addresses": otherAddresses})
		return
	}
	c.HTML(http.StatusOK, "manageAddresses", gin.H{"defaultAddress": defaultAddress, "addresses": otherAddresses})
}

func NewAddressPage(c *gin.Context) {
	user, err := getUserObjectFromSession(c)
	if err != nil {
		c.AbortWithStatus(401)
		return
	}
	redirect := c.Query("redirect")
	if redirect == "" {
		redirect = "addresses"
	}
	address := models.Address{}
	trackingMap := amplitude.GetTrackingMap(c)
	tmpl.TrackAmplitudeEvent("Address form opened", trackingMap)
	c.HTML(http.StatusOK, "enter-address", gin.H{"user": user, "address": address, "states": db.StateArr, "redirect": redirect, "new": true})
}

func EditAddressPage(c *gin.Context) {
	addressId := getAddressIDFromURL(c)
	user, err := getUserObjectFromSession(c)
	if err != nil {
		c.AbortWithStatus(401)
		return
	}
	redirect := c.Query("redirect")
	if redirect == "" {
		redirect = "addresses"
	}
	address, err := models.GetAddress(addressId)
	if err != nil {
		c.AbortWithStatus(404)
		return
	}

	c.HTML(http.StatusOK, "enter-address", gin.H{"user": user, "address": address, "redirect": redirect, "states": db.StateArr, "new": false})
}

func MakeNewAddress(c *gin.Context) {
	user, err := getUserObjectFromSession(c)
	if err != nil {
		c.AbortWithStatus(401)
		return
	}
	defaultAddress := user.GetDefaultAddress()

	newAddress := models.Address{}
	err = c.Bind(&newAddress)
	if err != nil {
		fmt.Println(err)
		return
	}

	newAddress.ID = data.GetUUIDString("address")
	newAddress.UserID = user.ID

	if defaultAddress.ID == "" {
		newAddress.IsDefault = true
	}

	err = newAddress.SaveToDB()
	if err != nil {
		c.AbortWithStatus(500)
		return
	}
	if newAddress.IsDefault {
		err = models.UpdateDefaultAddress(newAddress.ID, user.ID)
		if err != nil {
			c.AbortWithStatus(503)
			return
		}
	}

	setAddressIDInSession(c, newAddress.ID)

	c.JSON(http.StatusOK, newAddress)
}

func EditExistingAddress(c *gin.Context) {
	addressID := c.Query("id")

	oldAddress, err := models.GetAddress(addressID)
	if err != nil {
		c.AbortWithError(404, err)
		return
	}

	user, err := getUserObjectFromSession(c)
	if err != nil {
		c.AbortWithStatus(401)
		return
	}

	newAddress := oldAddress
	err = c.Bind(&newAddress)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = newAddress.UpdateInDB()
	if err != nil {
		c.AbortWithStatus(500)
		return
	}

	if newAddress.IsDefault {
		err = models.UpdateDefaultAddress(newAddress.ID, user.ID)
		if err != nil {
			c.AbortWithStatus(503)
			return
		}
	}

	setAddressIDInSession(c, newAddress.ID)

	c.JSON(http.StatusOK, newAddress)
}

func UpdateDefaultAddress(c *gin.Context) {
	addressID := c.Query("id")
	user, err := getUserObjectFromSession(c)
	if err != nil {
		c.AbortWithStatus(401)
		return
	}
	err = models.UpdateDefaultAddress(addressID, user.ID)
	if err != nil {
		c.AbortWithStatus(503)
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "Default Address Updated"})
}
