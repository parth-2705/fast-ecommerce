package controllers

import (
	"hermes/models"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

func InquiryWhatsapp(c *gin.Context) {
	productID := c.Query("productID")
	MessageWhatsapp(c, productID, "inquiry")
}

func OrderWhatsapp(c *gin.Context) {
	productID := c.Query("productID")
	MessageWhatsapp(c, productID, "order")
}

func LoginWhatsapp(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"response": "success"})
}

func MessageWhatsapp(c *gin.Context, productID string, method string) {
	product, err := models.GetCompleteProduct(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get product by ID"})
	}
	var text string
	if method == "order" {
		text = "Hey! I would like to order " + product.Name + " from " + product.Brand.Name + ".\nhttps://roovo.in/product/" + productID
	} else {
		text = "Hey! I am interested in " + product.Name + " from " + product.Brand.Name + ". Can you help me out with some questions?\nhttps://roovo.in/product/" + productID
	}
	redirect := "https://wa.me/+919910608373?text=" + url.QueryEscape(text)
	c.JSON(http.StatusOK, gin.H{"redirect": redirect})
}
