package controllers

import (
	"hermes/models"
	"hermes/utils/amplitude"
	"net/http"

	"github.com/gin-gonic/gin"
)

func JoinReferralProgramPage(c *gin.Context) {
	go amplitude.TrackEventByAuth("Join Referral Porgram Page", c)

	user, err := getUserObjectFromSession(c)
	loggedIn := true
	if err != nil {
		loggedIn = false
	} else {
		if user.HasJoinedReferralProgram {
			c.Redirect(http.StatusFound, "/referral")
			return
		}
	}

	c.HTML(http.StatusOK, "join-referral", gin.H{"loggedIn": loggedIn})
}

func ReferralProgramPage(c *gin.Context) {
	user, err := getUserObjectFromSession(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/auth/sign-in-up?next=/referral/join&back=/referral/join")
		return
	}

	if !user.HasJoinedReferralProgram {
		c.Redirect(http.StatusFound, "/referral/join")
		return
	}

	profile, err := user.GetProfile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	wallet, err := models.GetUserWallet(profile.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	referredUsers, err := profile.MyReferredProfiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "referral-page", gin.H{"user": user, "profile": profile, "wallet": wallet, "referredUsers": referredUsers})
}
