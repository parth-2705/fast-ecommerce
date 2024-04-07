package scripts

import (
	"fmt"
	"hermes/configs/Mysql"
	"hermes/models"
)

func addRepeatUserKeyToUsers() error {

	// Get all Users

	ordersFetchError := 0
	mysqlErr := 0
	mongoErr := 0

	mysqlErrList := make([]string, 0)
	mongoErrList := make([]string, 0)

	users, err := models.GetAllUsers()
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return err
	}

	fmt.Printf("len(users): %v\n", len(users))

	// Get all Orders for this User
	for _, user := range users {
		orders, err := user.GetAllOrders()
		if err != nil {
			ordersFetchError++
			continue
		}

		// If order length Zero, mark user as new User
		// Continue
		// if len(orders) == 0 {
		// 	user.RepeatUser = false
		// 	user.Update()
		// 	continue
		// }

		//  Mark User accordinglt
		user.RepeatUser = false

		// Iterate through Orders
		for _, order := range orders {

			// Replaec with new User Object
			order.User = user

			// Save Order in Mongo
			err = order.Update()
			if err != nil {
				mongoErr++
				mongoErrList = append(mongoErrList, order.ID)
			}

			// Save Order in MySql
			err = Mysql.DB.Model(&order).Save(&order).Error
			if err != nil {
				mysqlErr++
				mysqlErrList = append(mysqlErrList, order.ID)
			}

			if order.PaymentStatus == "Paid" {
				user.RepeatUser = true
			}
		}

		user.Update()
	}

	fmt.Printf("ordersFetchError: %v\n", ordersFetchError)
	fmt.Println("")
	fmt.Printf("mysqlErr: %v\n", mysqlErr)
	fmt.Printf("mysqlErrList: %+v\n", mysqlErrList)
	fmt.Println("")
	fmt.Printf("mongoErr: %v\n", mongoErr)
	fmt.Printf("mongoErrList: %v\n", mongoErrList)

	return nil

}
