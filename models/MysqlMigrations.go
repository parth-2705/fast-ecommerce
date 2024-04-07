package models

import (
	"hermes/configs/Mysql"
	"hermes/models/Logs"
)

func MysqlMigrations() (err error) {

	modelsToMigrate := []interface{}{
		URLHit{},
		OTPLog{},
		Order{},
		AdminSellerLogs{},
		DecentroWebHookPayload{},
		CouponUseRecord{},
		FullShiprocketOrder{},
		Shipping{},
		Logs.WhatsappMessageLog{},
		Logs.Session{},
		Logs.FBAdsConversionLog{},
		Cart{},
		BackwardShipment{},
		Ambassdor{},
		WalletTransaction{},
		ReferralRecord{},
	}

	// Migrate Models one by one and throw error, if any migration fails.
	for i := range modelsToMigrate {
		err = Mysql.DB.AutoMigrate(modelsToMigrate[i])
		if err != nil {
			return err
		}
	}
	return
}
