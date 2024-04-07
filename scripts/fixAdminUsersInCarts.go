package scripts

import (
	"context"
	"fmt"
	"hermes/db"
	"hermes/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func fixAdminUsersInCarts() error {

	filterTime := time.Now().AddDate(0, 0, -6)

	cursor, err := db.OrderCollection.Find(context.Background(), bson.M{"createdAt": bson.M{"$gte": filterTime}})
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return err
	}
	var orders []models.Order
	err = cursor.All(context.Background(), &orders)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return err
	}

	fmt.Printf("len(orders): %v\n", len(orders))
	success := 0
	for _, order := range orders {
		userID := order.UserID

		cart, err := models.GetCart(order.CartID)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			continue
		}

		cart.UserID = userID
		err = cart.Update()
		if err != nil {
			fmt.Printf("err: %v\n", err)
			continue
		}

		order.Cart = cart
		err = order.Update()
		if err != nil {
			fmt.Printf("err: %v\n", err)
			continue
		}

		success++
	}

	fmt.Printf("success: %v\n", success)
	return nil
}
