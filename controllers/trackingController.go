package controllers

import (
	"fmt"
	"hermes/services/Sentry"
	"hermes/utils/amplitude"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SendTrackingEvent(c *gin.Context) {
	eventName := c.Query("eventName")
	// trackingMap := amplitude.GetTrackingMap(c)
	eventProps := map[string]interface{}{}
	err := c.BindJSON(&eventProps)
	if err != nil {
		Sentry.SendErrorToSentry(c, err, nil)
	}
	fmt.Printf("eventProps: %v\n", eventProps)
	eventSent := amplitude.TrackEventWithPropertiesByAuth(eventName, eventProps, c)
	c.JSON(http.StatusAccepted, gin.H{
		"eventSent": eventSent,
	})
}
