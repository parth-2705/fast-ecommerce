package tmpl

import (
	"hermes/utils/amplitude"
	"os"
)

func TrackAmplitudeEvent(eventName string, trackingMap map[string]interface{}) bool {
	if os.Getenv("ENVIRONMENT") != "prod" {
		return false
	}

	if len(trackingMap) == 0 {
		return true
	}

	if trackingMap["loggedIn"] == nil || trackingMap["ID"] == nil {
		return true
	}

	if trackingMap["loggedIn"].(bool) {
		AmplitudeTrackUser(eventName, trackingMap["ID"].(string))
	} else {
		AmplitudeTrackDevice(eventName, trackingMap["ID"].(string))
	}
	return true
}

func AmplitudeTrackUser(eventName string, userID string) {
	go amplitude.TrackEventByUserID(eventName, userID)
}

func AmplitudeTrackDevice(eventName string, deviceID string) {
	go amplitude.TrackEventByDeviceID(eventName, deviceID)
}
