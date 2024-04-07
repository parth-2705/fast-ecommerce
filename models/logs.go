package models

import (
	"hermes/configs"
	"hermes/configs/Mysql"
	"time"

	"gorm.io/datatypes"
)

type AdminSellerLogs struct {
	ID            string            `json:"id"`
	Portal        PortalType        `json:"portal"`
	UserID        string            `json:"userId"`
	RequestMethod string            `json:"requestMethod"`
	Endpoint      string            `json:"endpoint"`
	Body          string            `json:"body"`
	Params        datatypes.JSONMap `json:"json"`
	StatusCode    string            `json:"statusCode"`
	CreatedAt     time.Time         `json:"createdAt"`
}

type PortalType string

const (
	AdminPortal  PortalType = configs.AdminHeader
	SellerPortal PortalType = configs.SellerHeader
	BackwardShip PortalType = "backwardShipment"
	Website      PortalType = "website"
	Scripts      PortalType = "script"
	Whatsapp     PortalType = "whatsapp"
)

func (log AdminSellerLogs) Create() (err error) {
	err = Mysql.DB.Create(&log).Error
	return
}
