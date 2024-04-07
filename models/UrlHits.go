package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"hermes/configs/Mysql"
	"hermes/utils/data"
	"time"
)

type URLParams map[string]string

func (URLParams) GormDataType() string {
	return "json"
}

func (data URLParams) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *URLParams) Scan(value interface{}) error {

	if value == nil {
		return nil
	}

	var byteSlice []byte
	switch v := value.(type) {
	case []byte:
		if len(v) > 0 {
			byteSlice = make([]byte, len(v))
			copy(byteSlice, v)
		}
	case string:
		byteSlice = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	err := json.Unmarshal(byteSlice, &data)
	return err
}

type URLHit struct {
	ID          string    `gorm:"column:id;size:100"`
	CreatedAt   time.Time `gorm:"column:createdAt;index:createdAt"`
	URL         string    `gorm:"column:url;index:url"`
	Method      string    `gorm:"column:method;size:100"`
	Params      URLParams `gorm:"column:params"`
	EventName   string    `gorm:"column:eventName"`
	UserAgentID string    `gorm:"column:userAgentID;size:100"`
	UserID      string    `gorm:"column:userID;size:100"`
	PhoneNumber string    `gorm:"column:phoneNumber;size:100"`
	Internal    bool      `gorm:"column:internal"`
	Headers     URLParams `gorm:"column:headers"`
	StatusCode  int       `gorm:"column:responseStatusCode"`
}

func (URLHit) TableName() string {
	return "urlHit"
}

func CreateURLHit(url, method, eventName, userAgenID, userID, phoneNumber string, params URLParams, internal bool, headers URLParams) (entry URLHit, err error) {
	entry = URLHit{
		ID:          data.GetUUIDString("urlHit"),
		URL:         url,
		Method:      method,
		EventName:   eventName,
		UserAgentID: userAgenID,
		UserID:      userID,
		Params:      params,
		PhoneNumber: phoneNumber,
		Internal:    internal,
		Headers:     headers,
	}

	err = Mysql.DB.Create(&entry).Error
	if err != nil {
		fmt.Printf("mysql err: %v\n", err)
		return
	}
	return
}

func (log URLHit) LogResponseStatus(statusCode int) (err error) {
	log.StatusCode = statusCode
	err = Mysql.DB.Save(&log).Error
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
	return
}
