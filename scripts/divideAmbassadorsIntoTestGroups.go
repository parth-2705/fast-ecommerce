package scripts

import (
	"fmt"
	"hermes/configs/Mysql"
	"hermes/models"
)

func divideAmbassadorsIntoTestGroups() error {

	// Get Ambassdors that have been sent a deal
	var dealSentAmbassador []models.Ambassdor
	err := Mysql.DB.Model(&dealSentAmbassador).Where("deals_sent > ?", 0).Find(&dealSentAmbassador).Error
	if err != nil {
		fmt.Printf(" deals sent err: %v\n", err)
		return err
	}

	fmt.Printf("len(dealSentAmbassador): %v\n", len(dealSentAmbassador))
	sentFailed := 0

	// Iterate over them and push them in group 1
	for _, ambassador := range dealSentAmbassador {
		ambassador.TestGroup = 1
		err = Mysql.DB.Save(&ambassador).Error
		if err != nil {
			sentFailed++
			continue
		}
	}

	fmt.Printf("sentFailed: %v\n", sentFailed)

	// Get all other ambassadors
	var allAmbasssadors []models.Ambassdor
	err = Mysql.DB.Model(&allAmbasssadors).Find(&allAmbasssadors).Error
	if err != nil {
		fmt.Printf(" deals sent err: %v\n", err)
		return err
	}

	fmt.Printf("len(allAmbasssadors): %v\n", len(allAmbasssadors))
	sentFailed = 0
	// Put half of them in group 2
	//and Rest half in group 3
	for i, ambassador := range allAmbasssadors {
		if ambassador.TestGroup != 0 {
			continue
		}
		group := 2 + i%2
		ambassador.TestGroup = group
		err = Mysql.DB.Save(&ambassador).Error
		if err != nil {
			sentFailed++
			continue
		}
	}

	fmt.Printf("sentFailed: %v\n", sentFailed)

	return nil
}
