package controllers

import (
	"fmt"
	"hermes/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CamapignApplyRequest struct {
	InfluencerID string `json:"influencerID"`
	AddressID    string `json:"addressID"`
	CampaignID   string `json:"campaignID"`
	ProductID    string `json:"productID"`
}

func UpdateProductInInfluencerCampaignApplication(ctx *gin.Context) {
	var requestBody CamapignApplyRequest
	err := ctx.BindJSON(&requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body " + err.Error()})
		return
	}

	if len(requestBody.InfluencerID) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "empty influencer ID"})
		return
	}

	if len(requestBody.CampaignID) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "empty campaign ID"})
		return
	}

	if len(requestBody.ProductID) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "empty product ID"})
		return
	}

	application, err := models.GetInfluencerCampaignApplicaton(requestBody.InfluencerID, requestBody.CampaignID)
	if err != nil {
		fmt.Println(err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error while fetching the application " + err.Error()})
		return
	}

	err = application.UpdateProductSelection(requestBody.ProductID)
	if err != nil {
		fmt.Println(err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error while updating the application " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"application": application})
}

func UpdateAddressInInfluencerApplication(ctx *gin.Context) {
	var requestBody CamapignApplyRequest
	err := ctx.BindJSON(&requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body " + err.Error()})
		return
	}

	if len(requestBody.InfluencerID) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "empty influencer ID"})
		return
	}

	if len(requestBody.CampaignID) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "empty campaign ID"})
		return
	}

	if len(requestBody.AddressID) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "empty addressID ID"})
		return
	}

	application, err := models.GetInfluencerCampaignApplicaton(requestBody.InfluencerID, requestBody.CampaignID)
	if err != nil {
		fmt.Println(err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error while fetching the application " + err.Error()})
		return
	}

	address, err := models.GetAddress(requestBody.AddressID)
	if err != nil {
		fmt.Println(err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error while fetching the address " + err.Error()})
		return
	}

	err = application.UpdateAddressSelection(address)
	if err != nil {
		fmt.Println(err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error while updating the application address " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"application": application})
}

func SubmitInfluencerCampaignApplication(ctx *gin.Context) {
	var requestBody CamapignApplyRequest
	err := ctx.BindJSON(&requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body " + err.Error()})
		return
	}

	if len(requestBody.InfluencerID) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "empty influencer ID"})
		return
	}

	if len(requestBody.CampaignID) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "empty campaign ID"})
		return
	}

	application, err := models.GetInfluencerCampaignApplicaton(requestBody.InfluencerID, requestBody.CampaignID)
	if err != nil {
		fmt.Println(err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error while fetching the application " + err.Error()})
		return
	}

	cap, err := calculateCap(application.ProductID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	discount := models.CreateDiscount(10, cap, models.Percent)

	err = application.SubmitApplication(discount)
	if err != nil {
		fmt.Println(err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error while updating the application address " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"application": application})
}

func calculateCap(productID string) (cap float64, err error) {
	cap = 0
	var product models.Product
	product, err = models.GetProduct(productID)
	if err != nil {
		return
	}
	tempCap := product.Price.SellingPrice / 10
	if tempCap > cap {
		cap = tempCap
	}
	cap++
	return
}
