package common

import (
	"hermes/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func MarkShipmentItemAsDownloaded(c *gin.Context) {
	downloadedItem := c.Query("item")
	orderID := c.Query("orderID")

	if len(orderID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty order ID"})
		return
	}

	if len(downloadedItem) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty document name"})
		return
	}

	err := models.MarkAsTrue(orderID, downloadedItem)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "MarkAsTrue: " + err.Error()})
		return
	}

	shipment, err := models.GetShipmentByID(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "GetShipmentByID: " + err.Error()})
		return
	}

	isProcessed := false
	if shipment.LabelDownloaded && shipment.InvoiceDownaded && shipment.ManifestDownloaded {
		err = models.MarkAsProcessed(orderID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "MarkAsProcessed: " + err.Error()})
			return
		} else {
			isProcessed = true
		}
	}

	c.JSON(http.StatusOK, gin.H{"shipment": shipment, "processed": isProcessed})
}
