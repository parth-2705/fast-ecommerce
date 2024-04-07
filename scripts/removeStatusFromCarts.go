package scripts

import (
	"context"
	"fmt"
	"hermes/db"
	"hermes/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type Cart struct {
	ID          string             `json:"id" bson:"_id"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
	UserAgentID string             `json:"userAgentID" bson:"userAgentID"`
	UserID      string             `json:"userID" bson:"userID"`
	ProductID   string             `json:"productID" bson:"productID"`
	Store       string             `json:"store" bson:"store"`
	VariantID   string             `json:"variantID" bson:"variantID"`
	DealID      string             `json:"dealID" bson:"dealID"`
	CouponID    string             `json:"couponID" bson:"couponID"`
	CartAmount  models.OrderAmount `json:"cartAmount" bson:"cartAmount"`
}

func removeStatusFromCarts() (err error) {
	var oldCarts []Cart
	cursor, err := db.CartCollection.Find(context.Background(), bson.D{})
	if err != nil {
		return
	}

	err = cursor.All(context.Background(), &oldCarts)
	if err != nil {
		return
	}

	fmt.Printf("len(oldCarts): %v\n", len(oldCarts))

	var failed []string

	for i, cart := range oldCarts {

		fmt.Printf("%d : %s\n", i, cart.ID)

		_, err = db.CartCollection.DeleteOne(context.Background(), bson.M{"_id": cart.ID})
		if err != nil {
			failed = append(failed, cart.ID)
			continue
		}

		_, err = db.CartCollection.InsertOne(context.Background(), cart)
		if err != nil {
			fmt.Println(err)
			failed = append(failed, cart.ID)
		}
	}

	fmt.Printf("len(failed): %v\n", len(failed))
	for _, f := range failed {
		fmt.Println(f)
	}

	return
}
