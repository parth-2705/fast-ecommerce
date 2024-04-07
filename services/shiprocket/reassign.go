package shiprocket

import (
	"encoding/json"
	"fmt"
	"hermes/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type ReturnOrderResponse struct {
	OrderID    int `json:"order_id" bson:"order_id"`
	ShipmentID int `json:"shipment_id" bson:"shipment_id"`
}

func ReassignShipment(c *gin.Context) {
	orderID := c.Query("id")

	if len(orderID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Empty order id",
		})
		return
	}

	shipment, err := models.GetShipmentByID(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	token, err := GetShipRocketToken()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unable to get access token " + err.Error(),
		})
		return
	}
	headers := [][]string{{"Authorization", "Bearer " + token}}

	awbCode, err := CreateReAssignShipRocketOrderShipment(fmt.Sprint(shipment.ShippingId), headers)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	shipment.AWB = awbCode

	labelURL, err := CreateShipRocketShipmentLabel(fmt.Sprint(shipment.ShippingId), headers)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	shipment.LabelURL = labelURL

	invoiceURL, err := CreateShipRocketShipmentInvoice(fmt.Sprint(shipment.OrderId), headers)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	shipment.InvoiceURL = invoiceURL

	shipRocketPickUpForShipment := CreateShipRocketPickUpRequestBody(shipment.ShippingId, "")
	requestBody, _ := json.Marshal(shipRocketPickUpForShipment)

	err = CreateShipRocketPickUpUtil(requestBody, headers)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	manifestUrl, err := CreateShipRocketShipmentManifest(fmt.Sprint(shipment.ShippingId), headers)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	shipment.ManifestURL = manifestUrl
	shipment.Dispatched = false
	shipment.Processed = false
	err = shipment.UpdateFields(bson.M{
		"awb":         shipment.AWB,
		"labelUrl":    shipment.LabelURL,
		"invoiceUrl":  shipment.InvoiceURL,
		"manifestUrl": shipment.ManifestURL,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
}
