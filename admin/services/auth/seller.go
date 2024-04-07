package auth

import (
	"fmt"
	"hermes/configs"
	"hermes/models"
	"hermes/utils/messaging"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// func SellerLogin(c *gin.Context) {
// 	phone := c.Query("phone")
// 	c.HTML(http.StatusOK, "login", gin.H{"title": "Login Seller", "phone": phone})
// }

func VerifyOTPPage(c *gin.Context) {
	loginForm := make(map[string]string)
	if err := c.ShouldBind(loginForm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input in wrong format"})
		return
	}

	if loginForm["phone"] == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone is empty"})
		return
	}

	mobileNumber := "+91"
	mobileNumber += loginForm["phone"]
	err := messaging.SendPhoneOTP(mobileNumber)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.HTML(http.StatusOK, "otp", gin.H{"title": "Login Seller", "phone": loginForm["phone"]})
}

type OTP struct {
	MobileNumber string `json:"phone" form:"phone"`
	OTP          string `json:"otp" form:"otp"`
}

func AuthorizeAndCreateSellerMember(c *gin.Context, otp OTP) {

	session := sessions.Default(c)
	var sellerMember models.SellerMember
	var seller models.Seller
	var err error

	sellerMember.Phone = otp.MobileNumber

	if ok := models.DoesSellerMemberExist(sellerMember.Phone); ok {
		sellerMember, err = models.GetSellerMemberByPhone(sellerMember.Phone)
		if err != nil {
			fmt.Println(fmt.Errorf("error: %s", err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error in getting seller member " + err.Error()})
			return
		}
		seller, err = models.GetSellerByID(sellerMember.SellerID)
		if err != nil {
			fmt.Println(fmt.Errorf("error: %s", err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error in getting seller " + err.Error()})
			return
		}
	} else {
		sellerMember, err = sellerMember.Create()
		if err != nil {
			fmt.Println(fmt.Errorf("error: %s", err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error in creating seller member " + err.Error()})
			return
		}
		seller = models.Seller{}
		// seller.Create()
	}
	session.Set(configs.Userkey, sellerMember.ID)
	session.Set(configs.SellerHeader, sellerMember.SellerID)

	err = session.Save()

	if err != nil {
		fmt.Println(fmt.Errorf("error: %s", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !seller.ProfileCompleted {
		c.JSON(http.StatusOK, gin.H{"redirect": "/info"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"redirect": "/orders"})
	return
}

func OTPVerification(c *gin.Context) {

	// get mobile number and otp from request body which is json
	var otp OTP
	if err := c.ShouldBind(&otp); err != nil {
		fmt.Println(fmt.Errorf("error: %s", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP in wrong format " + err.Error()})
		return
	}

	// check if otp is valid
	isValid, _, err := messaging.CheckPhoneOTP(otp.MobileNumber, otp.OTP) // Need to capture the phone OTP was sent too.
	// otp.MobileNumber = phoneNumberReturnedByTwillio
	if err != nil {
		fmt.Println(fmt.Errorf("error: %s", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error in checking phone otp " + err.Error()})
		return
	}

	if !isValid {
		// c.Redirect(http.StatusFound, "/login")
		c.JSON(http.StatusOK, gin.H{"valid": false})
		return
	}

	AuthorizeAndCreateSellerMember(c, otp)
}

func GetSellerFromSession(c *gin.Context) (models.Seller, error) {
	session := sessions.Default(c)
	sellerID := session.Get(configs.SellerHeader)

	if sellerID == nil {
		return models.Seller{}, fmt.Errorf("invalid session token")
	}
	// convert interface to string
	sellerIDString, ok := sellerID.(string)
	if !ok || len(sellerIDString) == 0 {
		return models.Seller{}, fmt.Errorf("invalid session token")
	}
	seller, err := models.GetSellerByID(sellerIDString)
	if err != nil {
		SignOut2(c)
		return models.Seller{}, err
	}
	return seller, nil
}

func GetSellerMemberFromSession(c *gin.Context) (models.SellerMember, error) {
	session := sessions.Default(c)
	memberID := session.Get(configs.Userkey)

	if memberID == nil {
		return models.SellerMember{}, fmt.Errorf("invalid session token")
	}
	// convert interface to string
	memberIDString, ok := memberID.(string)
	if !ok || len(memberIDString) == 0 {
		return models.SellerMember{}, fmt.Errorf("invalid session token")
	}
	sellerMember, err := models.GetSellerMemberByID(memberIDString)
	if err != nil {
		SignOut2(c)
		return models.SellerMember{}, err
	}
	return sellerMember, nil
}

func GetSellerMapFromSession(c *gin.Context) (map[string]interface{}, error) {
	session := sessions.Default(c)
	sellerID := session.Get(configs.SellerHeader)

	if sellerID == nil {
		return map[string]interface{}{}, fmt.Errorf("invalid session token")
	}
	// convert interface to string
	sellerIDString, ok := sellerID.(string)
	if !ok || len(sellerIDString) == 0 {
		return map[string]interface{}{}, fmt.Errorf("invalid session token")
	}
	seller, err := models.GetSellerMapByID(sellerIDString)
	if err != nil {
		SignOut2(c)
		return map[string]interface{}{}, err
	}
	return seller, nil
}

func GetSellerIdFromSession(c *gin.Context) (string, error) {
	session := sessions.Default(c)
	userID := session.Get(configs.SellerHeader)

	if userID == nil {
		return "", fmt.Errorf("invalid session token")
	}
	// convert interface to string
	sellerIDString, ok := userID.(string)
	if !ok || len(sellerIDString) == 0 {
		SignOut2(c)
		return "", fmt.Errorf("invalid session token")
	}

	return sellerIDString, nil
}

func GetMemberIdFromSession(c *gin.Context) (string, error) {
	session := sessions.Default(c)
	userID := session.Get(configs.Userkey)

	if userID == nil {
		return "", fmt.Errorf("invalid session token")
	}
	// convert interface to string
	sellerIDString, ok := userID.(string)
	if !ok || len(sellerIDString) == 0 {
		SignOut2(c)
		return "", fmt.Errorf("invalid session token")
	}

	return sellerIDString, nil
}

func SignOut(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(configs.Userkey)
	if user == nil {
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return
	}
	session.Set(configs.Userkey, nil)
	session.Clear()
	session.Options(sessions.Options{Path: "/", MaxAge: -1})
	session.Save()
	if err := session.Save(); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, "/login")
}

func SellerSignInUpV2(c *gin.Context) {
	loginForm := make(map[string]string)
	if err := c.ShouldBind(&loginForm); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if loginForm["phone"] == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	mobileNumber := "+91"
	mobileNumber += loginForm["phone"]
	err := messaging.SendPhoneOTP(mobileNumber)
	if err != nil {
		c.AbortWithError(http.StatusServiceUnavailable, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func SignOut2(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(configs.Userkey)
	if user == nil {
		c.JSON(http.StatusOK, gin.H{"redirect": "/login"})
		return
	}
	session.Set(configs.Userkey, nil)
	session.Clear()
	session.Options(sessions.Options{Path: "/", MaxAge: -1})
	session.Save()
	if err := session.Save(); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"redirect": "/login"})
	return
}
