package controllers

import (
	"fmt"
	"hermes/db"
	"hermes/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func InfluencerLandingPage(c *gin.Context) {
	influencer, _ := getInfluencerObjectFromSession(c)
	c.HTML(http.StatusOK, "influencerLanding", gin.H{"influencer": influencer})
}

func InfluencerAuthorizePage(c *gin.Context) {
	influencer, err := getInfluencerObjectFromSession(c)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/auth/sign-in-up?next=/influencer/authorize&back=/influencer")
		return
	}

	if influencer.HasMinimumConnections() {
		c.Redirect(http.StatusTemporaryRedirect, "/influencer/campaigns")
		return
	}

	c.HTML(http.StatusOK, "influencerAuthorize", gin.H{"influencer": influencer})
}

func InfluencerApprovalPage(c *gin.Context) {
	influencer, err := getInfluencerObjectFromSession(c)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/auth/sign-in-up?next=/influencer/authorize&back=/influencer")
		return
	}

	if influencer.HasMinimumConnections() {
		c.Redirect(http.StatusTemporaryRedirect, "/influencer/campaigns")
		return
	}

	c.HTML(http.StatusOK, "influencerApproval", gin.H{"influencer": influencer})
}

func AllInfluencerCampaigns(c *gin.Context) {
	campaigns, err := models.GetAllCampaigns()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "Error getting campaigns " + err.Error()})
		return
	}
	c.HTML(http.StatusOK, "influencerCampaignList", gin.H{"campaigns": campaigns})
}

func InfluencerPageValidation(c *gin.Context, campaignID string) (models.Influencer, string, int, error) {

	redirect := ""
	if campaignID == "" {
		return models.Influencer{}, redirect, http.StatusBadRequest, fmt.Errorf("empty Campaign ID")
	}

	influencer, err := getInfluencerObjectFromSession(c)
	if err != nil {
		redirect = "/auth/sign-in-up?next=influencer/campaign/" + campaignID
		return influencer, redirect, http.StatusUnauthorized, fmt.Errorf("unauthorized")
	}

	if !influencer.HasMinimumConnections() {
		redirect = "/influencer/authorize"
		return influencer, redirect, http.StatusTemporaryRedirect, fmt.Errorf("no authorized platforms found")
	}

	return influencer, redirect, http.StatusOK, nil
}

func InfluencerSpecificCampaign(c *gin.Context) {

	campaignID := c.Param("id")
	influencer, redirectURL, statusCode, err := InfluencerPageValidation(c, campaignID)
	if err != nil {
		if len(redirectURL) == 0 {
			c.JSON(statusCode, gin.H{"error": err.Error()})
		} else {
			c.Redirect(statusCode, redirectURL)
		}
	}

	campaign, err := models.GetCampaignByIDWithProductsAndBrand(campaignID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Error getting campaigns " + err.Error()})
		return
	}

	hasApplied, err := influencer.HasApplied(campaign.ID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Error getting application status " + err.Error()})
		return
	}

	c.HTML(http.StatusOK, "influencerCampaignSpecific", gin.H{"campaign": campaign, "influencer": influencer, "hasApplied": hasApplied})
}

func CampaignApplicationPlatformsPage(c *gin.Context) {
	campaignID := c.Param("id")
	influencer, redirectURL, statusCode, err := InfluencerPageValidation(c, campaignID)
	if err != nil {
		if len(redirectURL) == 0 {
			c.JSON(statusCode, gin.H{"error": err.Error()})
		} else {
			c.Redirect(statusCode, redirectURL)
		}
	}

	campaign, err := models.GetCampaignByIDWithProductsAndBrand(campaignID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Error getting campaigns " + err.Error()})
		return
	}

	application, err := campaign.CreateApplication(influencer.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error while creating application " + err.Error()})
		return
	}

	if application.SubmittedAt != nil {
		c.Redirect(http.StatusFound, fmt.Sprintf("/influencer/campaign/%s", campaignID))
		return
	}

	if influencer.IsConnected() {
		c.Redirect(http.StatusFound, fmt.Sprintf("/influencer/campaign/apply/%s/2", campaignID))
		return
	}

	c.HTML(http.StatusOK, "campaignApplicationPlatformsPage", gin.H{"campaign": campaign, "influencer": influencer})
}

