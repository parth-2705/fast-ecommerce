package models

import (
	"context"
	"fmt"
	"hermes/db"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Pincode struct {
	// pincode is indexed
	ID      string `json:"id" bson:"_id"`
	Pincode string `bson:"pincode" json:"pincode"`
	City    string `bson:"city" json:"city"`
	State   string `bson:"state" json:"state"`
}

func (pincode Pincode) CreateIndexes() error {
	stateModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "state", Value: 1},
		},
	}

	indexName, err := db.PincodeCollection.Indexes().CreateOne(context.Background(), stateModel)
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

func GetPincodesForAListOfCities(cities []string) ([]Pincode, error) {
	var pincodes []Pincode
	cur, err := db.PincodeCollection.Aggregate(context.Background(), mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"state": bson.M{"$in": cities}}}},
	})
	if err != nil {
		return pincodes, err
	}
	err = cur.All(context.Background(), &pincodes)
	if err != nil {
		return pincodes, err
	}

	return pincodes, nil
}

func GetListOfAllStates() ([]interface{}, error) {
	var states []interface{}
	states, err := db.PincodeCollection.Distinct(context.Background(), "state", bson.M{})
	if err != nil {
		return states, err
	}

	return states, nil
}
