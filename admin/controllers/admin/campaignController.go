package controllers

import (
	"fmt"
	"hermes/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetAllCampaigns(c *gin.Context) {
	campaigns, err := models.GetAllCampaigns()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Unable to get campaigns " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"campaigns": campaigns})
}

func GetCampaignByID(c *gin.Context) {
	campaignID := c.Query("id")
	fmt.Printf("campaignID: %+v\n", campaignID)
	if campaignID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty campaign ID"})
		return
	}
	campaign, err := models.GetCampaignByID(campaignID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Unable to get campaigns " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"campaign": campaign})
}

func CreateCampaign(c *gin.Context) {
	var campaign models.Campaign
	err := c.BindJSON(&campaign)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err})
		return
	}
	err = campaign.Create()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create campaign " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, campaign)
}

func UpdateCampaign(c *gin.Context) {
	var campaign models.Campaign
	err := c.BindJSON(&campaign)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	err = campaign.Update()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update campaign " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, campaign)
}
