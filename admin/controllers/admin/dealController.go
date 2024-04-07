package controllers

import (
	"fmt"
	"hermes/models"
	"hermes/utils/data"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func AllDealsPage(c *gin.Context) {

	deals, err := models.GetAllDeals()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deals not found" + err.Error()})
	}
	c.HTML(http.StatusOK, "root", gin.H{"title": "Admin | Roovo", "deals": deals, "template": "deal"})
}

func NewDealPage(c *gin.Context) {
	var deal models.Deal
	deal.StartsAt = time.Now()
	deal.EndsAt = time.Now()
	products, err := models.GetAllProducts()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deals not found" + err.Error()})
	}

	c.HTML(http.StatusOK, "root", gin.H{"title": "Admin | Roovo", "deal": deal, "products": products, "template": "deal-edit", "new": true})
}

func DealEditPage(c *gin.Context) {
	dealID := c.Param("dealID")
	deal, err := models.GetDealByID(dealID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deal not found" + err.Error()})
	}

	products, err := models.GetAllProducts()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deals not found" + err.Error()})
	}

	c.HTML(http.StatusOK, "root", gin.H{"title": "Admin | Roovo", "deal": deal, "products": products, "template": "deal-edit", "new": false})
}

func DealCreate(c *gin.Context) {
	var deal models.Deal
	err := c.BindJSON(&deal)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
		return
	}
	deal.ID = data.GetUUIDString("deal")
	if time.Now().After(deal.StartsAt) {
		deal.Active = true
	}
	err = models.CreateDeal(deal)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to create deal"})
		return
	}
	c.JSON(http.StatusOK, deal)
}

func DealEdit(c *gin.Context) {
	var deal models.Deal
	dealID := c.Query("dealID")
	deal.ID = dealID
	err := c.BindJSON(&deal)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
		return
	}
	err = models.ReplaceDeal(deal)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to edit deal"})
		return
	}
	c.JSON(http.StatusOK, deal)
}

func SetDealInactive(c *gin.Context) {
	dealID := c.Query("dealID")
	err := models.DealActiveStatusUpdate(dealID, false)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to deactivate deal"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"dealStatus": "inactive"})
}
