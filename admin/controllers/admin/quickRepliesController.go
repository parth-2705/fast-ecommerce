package controllers

import (
	"fmt"
	"hermes/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateQuickReply(c *gin.Context) {
	quickReply := models.QuickReply{}
	err := c.BindJSON(&quickReply)
	if err != nil {
		fmt.Println("create quick reply", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
		return
	}
	err = quickReply.CreateQuickReply()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to add to save quick reply"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "Successfully created quick reply"})
}

func UpdateQuickReply(c *gin.Context) {
	id := c.Query("id")
	quickReply, err := models.GetQuickReplyByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unable to find reply"})
		return
	}
	err = c.BindJSON(&quickReply)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
		return
	}
	err = quickReply.EditQuickReply()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to save quick reply"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "Successfully updated quick reply"})
}

func DeleteQuickReply(c *gin.Context) {
	id := c.Query("id")
	quickReply, err := models.GetQuickReplyByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unable to find reply"})
		return
	}
	err = quickReply.DeleteQuickReply()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to delete quick reply"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "Successfully deleted quick reply"})
}

func GetAllQuickReplies(c *gin.Context) {
	quickReplies, err := models.GetAllQuickReplies()
	if err!=nil{
		c.JSON(http.StatusNotFound, gin.H{"error": "Unable to get replies"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"quickReplies": quickReplies})
}
