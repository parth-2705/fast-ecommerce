package controllers

import (
	"fmt"
	"hermes/configs"
	"hermes/models"
	"hermes/services/Sentry"
	"hermes/utils/amplitude"
	"hermes/utils/data"
	"hermes/utils/messaging"
	"hermes/utils/tmpl"

	// "hermes/utils/Twilio"

	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type MobileNumber struct {
	CountryCode  string `form:"countryCode" json:"countryCode"`
	MobileNumber string `form:"mobileNumber" json:"mobileNumber"`
	OTPLogId     string
	Referrer     string `json:"referrer"`
	DirtyInput   string `json:"dirtyInput"`
}

func (mobileNumber MobileNumber) GetCompleteMobileNumber() (completeMobileNumber string) {
	completeMobileNumber = mobileNumber.CountryCode + mobileNumber.MobileNumber
	return
}

func SendOTP(c *gin.Context) {

	fmt.Println("Send OTP Controller")

	go amplitude.TrackEventByAuth("send Page", c)
	session := sessions.Default(c)
	user_session := session.Get(configs.Userkey)

	if user_session != nil {
		// user is already logged in
		// redirect him to categories page
		c.Redirect(http.StatusFound, "/categories")
		return
	}

	var mobileNumber MobileNumber
	c.BindJSON(&mobileNumber)

	// Check if last OTP was sent less than 30 secs ago

	if !models.CheckIfOTPCanbeSent(mobileNumber.GetCompleteMobileNumber()) {
		fmt.Println("OTP Sending Throttled")
		c.JSON(429, nil)
		return
	}

	otpLog, _ := models.CreateOTPLog(mobileNumber.CountryCode+mobileNumber.MobileNumber, data.GetUserAgentIDFromSession(c), mobileNumber.Referrer, mobileNumber.DirtyInput)
	mobileNumber.OTPLogId = otpLog.ID

	err := setMobileNumberInSession(c, mobileNumber)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	err = messaging.SendPhoneOTP(mobileNumber.GetCompleteMobileNumber())
	if err != nil {
		Sentry.SendErrorToSentry(c, err, map[string]string{"phone": mobileNumber.GetCompleteMobileNumber()})
		c.AbortWithError(400, err)
		return
	}

	c.JSON(200, gin.H{
		"message": "otp sent",
	})
}

// func VerifyOTP(c *gin.Context) {
// 	go amplitude.TrackEventByAuth("VerifyOTP Page", c)

// 	session := sessions.Default(c)
// 	user_session := session.Get(configs.Userkey)

// 	if user_session != nil {
// 		// user is already logged in
// 		// redirect him to categories page
// 		c.Redirect(http.StatusFound, "/categories")
// 		return
// 	}

// 	var otp OTP
// 	if err := c.Bind(&otp); err != nil {
// 		c.AbortWithError(http.StatusBadRequest, err)
// 		return
// 	}

// 	// check if otp is valid
// 	isValid, err := messaging.CheckPhoneOTP(otp.MobileNumber, otp.OTP)
// 	if err != nil {
// 		c.AbortWithError(http.StatusBadRequest, err)
// 		return
// 	}

// 	if !isValid {
// 		c.JSON(200, gin.H{
// 			"valid": isValid,
// 		})
// 		return
// 	}

// }

func SubmitOTP(c *gin.Context) {

	go amplitude.TrackEventByAuth("SubmitOTP Page", c)
	session := sessions.Default(c)
	user_session := session.Get(configs.Userkey)

	next := c.Request.URL.Query().Get("next")
	back := c.Request.URL.Query().Get("back")

	if back == "" {
		back = "/"
	}

	if user_session != nil {
		// user is already logged in
		// redirect him to the next page
		c.Redirect(http.StatusFound, next)
		return
	}

	mobileNumber, err := GetMobileNumberFromSession(c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.Header("Cache-Control", "no-store, must-revalidate")

	c.HTML(200, "submitOTP", gin.H{
		"next":         next,
		"back":         back,
		"mobileNumber": mobileNumber,
	})
}

func GetOtp(c *gin.Context) {

	go amplitude.TrackEventByAuth("GetOTP Page", c)

	session := sessions.Default(c)
	user_session := session.Get(configs.Userkey)

	if user_session != nil {
		// user is already logged in
		// redirect him to categories page
		c.Redirect(http.StatusFound, "/categories")
		return
	}
	// get mobile number from query params
	mobileNumber := c.Query("mobile")

	// Check if last OTP was sent less than 30 secs ago

	err := messaging.SendPhoneOTP(mobileNumber)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	userExists := models.UserExists(mobileNumber)

	c.JSON(200, gin.H{
		"message": "otp sent",
		"exists":  userExists,
	})

}

type OTP struct {
	OTP      string `json:"otp" binding:"required"`
	Whatsapp bool   `json:"whatsapp"`
}

func CheckOTP(c *gin.Context) {

	go amplitude.TrackEventByAuth("OTP Verifcation", c)

	session := sessions.Default(c)

	MobileNumber, err := GetMobileNumberFromSession(c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	// get mobile number and otp from request body which is json
	var otp OTP
	if err := c.BindJSON(&otp); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	trackingMap := amplitude.GetTrackingMap(c)

	// check if otp is valid
	isValid, sanitizedPhoneNum, err := messaging.CheckPhoneOTP(MobileNumber.CountryCode+MobileNumber.MobileNumber, otp.OTP)
	if err != nil {
		Sentry.SendErrorToSentry(c, err, map[string]string{"phone": MobileNumber.GetCompleteMobileNumber()})
		tmpl.TrackAmplitudeEvent("Unable to Verify OTP", trackingMap)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if !isValid {
		tmpl.TrackAmplitudeEvent("wrong OTP input", trackingMap)
		c.JSON(200, gin.H{
			"valid": isValid,
		})
		return
	}

	// mark OTP log as successful in DB
	models.UpdateSuccessStatusOfOTPLog(MobileNumber.OTPLogId)

	var user models.User
	user.Phone = sanitizedPhoneNum

	if y := models.UserExists(user.Phone); y {
		user, err = models.GetUser(user.Phone)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
	} else {
		referral := session.Get(configs.Referral)
		fmt.Println("referral code: ", referral)

		if referral != nil {
			referralCode, ok := referral.(string)

			if ok && len(referralCode) > 0 {
				// create the user with referral code
				user, err = user.CreateWithReferralCode(referralCode)
				if err != nil {
					c.AbortWithError(http.StatusBadRequest, err)
					return
				}
			} else {
				// create the user
				user, err = user.Create()
				if err != nil {
					c.AbortWithError(http.StatusBadRequest, err)
					return
				}
			}
		} else {
			// create the user
			user, err = user.Create()
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, err)
				return
			}
		}

		// create user's profile
		var profile models.Profile
		profile.UserID = user.ID
		profile.WhatsappEnabled = otp.Whatsapp
		err = profile.Create()
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

	}

	var influencer models.Influencer
	influencer.Phone = sanitizedPhoneNum

	err = influencer.Create(user.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session.Set(configs.Influencerkey, influencer.ID)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// map User Agent with User
	userAgentID, _ := getUserAgentIDFromSession(c)
	if _, err := models.CreateUserToUserAgentMapping(user.ID, userAgentID); err != nil {
		c.AbortWithError(http.StatusFound, err)
		return
	}

	session.Set(configs.Userkey, user.ID)
	if err := session.Save(); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	data.SetInternalUserInSession(c, user.Internal)

	postLoginEvents(c, user.ID)

	// check if next is set in url params
	next := c.Query("next")
	fmt.Println("next: ", next)
	if next != "" && isValid {
		c.Redirect(http.StatusFound, next)
		return
	}

	tmpl.TrackAmplitudeEvent("login success", trackingMap)

	c.JSON(200, gin.H{
		"valid": isValid,
	})
}
