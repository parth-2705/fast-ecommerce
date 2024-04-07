package scripts

import (
	"context"
	"fmt"
	"hermes/db"
	"hermes/models"

	"go.mongodb.org/mongo-driver/bson"
)

func fillCartInOrders() (err error) {

	var orders []models.Order

	cursor, _ := db.OrderCollection.Find(context.TODO(), bson.D{})
	err = cursor.All(context.Background(), &orders)
	if err != nil {
		return
	}

	failedCount := 0
	newCartCreated := 0
	failedOrder := []string{}

	fmt.Printf("len(orders): %v\n", len(orders))

	for _, order := range orders {

		if order.Cart.ID != "" {
			continue
		}

		cart, err := models.GetCart(order.CartID)
		if err != nil {
			newCartCreated++
			items := []models.Item{{
				ProductID: order.Product.ID,
				VariantID: order.VariantID,
				Product:   order.Product,
				Variant:   order.Variant,
			}}
			cart, err = models.CreateNewCart2(order.UserID, "", items, models.Scripts)
			if err != nil {
				failedCount++
				failedOrder = append(failedOrder, order.ID)
				continue
			}
			cart.Status = models.Dummy
			order.CartID = cart.ID
		}

		cart.Items = []models.Item{
			{
				ProductID: order.Product.ID,
				Product:   order.Product,
				VariantID: order.Variant.ID,
				Variant:   order.Variant,
				Quantity:  1,
			},
		}

		cart.CalculateCartAmount()
		cart.Update()

		order.Cart = cart
		order.Update()
	}

	fmt.Printf("newCartCreated: %v\n", newCartCreated)
	fmt.Printf("failedCount: %v\n", failedCount)
	fmt.Println("Failed Cart Creations")

	for _, id := range failedOrder {
		fmt.Printf("id: %v\n", id)
	}

	return
}
