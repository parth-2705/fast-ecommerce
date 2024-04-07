package scripts

import (
	"context"
	"fmt"
	"hermes/db"
	"hermes/models"

	"go.mongodb.org/mongo-driver/bson"
)

func removeStreetNameFromAddress() (err error) {
	var addresses []models.Address
	var orders []models.Order
	cur, err := db.AddressCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return
	}
	err = cur.All(context.Background(), &addresses)
	if err != nil {
		return
	}
	for _, val := range addresses {
		var temp = val
		temp.HouseArea = val.HouseArea + " , " + val.StreetName
		temp.StreetName = ""
		_, err = db.AddressCollection.ReplaceOne(context.Background(), bson.M{"_id": temp.ID}, temp)
		if err != nil {
			return
		}
		fmt.Println("_id: ", temp.ID)
	}
	fmt.Println("Address Done")
	cur, err = db.OrderCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return
	}
	err = cur.All(context.Background(), &orders)
	if err != nil {
		return
	}
	for _, val := range orders {
		var temp = val
		temp.Address.HouseArea = val.Address.HouseArea + " , " + val.Address.StreetName
		temp.Address.StreetName = ""
		_, err = db.OrderCollection.ReplaceOne(context.Background(), bson.M{"_id": temp.ID}, temp)
		if err != nil {
			return
		}
		fmt.Println("_id: ", temp.ID)
	}
	return
}
