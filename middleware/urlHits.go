package middleware

import (
	"hermes/controllers"
	"hermes/models"
	"hermes/utils/data"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// Remake Query Params
func formatQueryParams(queryParams url.Values) models.URLParams {
	formatedQueryParams := make(models.URLParams)
	for key := range queryParams {
		formatedQueryParams[key] = queryParams.Get(key)
	}

	return formatedQueryParams
}

// Remkae Headers
func formatHeaders(queryParams http.Header) models.URLParams {
	formatedQueryParams := make(models.URLParams)
	for key := range queryParams {
		formatedQueryParams[key] = queryParams.Get(key)
	}

	return formatedQueryParams
}

// Check if this path is to be logged
func recordURLHit(path string) bool {
	for i := range pathsToNotLog {
		if strings.HasPrefix(path, pathsToNotLog[i]) {
			return false
		}
	}

	return true
}

var pathsToNotLog []string = []string{"/health", "/static", "/favicon.io"}

func UrlHitsLogging(c *gin.Context) {

	// get Information Required to log URL Hit from context
	url := c.Request.URL.Path
	var log models.URLHit
	if recordURLHit(url) {

		params := formatQueryParams(c.Request.URL.Query())
		method := c.Request.Method
		eventName := ""
		userAgentID := data.GetUserAgentIDFromSession(c)
		userID := data.GetUserIDFromSession(c)
		number, _ := controllers.GetMobileNumberFromSession(c)
		internal := data.GetIfUserIsInternalFromSession(c)
		headers := formatHeaders(c.Request.Header)

		// Log in DB
		log, _ = models.CreateURLHit(url, method, eventName, userAgentID, userID, number.GetCompleteMobileNumber(), params, internal, headers)
	}

	// move on to next handler
	c.Next()

	responseStatus := c.Writer.Status()
	if log.ID != "" { // Only Update if log was successfully created
		log.LogResponseStatus(responseStatus)
	}

}
