package controllers

import (
	"encoding/json"
	"hermes/models"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/tryamigo/themis"
	"gorm.io/datatypes"
)

func GetInstagramOAuthURL(c *gin.Context) {

	redirect := c.Query("redirect")
	if redirect == "" {
		redirect = "/influencer/authorize"
	}

	influencer, err := getInfluencerObjectFromSession(c)
	if err != nil || len(influencer.ID) == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"err": err.Error()})
		return
	}

	oauth := models.OAuth{
		InfluencerID: influencer.ID,
		Phone:        influencer.Phone,
		Redirect:     redirect,
	}

	err = oauth.Create()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}

	url := models.GetInstagramOAuthURL(oauth.ID)

	c.JSON(http.StatusOK, gin.H{"url": url})
}

func HandleInstagramOAuthCallback(c *gin.Context) {

	state := c.Query("state")
	code := c.Query("code")

	if len(state) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No Oauth state found"})
		return
	}

	if len(code) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No Authorization code found"})
		return
	}

	oauth, err := models.GetOauthByID(state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "state": "oauth"})
		return
	}

	token, err := models.HandleInstagramCallback(code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "state": "access_token"})
		return
	}

	vaultHandler := themis.NewVaultHandler(os.Getenv("VAULT_URL"), os.Getenv("VAULT_TOKEN"), oauth.InfluencerID+"/instagram")

	tokenBytes, err := json.Marshal(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "state": "token to bytes"})
		return
	}

	var tokenData datatypes.JSON
	err = json.Unmarshal(tokenBytes, &tokenData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "state": "token data"})
		return
	}

	statusCode, err := vaultHandler.AddSecretsToVault(tokenData)
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error(), "state": "add token secrets"})
		return
	}

	instagramProfile, err := models.GetInstagramUserProfile(token.AccessToken, oauth.InfluencerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "state": "instagramProfile"})
		return
	}

	influencer, err := models.GetInfluencerByID(oauth.InfluencerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "state": "influencer"})
		return
	}

	instagramProfile.IsConnected = true
	influencer.Instagram = instagramProfile
	err = influencer.Update()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "state": "influencer update"})
		return
	}

	c.Redirect(http.StatusFound, oauth.Redirect)
}

func UpdateInstagramProfile(c *gin.Context) {

	influencer, err := getInfluencerObjectFromSession(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "state": "get influencer from session"})
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

func GetYouTubeOAuthURL(c *gin.Context) {

	redirect := c.Query("redirect")
	if redirect == "" {
		redirect = "/influencer/authorize"
	}

	influencer, err := getInfluencerObjectFromSession(c)
	if err != nil || len(influencer.ID) == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"err": err.Error()})
		return
	}

	oauth := models.OAuth{
		InfluencerID: influencer.ID,
		Phone:        influencer.Phone,
		Redirect:     redirect,
	}

	err = oauth.Create()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}

	url := models.GetYouTubeOAuthURL(oauth.ID)

	c.JSON(http.StatusOK, gin.H{"url": url})
}

func HandleYouTubeOAuthCallback(c *gin.Context) {

	state := c.Query("state")
	code := c.Query("code")

	if len(state) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No Oauth state found"})
		return
	}

	if len(code) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No Authorization code found"})
		return
	}

	oauth, err := models.GetOauthByID(state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "state": "oauth"})
		return
	}

	token, err := models.HandleYouTubeCallback(code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "state": "access_token"})
		return
	}

	vaultHandler := themis.NewVaultHandler(os.Getenv("VAULT_URL"), os.Getenv("VAULT_TOKEN"), oauth.InfluencerID+"/youtube")

	tokenBytes, err := json.Marshal(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "state": "token to bytes"})
		return
	}

	var tokenData datatypes.JSON
	err = json.Unmarshal(tokenBytes, &tokenData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "state": "token data"})
		return
	}

	statusCode, err := vaultHandler.AddSecretsToVault(tokenData)
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error(), "state": "add token secrets"})
		return
	}

	youtubeProfile, err := models.GetYouTubeUserProfile(token.AccessToken, oauth.InfluencerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "state": "youtubeProfile"})
		return
	}

	influencer, err := models.GetInfluencerByID(oauth.InfluencerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "state": "influencer"})
		return
	}

	youtubeProfile.IsConnected = true
	influencer.YouTube = youtubeProfile
	err = influencer.Update()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "state": "influencer update"})
		return
	}

	c.Redirect(http.StatusFound, "/influencer/authorize")
}

func UpdateYouTubeProfile(c *gin.Context) {

	influencer, err := getInfluencerObjectFromSession(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "state": "get influencer from session"})
		return
	}

	// Get access token from vault
	vaultHandler := themis.NewVaultHandler(os.Getenv("VAULT_URL"), os.Getenv("VAULT_TOKEN"), influencer.ID+"/youtube")
	statusCode, respBody, err := vaultHandler.GetSecretsFromVault()
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error(), "state": "get secrets"})
		return
	}

	var youtubeSecret map[string]string
	err = json.Unmarshal(respBody, &youtubeSecret)
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error(), "state": "unmarshal secrets"})
		return
	}

	access_token := youtubeSecret["access_token"]
	youtubeProfile, err := models.GetYouTubeUserProfile(access_token, influencer.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "state": "instagram profile"})
		return
	}

	// don't change the createdAt time while updating the profile
	creationTime := influencer.YouTube.CreatedAt
	influencer.YouTube = youtubeProfile
	influencer.YouTube.CreatedAt = creationTime

	// update the YouTube profile in db
	err = influencer.Update()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "state": "influencer update"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": "Success"})
}
