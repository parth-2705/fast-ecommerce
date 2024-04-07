package controllers

import (
	"encoding/json"
	"fmt"
	"hermes/controllers"
	"hermes/models"
	"hermes/utils/whatsapp"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/tryamigo/themis"
	"go.mongodb.org/mongo-driver/bson"
)

func GetInfluencersList(c *gin.Context) {
	limitInt := controllers.GetLimitFromQueryValue(c.Query("limit"))
	pageInt := controllers.GetPageFromQueryValue(c.Query("page"))
	filter := c.Query("filter")

	if len(filter) == 0 {
		filter = "pending"
	}

	var Paginater controllers.Pagination = controllers.Pagination{
		Limit: limitInt,
		Page:  pageInt,
	}

	var influencerStruct []*models.Influencer
	preSkip := bson.A{
		bson.D{{Key: "$sort", Value: bson.D{{Key: "instagram.createdAt", Value: -1}}}},
		bson.D{{Key: "$match", Value: bson.D{{
			Key: "$or",
			Value: bson.A{
				bson.D{{Key: "instagram.isConnected", Value: true}},
				bson.D{{Key: "youtube.isConnected", Value: true}},
				bson.D{{Key: "snapchat.isConnected", Value: true}},
			},
		}}}},
	}

	switch filter {
	case "pending":
		preSkip = append(preSkip, bson.D{{Key: "$match", Value: bson.D{{
			Key: "$or",
			Value: bson.A{
				bson.D{{Key: "instagram.isVerified", Value: false}},
				bson.D{{Key: "instagram.isVerified", Value: bson.D{{Key: "$exists", Value: false}}}},
				bson.D{{Key: "youtube.isVerified", Value: false}},
				bson.D{{Key: "youtube.isVerified", Value: bson.D{{Key: "$exists", Value: false}}}},
				bson.D{{Key: "snapchat.isVerified", Value: false}},
				bson.D{{Key: "snapchat.isVerified", Value: bson.D{{Key: "$exists", Value: false}}}},
			},
		}}}})
	case "approved":
		preSkip = append(preSkip, bson.D{{Key: "$match", Value: bson.D{{
			Key: "$or",
			Value: bson.A{
				bson.D{{Key: "instagram.approved", Value: true}},
				bson.D{{Key: "youtube.approved", Value: true}},
				bson.D{{Key: "snapchat.approved", Value: true}},
			},
		}}}})
	case "disapproved":
		preSkip = append(preSkip, bson.D{{Key: "$match", Value: bson.D{{
			Key: "$or",
			Value: bson.A{
				bson.D{{Key: "instagram.isVerified", Value: true}},
				bson.D{{Key: "instagram.approved", Value: false}},
				bson.D{{Key: "youtube.isVerified", Value: true}},
				bson.D{{Key: "youtube.approved", Value: false}},
				bson.D{{Key: "snapchat.isVerified", Value: true}},
				bson.D{{Key: "snapchat.approved", Value: false}},
			},
		}}}})
	case "all":
	default:
		// skip
	}

	influencer, err := controllers.Paginate("influencer", &Paginater, influencerStruct, preSkip, bson.A{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"influencer": influencer.Rows,
		"totalRows":  influencer.TotalRows,
		"totalPages": influencer.TotalPages,
		"filter":     filter,
	})
}

func UpdateInstagramProfile(c *gin.Context) {

	influencerID := c.Query("influencerID")

	if len(influencerID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty influencer ID"})
		return
	}

	influencer, err := models.GetInfluencerByID(influencerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "state": "get influencer by id"})
		return
	}

	// Get access token from vault
	vaultHandler := themis.NewVaultHandler(os.Getenv("VAULT_URL"), os.Getenv("VAULT_TOKEN"), influencer.ID+"/instagram")
	statusCode, respBody, err := vaultHandler.GetSecretsFromVault()
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error(), "state": "get secrets"})
		return
	}

	var instagramSecret map[string]string
	err = json.Unmarshal(respBody, &instagramSecret)
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error(), "state": "unmarshal secrets"})
		return
	}

	access_token := instagramSecret["access_token"]
	instagramProfile, err := models.GetInstagramUserProfile(access_token, influencer.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "state": "instagram profile"})
		return
	}

	// don't change the createdAt time while updating the profile
	creationTime := influencer.Instagram.CreatedAt
	influencer.Instagram = instagramProfile
	influencer.Instagram.CreatedAt = creationTime

	// update the instagram profile in db
	err = influencer.Update()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "state": "influencer update"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": "Success"})
}

func ApproveInfluencer(c *gin.Context) {

	influencerID := c.Query("influencerID")
	platform := c.Query("platform")

	if len(influencerID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty influencer ID"})
		return
	}

	influencer, err := models.GetInfluencerByID(influencerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	switch platform {
	case "instagram":
		influencer.Instagram.Approved = true
		influencer.Instagram.IsVerified = true
	case "youtube":
		influencer.YouTube.Approved = true
		influencer.YouTube.IsVerified = true
	case "snapchat":
		influencer.Snapchat.Approved = true
		influencer.Snapchat.IsVerified = true
	}

	// update the instagram profile in db
	err = influencer.Update()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "state": "influencer update"})
		return
	}

	err = whatsapp.SendInfluencerProfileApprovalTemplate(influencer.Phone)
	if err != nil {
		fmt.Println("err:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "state": "whatsapp message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": "Success"})
}

func DisapproveInfluencer(c *gin.Context) {

	influencerID := c.Query("influencerID")
	platform := c.Query("platform")

	if len(influencerID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty influencer ID"})
		return
	}

	influencer, err := models.GetInfluencerByID(influencerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	switch platform {
	case "instagram":
		influencer.Instagram.Approved = false
		influencer.Instagram.IsVerified = true
	case "youtube":
		influencer.YouTube.Approved = false
		influencer.YouTube.IsVerified = true
	case "snapchat":
		influencer.Snapchat.Approved = false
		influencer.Snapchat.IsVerified = true
	}

	// update the instagram profile in db
	err = influencer.Update()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "state": "influencer update"})
		return
	}

	err = whatsapp.SendInfluencerProfiledDisapprovalTemplate(influencer.Phone)
	if err != nil {
		fmt.Println("err:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "state": "whatsapp message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": "Success"})
}
