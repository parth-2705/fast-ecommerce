package middleware

import (
	"hermes/configs"
	"hermes/models/Logs"
	"hermes/utils/data"
	"strings"

	"github.com/gin-gonic/gin"
)

func SessionManagementMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		IPAddress := strings.Split(c.Request.RemoteAddr, ":")[0]

		// check if a session ID exists already
		userAgentIdentifier := data.GetSessionValue(c, configs.UserAgentIdentifier)
		if userAgentIdentifier == nil {
			// if does not exist create one

			session, err := Logs.CreateNewSession(c.Request.Header.Get("User-Agent"), IPAddress)
			if err != nil {
				c.JSON(500, gin.H{"error": "error writing session"})
				c.Abort()
			}

			data.SetSessionValue(c, configs.UserAgentIdentifier, session.ID)
			data.SetUserAgentInSession(c, session.UserAgent)
		}

		userAgent := data.GetUserAgentFromSession(c)
		if userAgent == "" {
			data.SetUserAgentInSession(c, c.Request.Header.Get("User-Agent"))
		}

		data.SetIPAddressinSession(c, IPAddress)
	}
}
