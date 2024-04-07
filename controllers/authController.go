package controllers

import (
	"encoding/json"
	"fmt"
	"hermes/models"
	"hermes/utils/data"
	"hermes/utils/tmpl"
	"hermes/utils/whatsapp"
	"log"
	"net/http"
	"os"

	"hermes/configs"

	"hermes/utils/amplitude"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/tryamigo/themis"
)

func SignInUpPage(c *gin.Context) {
	// Amplitude tracking
	go amplitude.TrackEventByAuth("SignInUp Page", c)

	session := sessions.Default(c)
	user_session := session.Get(configs.Userkey)

	next := c.Request.URL.Query().Get("next")
	back := c.Request.URL.Query().Get("back")
	referral := c.Request.URL.Query().Get("code")

	if user_session != nil {
		// user is already logged in
		// redirect him to home page

		user, err := models.GetUserByID(user_session.(string))
		if err == nil {

			var influencer models.Influencer
			influencer.Phone = user.Phone
			err = influencer.Create(user.ID)
			if err != nil {
				session.Clear()
				session.Save()
			} else {
				session.Set(configs.Influencerkey, influencer.ID)
				if err := session.Save(); err != nil {
					c.Redirect(http.StatusFound, next)
					return
				}
			}

			log.Println("User already logged in././.")
			c.Redirect(http.StatusFound, next)
			return
		}
		session.Clear()
		session.Save()
	}

	// get query parameters from url
	waId := c.Query("waId")

	wishlist := c.Request.URL.Query().Get("wishlist")
	data.SetSessionValue(c, "setProductToSession", wishlist)
	notify := c.Request.URL.Query().Get("notify")
	data.SetSessionValue(c, "setNotifyToSession", notify)

	if waId != "" {
		// user is coming from whatsapp auth link
		authResponse, err := getDataFromWhatsappAuthlink(waId)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		var user models.User
		user.Phone = authResponse.User.WaNumber

		if y := models.UserExists(user.Phone); y {
			user, err = models.GetUser(user.Phone)
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, err)
				return
			}
		} else {
			user, err = CreateUserFromAuthLinkResponse(authResponse)
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, err)
				return
			}
		}

		session.Set(configs.Userkey, user.ID)
		if err := session.Save(); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		var influencer models.Influencer
		influencer.Phone = authResponse.User.WaNumber

		err = influencer.Create(user.ID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// check if next is set in url params
		if next != "" {
			c.Redirect(http.StatusTemporaryRedirect, next)
			return
		}

		c.Redirect(http.StatusTemporaryRedirect, "/categories")
	}

	c.Header("Cache-Control", "no-store, must-revalidate")
	data.SetSessionValue(c, "next", next)

	referralCodeLink := session.Get(configs.Referral)
	var referredByUserCode string

	if referralCodeLink == nil {
		referredByUserCode = ""
	} else {
		code, ok := referralCodeLink.(string)
		if ok && len(code) > 0 {
			referredByUserCode = code
		} else {
			referredByUserCode = ""
		}
	}

	if len(referral) > 0 {
		fmt.Println("referral code set in session: ", referral)
		data.SetSessionValue(c, configs.Referral, referral)
	} else if len(referredByUserCode) > 0 {
		fmt.Println("referredByUser code set in session: ", referredByUserCode)
		data.SetSessionValue(c, configs.Referral, referredByUserCode)
	}

	number := os.Getenv("WHATSAPP_PHONE_NUMBER")

	c.HTML(http.StatusOK, "sign-in-up-2", gin.H{
		"back":            back,
		"next":            next,
		"whatsappNum":     number,
		"referralCode":    referral,
		"hasReferralCode": len(referral) > 0,
	})
}

