package models

import (
	"fmt"
	"hermes/configs/Mysql"
	"hermes/utils/data"
	"time"

	"gorm.io/datatypes"
)

type DecentroWebHookPayload struct {
	ID        string         `gorm:"column:id"`
	CreatedAt time.Time      `gorm:"column:createdAt"`
	Payload   datatypes.JSON `gorm:"column:payload"`
}

func (DecentroWebHookPayload) TableName() string {
	return "DecentroWebhookLogs"
}

func CreateDecentroWebhookHitLog(payload []byte) (DecentroWebHookPayload, error) {
	newLog := DecentroWebHookPayload{
		ID:      data.GetUUIDString("decentro"),
		Payload: payload,
	}

	err := Mysql.DB.Model(&newLog).Create(&newLog).Error
	if err != nil {
		fmt.Printf("decentro Log err: %v\n", err)
	}

	return newLog, err
}
