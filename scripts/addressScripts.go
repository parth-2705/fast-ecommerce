package scripts

import (
	"fmt"
	"hermes/models"
)

func makeOnlyAddressDefault() (err error) {

	failed, updated, unaffected, total := 0, 0, 0, 0
	users, err := models.GetAllUsers()
	if err != nil {
		return nil
	}
	for _, user := range users {
		total++
		addresses, err := user.GetAddresses()
		if err != nil {
			failed++
			continue
		}
		if len(addresses) == 1 {
			if addresses[0].IsDefault {
				unaffected++
				continue
			}
			addresses[0].IsDefault = true
			err := addresses[0].UpdateInDB()
			if err != nil {
				failed++
				continue
			}
			updated++
		} else {
			unaffected++
		}
	}
	fmt.Printf("failed: %v\n", failed)
	fmt.Printf("updated: %v\n", updated)
	fmt.Printf("unaffected: %v\n", unaffected)
	fmt.Printf("total: %v\n", total)
	return
}
