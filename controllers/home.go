package controllers

import (
	"hermes/models"
	"hermes/utils/amplitude"
	"hermes/utils/data"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func AboutUsPage(c *gin.Context) {
	go amplitude.TrackEventByAuth("About Us Page", c)
	c.HTML(http.StatusOK, "about-us.html", gin.H{})
}

func ContactUsGetHandler(c *gin.Context) {
	go amplitude.TrackEventByAuth("Contact Us Page", c)
	c.HTML(http.StatusOK, "contact-us.html", gin.H{})
}

func ContactUsPostHandler(c *gin.Context) {

	go amplitude.TrackEventByAuth("Contact Us Form Submitted", c)

	// will handle form post data
	contact := models.Contact{}
	c.Bind(&contact)

	// will send email to admin
	// TODO : send email to admin

	// save contact to database
	err := contact.SaveToDB()
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// set session flash message
	data.AddSessionFlashMessage(c, "Thank you for contacting us. We will get back to you soon.")

	c.HTML(http.StatusOK, "contact-us.html", gin.H{
		"flashes": data.GetSessionFlashMessages(c),
	})
}

func PrivacyPolicyPage(c *gin.Context) {
	go amplitude.TrackEventByAuth("Privacy Policy Page", c)
	c.HTML(http.StatusOK, "privacy-policy.html", gin.H{})
}

func TermsAndConditionsPage(c *gin.Context) {
	go amplitude.TrackEventByAuth("Terms And Condition Page", c)
	c.HTML(http.StatusOK, "tnc.html", gin.H{})
}

func ExchangeAndReturnPolicyPage(c *gin.Context) {
	go amplitude.TrackEventByAuth("Exchange and Return Policy Page", c)
	c.HTML(http.StatusOK, "exchange-and-return-policy.html", gin.H{})
}

func OnlineRegistrationPolicyPage(c *gin.Context) {
	go amplitude.TrackEventByAuth("Online Registration Policy Page", c)
	c.HTML(http.StatusOK, "online-registration-policy.html", gin.H{})
}

func PriceAndPaymentPolicyPage(c *gin.Context) {
	go amplitude.TrackEventByAuth("Price And Payment Policy Page", c)
	c.HTML(http.StatusOK, "price-and-payment-policy.html", gin.H{})
}
