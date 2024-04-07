package shiprocket

import (
	"encoding/json"
	"fmt"
	"hermes/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tryamigo/themis"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetOrderShippingCharges(orderID string, shiprocketOrderID string) (models.ShippingCharges, error) {
	token, err := GetShipRocketToken()
	if err != nil {
		return models.ShippingCharges{}, err
	}

	url := "https://apiv2.shiprocket.in/v1/external/courier/serviceability/"
	headers := [][]string{{"Authorization", "Bearer " + token}}
	params := [][]string{{"order_id", shiprocketOrderID}}

	statusCode, responseBody, err := themis.HitAPIEndpoint2(url, http.MethodGet, []byte{}, headers, params)
	if err != nil {
		return models.ShippingCharges{}, err
	}
	if statusCode >= 400 {
		return models.ShippingCharges{}, fmt.Errorf("erorr: %s with status code: %d", string(responseBody), statusCode)
	}

	var shipRocketChargesResponse models.ShippingCharges
	err = json.Unmarshal(responseBody, &shipRocketChargesResponse)
	if err != nil {
		return models.ShippingCharges{}, err
	}

	if shipRocketChargesResponse.Status != 200 {
		return models.ShippingCharges{}, fmt.Errorf("erorr: %s with status code: %d", string(responseBody), shipRocketChargesResponse.Status)
	}

	shipRocketChargesResponse.OrderID = orderID
	shipRocketChargesResponse.ShiprocketOrderID = shiprocketOrderID
	err = shipRocketChargesResponse.Create()
	if err != nil {
		return models.ShippingCharges{}, err
	}

	return shipRocketChargesResponse, err
}

func GetShippingCharges(c *gin.Context) {
	orderID := c.Query("orderID")
	shiprocketOrderID := c.Query("shiprocketOrderID")

	if len(shiprocketOrderID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Shiprocket Order ID is empty",
		})
		return
	}

	shippingCharges, err := models.GetShippingChargesByOrderID(shiprocketOrderID)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Unable to get shipping charges from database" + err.Error(),
			})
			return
		} else {

			if len(orderID) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Order ID is empty",
				})
				return
			}

			shippingCharges, err = GetOrderShippingCharges(orderID, shiprocketOrderID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Unable to get shipping charges " + err.Error(),
				})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"response": shippingCharges,
	})
}
