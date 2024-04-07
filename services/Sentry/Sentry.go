package Sentry

import (
	"hermes/utils/data"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

func addSessionInfoToTags(c *gin.Context, tags map[string]string) map[string]string {
	if tags == nil {
		tags = make(map[string]string)
	}

	tags["userAgentID"] = data.GetUserAgentIDFromSession(c)
	tags["userID"] = data.GetUserIDFromSession(c)

	return tags
}

// Non Blocking Function, main login runs in a go routine
func SendErrorToSentry(c *gin.Context, err error, tags map[string]string) {

	tags = addSessionInfoToTags(c, tags)

	if hub := sentrygin.GetHubFromContext(c); hub != nil {
		hub.Scope().SetTags(tags)
		hub.Scope().SetUser(sentry.User{ID: data.GetUserIDFromSession(c)})
		hub.CaptureException(err)
	}
}

func SentryCaptureException(err error) {
	sentry.CaptureException(err)
}

func SentryCaptureMessage(msg string) {
	sentry.CaptureMessage(msg)
}
