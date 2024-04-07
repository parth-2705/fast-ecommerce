package scripts

import (
	"context"
	"fmt"
	"hermes/db"
	"hermes/models"
	"hermes/services/shiprocket"

	"go.mongodb.org/mongo-driver/bson"
)

func updateShippingParent() (err error) {
	var shippings []models.Shipping
	cur, err := db.ShippingCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return
	}
	err = cur.All(context.Background(), &shippings)
	if err != nil {
		return
	}
	for idx, val := range shippings {
		fmt.Println("idx:", idx)
		var temp = val
		temp.ParentRelation = 0
		_, err = db.ShippingCollection.ReplaceOne(context.Background(), bson.M{"_id": temp.Id}, temp)
		if err != nil {
			return
		}
	}
	return
}

func createReturn(orderID string) (err error) {
	quantityInfo := map[string]shiprocket.QuantityInfo{}
	err = shiprocket.CreateReturnUtil(orderID, quantityInfo)
	return
}
