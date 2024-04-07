package controllers

import (
	"hermes/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ApproveInfluencerCampaignApplication(ctx *gin.Context) {
	id := ctx.Param("id")
	application, err := models.GetInfluencerApplicationByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}
	err = application.Approve()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Falied to approve"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"approved": true})
}
