package shiprocket

import (
	"encoding/json"
	"fmt"
	"hermes/admin/services/auth"
	"hermes/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tryamigo/themis"
)

type ShipRocketPickUpAddress struct {
	PickupLocation string `json:"pickup_location,omitempty"` //required
	Name           string `json:"name,omitempty"`            //required
	Email          string `json:"email,omitempty"`           //required
	Phone          string `json:"phone,omitempty"`           //required
	Address        string `json:"address,omitempty"`         //required
	Address2       string `json:"address_2,omitempty"`
	City           string `json:"city,omitempty"`     //required
	State          string `json:"state,omitempty"`    //required
	Country        string `json:"country,omitempty"`  //required
	PinCode        string `json:"pin_code,omitempty"` //required
}

func CreateShipRocketPickUpAddressRequestBody(seller models.Seller) ShipRocketPickUpAddress {

	var shipRocketPickUpAddress ShipRocketPickUpAddress

	shipRocketPickUpAddress.PickupLocation = seller.ID
	shipRocketPickUpAddress.Name = seller.Name
	shipRocketPickUpAddress.Email = seller.Email
	shipRocketPickUpAddress.Phone = strings.Split(seller.Phone, "+91")[1]
	shipRocketPickUpAddress.Address = seller.HouseArea + "," + seller.StreetName
	shipRocketPickUpAddress.City = seller.City
	shipRocketPickUpAddress.State = seller.State
	shipRocketPickUpAddress.Country = "India"
	shipRocketPickUpAddress.PinCode = seller.PinCode

	return shipRocketPickUpAddress

}

func CreateShipRocketPickUpAddress(seller models.Seller) error {
	shipRocketPickUpAddress := CreateShipRocketPickUpAddressRequestBody(seller)
	requestBody, _ := json.Marshal(shipRocketPickUpAddress)

	token, err := GetShipRocketToken()
	if err != nil {
		return err
	}

	url := "https://apiv2.shiprocket.in/v1/external/settings/company/addpickup"
	headers := [][]string{{"Authorization", "Bearer " + token}}

	statusCode, responseBody, err := themis.HitAPIEndpoint2(url, http.MethodPost, requestBody, headers, [][]string{})
	if err != nil {
		return err
	}
	if statusCode >= 400 {
		return fmt.Errorf("erorr: %s with status code: %d", string(responseBody), statusCode)
	}
	return nil
}

func AddShiprocketPickUpAddress(c *gin.Context) {

	seller, err := auth.GetSellerFromSession(c)
	if err != nil {
		fmt.Println(fmt.Errorf("error: %s", err.Error()))
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	err = CreateShipRocketPickUpAddress(seller)
	if err != nil {
		fmt.Println(fmt.Errorf("error: %s", err.Error()))
		c.AbortWithError(http.StatusBadGateway, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": "Success"})
}
