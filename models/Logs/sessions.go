package Logs

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"hermes/configs/Mysql"
	"hermes/utils/data"
	"time"
)

type Session struct {
	ID        string
	CreatedAt *time.Time
	UpdatedAt *time.Time
	UserAgent string
	IPAddress string
	FBC       string
	FBP       string
}

func (Session) GormDataType() string {
	return "json"
}

func (data Session) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *Session) Scan(value interface{}) error {

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

func CreateNewSession(userAgent string, ipAddress string) (Session, error) {

	session := Session{
		ID:        data.GetUUIDString("session"),
		UserAgent: userAgent,
		IPAddress: ipAddress,
	}

	err := Mysql.DB.Create(&session).Error
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}

	return session, err
}
