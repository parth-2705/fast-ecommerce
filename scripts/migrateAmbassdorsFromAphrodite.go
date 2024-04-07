package scripts

import (
	"encoding/json"
	"fmt"
	"hermes/models"
	"os"
	"time"

	"gorm.io/datatypes"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type CampusAmbassdorResponses struct {
	gorm.Model
	Responses datatypes.JSON `json:"responses"`
	UserPhone string         `json:"userPhone" gorm:"primaryKey"`
	Name      string         `json:"name"`
	City      string         `json:"city"`
}

func makeAmbassdorsForWebsiteReferrals() error {

	ambassdors, err := models.GetAllAmbassdors()
	if err != nil {
		return err
	}

	failed := 0
	success := 0

	for _, ambassdorUser := range ambassdors {

		_, err := models.GetAmbassdorByUserID(ambassdorUser.ID)
		if err != nil {

			profile, err := ambassdorUser.GetProfile()
			if err != nil {
				failed++
				continue
			}

			_, err = ambassdorUser.MakeAmbassdor("", []string{}, profile.ReferralCode, 0)
			if err != nil {
				failed++
				continue
			}
			success++
		}
	}

	fmt.Printf("failed: %v\n", failed)
	fmt.Printf("success: %v\n", success)

	return nil
}

func migrateAmbassdorsFromAphrodite() error {

	DNS := os.Getenv("WHATSAPP_DB")

	// connect to Aphrodite DB
	aphro, err := gorm.Open(mysql.Open(DNS), &gorm.Config{})
	if err != nil {
		return err
	}

	sql, err := aphro.DB()
	if err != nil {
		return err
	}
	defer sql.Close()
	sql.SetConnMaxIdleTime(time.Hour)
	sql.SetMaxOpenConns(100)

	// Get all Responses
	var caResponses []CampusAmbassdorResponses
	err = aphro.Model(&caResponses).Find(&caResponses).Error
	if err != nil {
		return err
	}

	fmt.Printf("len(caResponses): %v\n", len(caResponses))

	userNotFound := 0
	success := 0
	failed := 0
	for _, response := range caResponses {

		user, err := models.GetUser("+" + response.UserPhone)
		if err != nil {
			userNotFound++
			continue
		}

		user.HasJoinedReferralProgram = false

		var answers []string
		json.Unmarshal(response.Responses, &answers)

		err = user.JoinReferralProgram(response.Name, response.City, answers, 0)
		if err != nil {
			failed++
			continue
		}

		success++
	}

	fmt.Printf("userNotFound: %v\n", userNotFound)
	fmt.Printf("failed: %v\n", failed)
	fmt.Printf("success: %v\n", success)

	return nil
}