func SignOut(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(configs.Userkey)
	log.Println("logging out user:", user)
	if user == nil {
		log.Println("Invalid session token")
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}
	session.Set(configs.Userkey, nil)
	session.Set(configs.Influencerkey, nil)
	userAgentID, _ := getUserAgentIDFromSession(c)
	session.Clear()
	session.Options(sessions.Options{Path: "/", MaxAge: -1})
	data.SetSessionValue(c, configs.UserAgentIdentifier, userAgentID)
	session.Save()
	if err := session.Save(); err != nil {
		log.Println("Failed to save session:", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	models.UpdateUserUserAgenMappingWithLogoutTime(user.(string), userAgentID)

	c.Redirect(http.StatusTemporaryRedirect, "/")
}

func getUserObjectFromSession(c *gin.Context) (models.User, error) {
	session := sessions.Default(c)
	userID := session.Get(configs.Userkey)
	if userID == nil {
		return models.User{}, fmt.Errorf("invalid session token")
	}
	// convert interface to string
	userIDString, ok := userID.(string)
	if !ok {
		return models.User{}, fmt.Errorf("invalid session token")
	}
	user, err := models.GetUserByID(userIDString)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func getInfluencerObjectFromSession(c *gin.Context) (models.Influencer, error) {
	session := sessions.Default(c)
	influencerID := session.Get(configs.Influencerkey)
	if influencerID == nil {
		return models.Influencer{}, fmt.Errorf("invalid session token")
	}
	// convert interface to string
	influencerIDString, ok := influencerID.(string)
	if !ok {
		return models.Influencer{}, fmt.Errorf("invalid session token")
	}

	influencer, err := models.GetInfluencerByID(influencerIDString)
	if err != nil {
		return models.Influencer{}, err
	}
	return influencer, nil
}

func setAddressIDInSession(c *gin.Context, addressID string) (err error) {
	session := sessions.Default(c)
	session.Set(configs.AddressKey, addressID)

	err = session.Save()
	if err != nil {
		fmt.Println(err)
		return
	}

	return
}

func getAddressObjectFromSession(c *gin.Context) (address models.Address, err error) {
	session := sessions.Default(c)
	addressID := session.Get(configs.AddressKey)
	if addressID == nil {
		return address, fmt.Errorf("invalid session token")
	}
	// convert interface to string
	addressIDDString, ok := addressID.(string)
	if !ok {
		return address, fmt.Errorf("invalid session token")
	}
	address, err = models.GetAddress(addressIDDString)
	if err != nil {
		return
	}
	return
}

func setMobileNumberInSession(c *gin.Context, mobileNumber MobileNumber) (err error) {
	session := sessions.Default(c)
	// session.Set(configs.MobileNumberKey, mobileNumber.MobileNumber)
	// session.Set(configs.CountryCodeKey, mobileNumber.CountryCode)

	session.Set(configs.MobileNumberKey, mobileNumber)

	err = session.Save()
	if err != nil {
		fmt.Println(err)
		return
	}

	return
}

func GetMobileNumberFromSession(c *gin.Context) (completeMobileNumber MobileNumber, err error) {
	session := sessions.Default(c)
	// mobileNumber := session.Get(configs.MobileNumberKey)
	// if mobileNumber == nil {
	// 	return completeMobileNumber, fmt.Errorf("invalid session token")
	// }

	// countryCode := session.Get(configs.CountryCodeKey)
	// if countryCode == nil {
	// 	return completeMobileNumber, fmt.Errorf("invalid session token")
	// }

	// completeMobileNumber.MobileNumber = mobileNumber.(string)
	// completeMobileNumber.CountryCode = countryCode.(string)

	mobileNumber2 := session.Get(configs.MobileNumberKey)
	completeMobileNumber, ok := mobileNumber2.(MobileNumber)

	if !ok {
		return completeMobileNumber, fmt.Errorf("no mobile number in session")
	}

	return
}

func getUserAgentIDFromSession(c *gin.Context) (string, error) {
	userAgentIdenifierInterface := sessions.Default(c).Get(configs.UserAgentIdentifier)

	userAgentID, ok := userAgentIdenifierInterface.(string)
	if !ok {
		return userAgentID, fmt.Errorf("session ID not a string")
	}

	return userAgentID, nil
}

func ReferralSignUp(c *gin.Context) {
	code := c.Param("code")
	_, err := getUserObjectFromSession(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/auth/sign-in-up?code="+code)
		return
	}

	c.Redirect(http.StatusFound, "/")
}

func MagicLinkLogin(c *gin.Context) {
	session := sessions.Default(c)
	code := c.Param("code")
	endpoint := whatsapp.WHATSAPP_SERVICE_BASE_URL + "/magic"
	jsonBody := map[string]string{
		"code": code,
	}
	body, _ := json.Marshal(jsonBody)
	headers := [][]string{
		{"Authorization", fmt.Sprintf("Bearer %s", whatsapp.WHATSAPP_SERVICE_API_KEY)},
	}
	status, resp, err := themis.HitAPIEndpoint2(endpoint, "POST", body, headers, nil)
	fmt.Println("RESP MAGIC LOGIN", string(resp), endpoint)
	if err != nil {
		fmt.Println("error in sending request", err)
		return
	}
	response := map[string]string{}
	json.Unmarshal(resp, &response)
	if status == 403 || status == 400 {
		if session.Get(configs.Userkey) != nil {
			c.Redirect(http.StatusFound, os.Getenv("BASE_URL")+response["path"])
		}
		c.JSON(status, response)
		return
	} else if status > 400 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	mobileNumber := MobileNumber{
		DirtyInput:   response["phone"],
		MobileNumber: response["phone"][3:13],
		CountryCode:  response["phone"][0:3],
	}

	var user models.User
	user.Phone = mobileNumber.GetCompleteMobileNumber()

	if y := models.UserExists(user.Phone); y {
		user, err = models.GetUser(user.Phone)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
	} else {

		referralCode := response["Referred Code"]
		if len(referralCode) > 0 {
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

		// create user's profile
		var profile models.Profile
		profile.UserID = user.ID
		profile.WhatsappEnabled = true
		err = profile.Create()
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
	}

	var influencer models.Influencer
	influencer.Phone = mobileNumber.GetCompleteMobileNumber()

	err = influencer.Create(user.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	err = setMobileNumberInSession(c, mobileNumber)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	postLoginEvents(c, user.ID)

	redirect := data.GetSessionValue(c, "next")
	fmt.Println("redirect:", redirect)
	next := ""
	if redirect != nil {
		next = redirect.(string)
	}
	if response["path"] != "" {
		next = response["path"]
	}

	trackingMap := amplitude.GetTrackingMap(c)
	tmpl.TrackAmplitudeEvent("WA login complete", trackingMap)

	c.Redirect(http.StatusFound, os.Getenv("BASE_URL")+next)
}

func postLoginEvents(c *gin.Context, userID string) error {
	// if Cart is present in Session. Associate it with logged in User
	cartID := data.GetCartFromSession(c)
	if cartID == "" {
		return nil
	}

	models.AssociateCartWithUser(cartID, userID)

	return nil
}
