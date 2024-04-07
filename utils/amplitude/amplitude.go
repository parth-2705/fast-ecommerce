package amplitude

import (
	"fmt"
	"hermes/configs"
	"hermes/models"
	"hermes/services/sendgridEmail"
	"os"

	"github.com/amplitude/analytics-go/amplitude"
	"github.com/getsentry/sentry-go"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

var ANALYTICS amplitude.Client

func Init() {
	analytics := amplitude.NewClient(
		amplitude.NewConfig("1ac59e70edfb1fdd5fd12c82ebde83ce"),
	)
	ANALYTICS = analytics
}

func TrackEventByAuth(eventName string, c *gin.Context) {

	defer sentry.Recover()

	if os.Getenv("ENVIRONMENT") != "prod" {
		return
	}

	session := sessions.Default(c)
	user_session := session.Get(configs.Userkey)

	if user_session == nil {
		TrackEvent(eventName, c)
	} else {
		TrackEventByUserID(eventName, user_session.(string))
	}
}

func TrackEvent(eventName string, c *gin.Context) {

	// Track a basic event
	// One of UserID and DeviceID is required
	deviceID := c.Request.Header.Get("User-Agent")

	event := amplitude.Event{
		EventType: eventName,
		DeviceID:  deviceID,
	}
	ANALYTICS.Track(event)

	// Flush the event buffer
	ANALYTICS.Flush()
}

func TrackEventByUserID(eventName string, userID string) {

	defer sentry.Recover()

	// Track a basic event
	// One of UserID and DeviceID is required

	internal := isUserInternal(userID)

	eventProps := map[string]interface{}{
		"internal": internal,
	}

	event := amplitude.Event{
		EventProperties: eventProps,
		EventType:       eventName,
		UserID:          userID,
	}

	ANALYTICS.Track(event)

	// Flush the event buffer
	ANALYTICS.Flush()
}

func TrackEventByDeviceID(eventName string, deviceID string) {

	defer sentry.Recover()
	// Track a basic event
	// One of UserID and DeviceID is required

	event := amplitude.Event{
		EventType: eventName,
		DeviceID:  deviceID,
	}
	ANALYTICS.Track(event)

	// Flush the event buffer
	ANALYTICS.Flush()
}

func TrackEventWithPropertiesByAuth(eventName string, eventProperties map[string]interface{}, c *gin.Context) bool {

	defer sentry.Recover()

	if os.Getenv("ENVIRONMENT") != "prod" {
		return false
	}

	session := sessions.Default(c)
	user_session := session.Get(configs.Userkey)

	if user_session == nil {
		go TrackEventWithProperties(eventName, eventProperties, c)
	} else {
		go TrackEventWithPropertiesByUserID(eventName, eventProperties, user_session.(string))
	}
	return true
}

func TrackEventWithProperties(eventName string, eventProperties map[string]interface{}, c *gin.Context) {

	defer sentry.Recover()

	deviceID := c.Request.Header.Get("User-Agent")
	fmt.Printf("eventProperties: %v\n", eventProperties)
	// Track events with optional properties
	ANALYTICS.Track(amplitude.Event{
		DeviceID:        deviceID,
		EventType:       eventName,
		EventProperties: eventProperties,
	})

	// Flush the event buffer
	ANALYTICS.Flush()
}

func TrackEventWithPropertiesByUserID(eventName string, eventProperties map[string]interface{}, userID string) {

	defer sentry.Recover()

	internal := isUserInternal(userID)

	eventProperties["internal"] = internal
	fmt.Printf("eventProperties: %v\n", eventProperties)

	// Track events with optional properties
	ANALYTICS.Track(amplitude.Event{
		UserID:          userID,
		EventType:       eventName,
		EventProperties: eventProperties,
	})

	// Flush the event buffer
	ANALYTICS.Flush()
}

func KillClient() {
	// Shutdown the client
	ANALYTICS.Shutdown()
}

func GetTrackingMap(c *gin.Context) map[string]interface{} {
	session := sessions.Default(c)
	user_session := session.Get(configs.Userkey)

	trackingMap := make(map[string]interface{})
	if user_session == nil {
		trackingMap["loggedIn"] = false
		trackingMap["ID"] = c.Request.Header.Get("User-Agent")
	} else {
		trackingMap["loggedIn"] = true
		trackingMap["ID"] = user_session.(string)
	}

	return trackingMap
}

func isUserInternal(userID string) (isInternal bool) {
	user, _ := models.GetUserByID(userID)
	_, ok := sendgridEmail.InternalUsers[user.Phone]
	return ok
}
