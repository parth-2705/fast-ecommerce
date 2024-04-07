package shiprocket

import (
	"hermes/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func MarkDispatch(c *gin.Context) {
	shippingId := c.Query("id")
	if shippingId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "shipping id is missing"})
		return
	}
	err := models.MarkAsDispatched(shippingId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to mark as dispatched " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "success"})
}

func UnmarkDispatch(c *gin.Context) {
	shippingId := c.Query("id")
	if shippingId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "shipping id is missing"})
		return
	}
	err := models.UnmarkDispatch(shippingId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to unmark as dispatched " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "success"})
}

func MarkProcessed(c *gin.Context) {
	shippingId := c.Query("id")
	if shippingId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "shipping id is missing"})
		return
	}
	err := models.MarkAsProcessed(shippingId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to mark as processed " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "success"})
}

func UnmarkProcessed(c *gin.Context) {
	shippingId := c.Query("id")
	if shippingId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "shipping id is missing"})
		return
	}
	err := models.UnmarkProcessed(shippingId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to unmark as processed " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "success"})
}
