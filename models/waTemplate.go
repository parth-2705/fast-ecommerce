package models

import (
	"context"
	"fmt"
	"hermes/db"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type WATemplate struct {
	ProductID      string   `json:"productID" bson:"productID"`
	TemplateID     string   `json:"templateID" bson:"templateID"`
	SleepTime      int64    `json:"sleepTime" bson:"sleepTime"`
	BodyParameters []string `json:"bodyParametes" bson:"bodyParamters"`
	ImageLink      string   `json:"imageLink" bson:"imageLink"`
}

func (template WATemplate) CreateIndexes() error {
	t := true
	options := options.Index()
	options.Unique = &t
	indexModels := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "productID", Value: 1},
			},
			Options: options,
		},
	}

	indexName, err := db.WATemplatesCollection.Indexes().CreateMany(context.Background(), indexModels)
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

func GetTemplateConfig(productID string) (template WATemplate, err error) {

	err = db.WATemplatesCollection.FindOne(context.Background(), bson.M{"productID": productID}).Decode(&template)
	if err != nil {
		fmt.Println("template err:", err)
	}

	return
}
