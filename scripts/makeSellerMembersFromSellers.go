package scripts

import (
	"context"
	"hermes/db"
	"hermes/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func makeSellerMembersFromSellers() (err error) {
	var sellerMembers []interface{}
	var sellers []models.Seller

	cur, err := db.SellerCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return
	}
	err = cur.All(context.Background(), &sellers)
	if err != nil {
		return
	}
	for _, val := range sellers {
		var tempSellerMember models.SellerMember = models.SellerMember{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			ID:        val.ID,
			Name:      val.Name,
			Phone:     val.Phone,
			SellerID:  val.ID,
		}
		sellerMembers = append(sellerMembers, tempSellerMember)
	}
	_, err = db.SellerMembersCollection.InsertMany(context.Background(), sellerMembers)
	return
}
