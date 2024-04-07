package Logs

import (
	"fmt"
	"hermes/configs/Mysql"
	"hermes/utils/data"
	"time"

	"gorm.io/datatypes"
)

type WhatsappMessageLog struct {
	ID                 string
	CreatedAt          *time.Time
	UpdatedAt          *time.Time
	PhoneNumber        string
	TemplateID         string
	ResponseStatusCode int
	Error              string
}

type FBAdsConversionLog struct {
	ID                 string
	CreatedAt          *time.Time
	UpdatedAt          *time.Time
	IPAddress          string
	UserAgent          string
	FBC                string
	FBP                string
	EventName          string
	ActionSource       string
	UserParams         datatypes.JSON
	ResponseStatusCode int
	Response           string
}

func CreateFBAdsConversionLog(eventName string, actionSource string, userParams datatypes.JSON) (FBAdsConversionLog, error) {

	log := FBAdsConversionLog{
		ID:           data.GetUUIDString("conv"),
		EventName:    eventName,
		ActionSource: actionSource,
		UserParams:   userParams,
	}

	err := Mysql.DB.Model(&log).Create(&log).Error
	if err != nil {
		fmt.Printf("err log: %v\n", err)
	}

	return log, err
}

func (log *FBAdsConversionLog) UpdateResponseStatus(statusCode int, errString string) (err error) {
	log.ResponseStatusCode = statusCode
	log.Response = errString
	err = Mysql.DB.Model(log).Save(log).Error
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
	return
}

func CreateWhatsappMessageLog(phoneNumber string, templateID string) (WhatsappMessageLog, error) {

	log := WhatsappMessageLog{
		ID:          data.GetUUIDString("log"),
		PhoneNumber: phoneNumber,
		TemplateID:  templateID,
	}

	err := Mysql.DB.Model(&log).Create(&log).Error
	if err != nil {
		fmt.Printf("err log: %v\n", err)
	}

	return log, err
}

func (log *WhatsappMessageLog) UpdateResponseStatus(statusCode int, errString string) (err error) {
	log.ResponseStatusCode = statusCode
	log.Error = errString
	err = Mysql.DB.Model(log).Save(log).Error
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
	return
}