func CampaignApplicationSelectProductsPage(c *gin.Context) {
	campaignID := c.Param("id")
	influencer, redirectURL, statusCode, err := InfluencerPageValidation(c, campaignID)
	if err != nil {
		if len(redirectURL) == 0 {
			c.JSON(statusCode, gin.H{"error": err.Error()})
		} else {
			c.Redirect(statusCode, redirectURL)
		}
	}

	application, err := models.GetApplicationWithFullCampaign(campaignID, influencer.ID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Error getting campaigns " + err.Error()})
		return
	}

	if application.SubmittedAt != nil {
		c.Redirect(http.StatusFound, fmt.Sprintf("/influencer/campaign/%s", campaignID))
		return
	}

	c.HTML(http.StatusOK, "campaignApplicationSelectProductsPage", gin.H{"campaign": application.Campaign, "influencer": influencer, "selectedProduct": application.ProductID})
}

func CampaignApplicationAddAddressPage(c *gin.Context) {
	campaignID := c.Param("id")
	influencer, redirectURL, statusCode, err := InfluencerPageValidation(c, campaignID)
	if err != nil {
		if len(redirectURL) == 0 {
			c.JSON(statusCode, gin.H{"error": err.Error()})
		} else {
			c.Redirect(statusCode, redirectURL)
		}
	}

	application, err := models.GetApplicationWithFullCampaign(campaignID, influencer.ID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Error getting campaigns " + err.Error()})
		return
	}

	if application.SubmittedAt != nil {
		c.Redirect(http.StatusFound, fmt.Sprintf("/influencer/campaign/%s", campaignID))
		return
	}

	user, err := getUserObjectFromSession(c)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/auth/sign-in-up?next=/influencer/authorize&back=/influencer/campaign/"+campaignID)
		return
	}

	address, err := user.FindDefaultAddress()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.HTML(http.StatusOK, "campaignApplicationAddAddressPage", gin.H{"campaign": application.Campaign, "influencer": influencer, "address": address, "selectedProduct": application.ProductID, "states": db.StateArr})
			return
		} else {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Error getting address " + err.Error()})
			return
		}
	}

	// Get all te addresses for this User
	addresses, err := user.GetAddresses()
	if err != nil {
		fmt.Println("err getting user addresses: ", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Error getting addresses " + err.Error()})
		return
	}

	c.HTML(http.StatusOK, "campaignApplicationSummaryPage", gin.H{"campaign": application.Campaign, "influencer": influencer, "address": address, "selectedProduct": application.ProductID, "addressOptions": addresses})
}

func CampaignApplicationTermsOfService(c *gin.Context) {
	campaignID := c.Param("id")
	influencer, redirectURL, statusCode, err := InfluencerPageValidation(c, campaignID)
	if err != nil {
		if len(redirectURL) == 0 {
			c.JSON(statusCode, gin.H{"error": err.Error()})
		} else {
			c.Redirect(statusCode, redirectURL)
		}
	}

	application, err := models.GetApplicationWithFullCampaign(campaignID, influencer.ID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Error getting campaigns " + err.Error()})
		return
	}

	if application.SubmittedAt != nil {
		c.Redirect(http.StatusFound, fmt.Sprintf("/influencer/campaign/%s", campaignID))
		return
	}

	c.HTML(http.StatusOK, "campaignApplicationTermsAndServices", gin.H{"campaignID": campaignID, "influencer": influencer})
}

func CampaignApplicationSuccessfulSubmission(c *gin.Context) {
	campaignID := c.Param("id")
	influencer, redirectURL, statusCode, err := InfluencerPageValidation(c, campaignID)
	if err != nil {
		if len(redirectURL) == 0 {
			c.JSON(statusCode, gin.H{"error": err.Error()})
		} else {
			c.Redirect(statusCode, redirectURL)
		}
	}

	campaign, err := models.GetCampaignByID(campaignID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Error getting campaigns " + err.Error()})
		return
	}

	c.HTML(http.StatusOK, "campaignApplicationSubmissionSuccess", gin.H{"campaignID": campaign.ID, "influencer": influencer})
}

func MyInfluencerCampaigns(c *gin.Context) {

	influencer, err := getInfluencerObjectFromSession(c)
	if err != nil {
		c.Redirect(http.StatusUnauthorized, "/auth/sign-in-up?next=influencer/")
		return
	}

	campaigns, err := influencer.GetMyCampaigns()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "Error getting campaigns " + err.Error()})
		return
	}
	c.HTML(http.StatusOK, "influencerCampaignList", gin.H{"campaigns": campaigns})
}
