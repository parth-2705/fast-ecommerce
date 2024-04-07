package Mysql

import (
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() (err error) {

	DNS := os.Getenv("MYSQL_CONNECTION_STRING")
	// init db with connection pool settings
	DB, err = gorm.Open(mysql.Open(DNS), &gorm.Config{})
	if err != nil {
		return err
	}

	sql, err := DB.DB()
	if err != nil {
		return err
	}
	sql.SetConnMaxIdleTime(time.Hour)
	sql.SetMaxOpenConns(100)

	return
}
