package controllers

import (
	"hermes/utils/amplitude"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func JoinReferralProgram(c *gin.Context) {
	go amplitude.TrackEventByAuth("Joined Referral Program", c)

	user, err := getUserObjectFromSession(c)
	if err != nil {
		log.Println("Failed to get user from session:", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	err = user.JoinReferralProgram(user.Name, "", make([]string, 0), 0)
	if err != nil {
		log.Println("Failed to join referral program:", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": "Successfully joined the program"})
}
