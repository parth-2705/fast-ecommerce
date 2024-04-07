package scripts

import (
	"context"
	"fmt"
	"hermes/configs/Mysql"
	"hermes/db"
	"hermes/models"

	"go.mongodb.org/mongo-driver/bson"
)

func exportCartsToMySql() error {

	// Get all Carts from Mongo
	cursor, err := db.CartCollection.Find(context.Background(), bson.D{})
	if err != nil {
		return err
	}
	defer cursor.Close(context.Background())

	var carts []models.Cart
	err = cursor.All(context.Background(), &carts)
	if err != nil {
		fmt.Printf("Cart Reading err: %v\n", err)
		return err
	}

	// Push all Carts to Mysql
	err = Mysql.DB.CreateInBatches(&carts, 50).Error
	if err != nil {
		fmt.Printf("Cart Wrtigin err: %v\n", err)
		return err
	}

	return nil
}

func exportBackwardShipmentsToMySql() (err error) {
	cursor, err := db.BackwardShipmentCollection.Find(context.Background(), bson.D{})
	if err != nil {
		return err
	}
	defer cursor.Close(context.Background())

	var backwardShipments []models.BackwardShipment
	err = cursor.All(context.Background(), &backwardShipments)
	if err != nil {
		fmt.Printf("Cart Reading err: %v\n", err)
		return err
	}

	// Push all backward shipments to Mysql
	err = Mysql.DB.CreateInBatches(&backwardShipments, 50).Error
	if err != nil {
		fmt.Printf("Cart Wrtigin err: %v\n", err)
		return err
	}

	return nil
}
