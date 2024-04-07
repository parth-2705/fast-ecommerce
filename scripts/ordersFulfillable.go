package scripts

import (
	"fmt"
	"hermes/models"
	"time"
)

func markAllOldPaidOrdersAsFulfillable() (err error) {
	orders, err := models.GetAllOrdersInDB()
	if err != nil {
		return
	}
	releaseTime := time.Now().Add(time.Hour * -4)
	total, failed, passed := 0, 0, 0
	for _, order := range orders {
		if order.PaymentStatus == "Paid" && order.CreatedAt.Before(releaseTime) {
			total++
			err = order.SetFullfillable(true)
			if err != nil {
				failed++
				continue
			}
			passed++
		}
	}
	fmt.Printf("total: %v\n", total)
	fmt.Printf("failed: %v\n", failed)
	fmt.Printf("passed: %v\n", passed)
	return
}
