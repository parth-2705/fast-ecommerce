package scripts

import (
	"context"
	"fmt"
	"hermes/db"
	"hermes/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func addCommissionToOrders() (err error) {
	var orders []models.Order
	cur, err := db.OrderCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return
	}
	err = cur.All(context.Background(), &orders)
	if err != nil {
		return
	}

	for idx, val := range orders {
		var commission models.Commission
		err = db.CommissionCollection.FindOne(context.Background(), bson.M{"brandID": val.Cart.Items[0].Product.BrandID}).Decode(&commission)
		if err == mongo.ErrNoDocuments {
			fmt.Println("skipping ", idx)
			val.Commission = 0
			_, err := db.OrderCollection.ReplaceOne(context.Background(), bson.M{"_id": val.ID}, val)
			if err != nil {
				return fmt.Errorf("error in idx:%d %s", idx, err)
			}
			continue
		}
		if err != nil && err != mongo.ErrNoDocuments {
			return fmt.Errorf("error in idx:%d %s", idx, err)
		}
		if val.CreatedAt.After(commission.From) || val.CreatedAt.Equal(commission.From) {
			fmt.Println("Updating ", idx, val.ID)
			val.Commission = commission.Commission
		} else {
			val.Commission = 0
		}
		_, err := db.OrderCollection.ReplaceOne(context.Background(), bson.M{"_id": val.ID}, val)
		if err != nil {
			return fmt.Errorf("error in idx:%d %s", idx, err)
		}
	}
	return nil
}
