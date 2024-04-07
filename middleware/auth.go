package middleware

import (
	"fmt"
	"hermes/configs"
	"hermes/controllers"
	"hermes/services/Sentry"
	"hermes/utils/network"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// define a gin middleware function
func AuthRequired(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(configs.Userkey)
	backURL, _ := url.Parse(c.Request.Referer())
	back := backURL.Path
	if user == nil {
		if network.MobileRequest(c) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		log.Println("User not logged in")

		// Don't see the need of clearing the session, when User tries to access a page behind login. Just need to redirect User to Login Flow and add UserID to session
		// the redirect url will be /auth/sign-in-up + ?next=the-requested-url
		// session.Set(configs.Userkey, nil)
		// session.Clear()
		// session.Options(sessions.Options{Path: "/", MaxAge: -1})
		// session.Save()
		url := c.Request.URL.String()
		redirect_url := "/auth/sign-in-up?next=" + url + "&back=" + back
		// check if the request url is /auth/sign-out
		if url == "/auth/sign-out" {
			redirect_url = "/auth/sign-in-up"
		}

		c.Redirect(302, redirect_url)
		c.Abort()
		return
	}
	c.Next()
}

// // Middleware for API Authentication // Left to do
// func APIAuthentication(c *gin.Context) {
// 	// Get Secret from Header

// 	// Validate Secret

// 	// Allow Disallow Request
// 	c.Next()
// }


func getAuthorizationToken(authHeader string) (token string, err error) {
	authSplit := strings.Split(authHeader, " ")
	if len(authSplit) != 2 {
		return "", fmt.Errorf("auth Header not correctly formatted")
	}

	if authSplit[0] != "Bearer" {
		return "", fmt.Errorf("bearer not present")
	}

	return authSplit[1], nil
}

func DecentroWebhookAuth(c *gin.Context) {
	// Get Token

	token, err := getAuthorizationToken(c.GetHeader("Authorization"))

	if err != nil {

		Sentry.SendErrorToSentry(c, fmt.Errorf("Could not get token from Auth Header: "+c.GetHeader("Authorization")), nil)

		c.JSON(401, controllers.DecentroWebhook401Response)
		c.Abort()
		return
	}

	// Check if Token is Valid
	if token != os.Getenv("DECENTRO_WEBHOOK_TOKEN") {

		Sentry.SendErrorToSentry(c, fmt.Errorf("Auth Token did not match: "+token), nil)

		// return error if Invalid token
		c.JSON(401, controllers.DecentroWebhook401Response)
		c.Abort()
		return
	}

	// Move to Next Handler if Valid token
	c.Next()
}
