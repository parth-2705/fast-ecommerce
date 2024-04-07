package common

import (
	"context"
	"hermes/db"
	"hermes/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func FilterProduct(searchTerm string, filters bson.M) ([]models.Product, error) {

	var products []models.Product

	if searchTerm != "" {
		filters["$text"] = bson.M{"$search": searchTerm, "$caseSensitive": false}
	}
	opts := options.Find()
	opts.SetLimit(20)

	cur, err := db.ProductCollection.Find(context.Background(), filters, opts)
	if err != nil {
		return products, err
	}
	err = cur.All(context.Background(), &products)
	if err != nil {
		return products, err
	}

	return products, nil

}
