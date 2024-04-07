package scripts

import (
	"context"
	"fmt"
	"hermes/configs/Mysql"
	"hermes/db"
	"hermes/models"
	"hermes/services/shiprocket"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func addOrdersToOrderLogs() (err error) {
	var orders []models.Order
	ordersCursor, err := db.OrderCollection.Find(context.Background(), bson.M{})

	if err != nil {
		fmt.Println(err)
		return
	}

	err = ordersCursor.All(context.Background(), &orders)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, order := range orders {
		err = Mysql.DB.FirstOrCreate(&order).Error
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	return
}

func addShiprocketOrdersToOrderLogs() (err error) {
	var srOrders []models.FullShiprocketOrder
	ordersCursor, err := db.ShiprocketOrderCollection.Find(context.Background(), bson.M{})

	if err != nil {
		fmt.Println(err)
		return
	}

	err = ordersCursor.All(context.Background(), &srOrders)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, srOrder := range srOrders {
		err = Mysql.DB.FirstOrCreate(&srOrder).Error
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	return
}

func addShippingsToLogs() (err error) {
	var shippings []models.Shipping
	ordersCursor, err := db.ShippingCollection.Find(context.Background(), bson.M{})

	if err != nil {
		fmt.Println(err)
		return
	}

	err = ordersCursor.All(context.Background(), &shippings)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, shipping := range shippings {
		err = Mysql.DB.FirstOrCreate(&shipping).Error
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	return
}

func populateShippingCharges() (err error) {

	var shippings []models.Shipping
	ordersCursor, err := db.ShippingCollection.Find(context.Background(), bson.M{})

	if err != nil {
		panic(err)
	}

	err = ordersCursor.All(context.Background(), &shippings)
	if err != nil {
		panic(err)
	}

	for _, shipping := range shippings {
		_, err = models.GetShippingChargesByOrderID(fmt.Sprint(shipping.OrderId))
		if err != nil {
			if err == mongo.ErrNoDocuments {
				_, err = shiprocket.GetOrderShippingCharges(shipping.Id, fmt.Sprint(shipping.OrderId))
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println("Order ID: ", shipping.Id)
				}
			}
		}
	}

	return
}
