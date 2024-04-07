package models

import (
	"context"
	"fmt"
	"hermes/db"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type BulkAction struct {
	ID                   string      `json:"id" bson:"_id"`
	CreatedAt            time.Time   `json:"createdAt" bson:"createdAt"`
	CSV                  string      `json:"csv" bson:"csv"`
	ActionTakenBy        string      `json:"actionTakenBy" bson:"actionTakenBy"`
	Version              int         `json:"version" bson:"version"`
	ProductsBeforeChange []Product   `json:"products" bson:"products"`
	VariantsBeforeChange []Variation `json:"variants" bson:"variants"`
}

func (bulkAction BulkAction) CreateIndexes() error {
	indexModels := []mongo.IndexModel{}

	versionModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "version", Value: 1},
		},
	}
	indexModels = append(indexModels, versionModel)

	indexName, err := db.BulkAction.Indexes().CreateMany(context.Background(), indexModels)
	if err != nil {
		if strings.Contains(err.Error(), "Index with name") && strings.Contains(err.Error(), "already exists") {
			fmt.Println("Indexes already exist")
		} else {
			// Handle other errors
			fmt.Println("Error creating indexes:", err)
			return err
		}
	} else {
		fmt.Println("Created index:", indexName)
	}

	return nil
}

func (bulkAction *BulkAction) Create() error {
	bulkAction.CreatedAt = time.Now()
	_, err := db.BulkAction.InsertOne(context.Background(), bulkAction)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func GetBulkActionVersion() (int64, error) {
	count, err := db.BulkAction.CountDocuments(context.TODO(), bson.M{})
	if err != nil {
		return count, err
	}
	return count, nil
}
