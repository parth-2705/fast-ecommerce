package controllers

import (
	"fmt"
	"hermes/models"
	"hermes/utils/data"
	"net/http"

	"github.com/gin-gonic/gin"
)

func NotifyToggle(c *gin.Context) {
	var notifyToggleBody WishlistPostBody
	err := c.BindJSON(&notifyToggleBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
		return
	}
	user, err := getUserObjectFromSession(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get User"})
		data.SetSessionValue(c, "setNotifyToSession", notifyToggleBody.ProductID)
		return
	}
	notify, err := addOrRemoveNotifyObjectFromDB(user.ID, notifyToggleBody.ProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"notify": notify})
}

func addOrRemoveNotifyObjectFromDB(userID string, productID string) (bool, error) {
	isInNotifylist := models.GetIfItemIsInNotifylistOfUser(userID, productID)
	fmt.Println()
	fmt.Println(isInNotifylist)
	fmt.Println()
	notify := models.Notify{
		UserID:    userID,
		ProductID: productID,
	}
	if isInNotifylist {
		err := models.RemoveProductFromNotify(notify)
		if err != nil {
			return true, err
		}
		return false, nil
	} else {
		err := models.AddProductToNotify(notify)
		if err != nil {
			return false, err
		}
		return true, nil
	}
}
