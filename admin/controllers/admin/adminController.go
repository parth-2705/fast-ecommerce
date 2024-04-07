package controllers

import (
	"fmt"
	"hermes/configs"
	"hermes/models"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func GetAdminInfo(c *gin.Context) {
	session := sessions.Default(c)
	userFromSession := session.Get(configs.Userkey)
	userID := ""
	if userFromSession != nil {
		userID = userFromSession.(string)
	}
	admin, _ := models.GetAdminbyID(userID)
	fmt.Println("ADMIN: ", userID, admin)
	c.JSON(http.StatusOK, admin)
}
