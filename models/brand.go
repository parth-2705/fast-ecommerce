package models

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"hermes/configs/Redis"
	"hermes/db"
	"strings"
	"time"

	"html/template"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Brand struct {
	ID               string        `json:"id" bson:"_id"`
	CreatedAt        time.Time     `json:"createdAt" bson:"createdAt"`
	UpdatedAt        time.Time     `json:"updatedAt" bson:"updatedAt"`
	SellerID         string        `json:"sellerID" bson:"sellerID"`
	Name             string        `json:"name" bson:"name"`
	Description      template.HTML `json:"description" bson:"description"`
	Logo             string        `json:"logo" bson:"logo"`
	Features         []string      `json:"features" bson:"features"`
	SelfFulfilled    bool          `json:"selfFulfilled" bson:"selfFulfilled"`
	ServicableCities []string      `json:"servicableCities" bson:"servicableCities"`
}

func (Brand) GormDataType() string {
	return "json"
}

func (data Brand) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *Brand) Scan(value interface{}) error {

	if value == nil {
		return nil
	}

	var byteSlice []byte
	switch v := value.(type) {
	case []byte:
		if len(v) > 0 {
			byteSlice = make([]byte, len(v))
			copy(byteSlice, v)
		}
	case string:
		byteSlice = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	err := json.Unmarshal(byteSlice, &data)
	return err
}

func (brand Brand) CreateIndexes() error {
	sellerIDModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "sellerID", Value: 1},
		},
	}

	indexName, err := db.BrandsCollection.Indexes().CreateOne(context.Background(), sellerIDModel)
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

func (brand *Brand) Create() error {
	brand.CreatedAt = time.Now()
	brand.UpdatedAt = time.Now()

	if len(brand.ServicableCities) > 0 {
		pincodes, err := GetPincodesForAListOfCities(brand.ServicableCities)
		if err != nil {
			return fmt.Errorf("Unable to read cities" + err.Error())
		}

		var servicablePincodes []string

		for _, pincode := range pincodes {
			servicablePincodes = append(servicablePincodes, pincode.Pincode)
		}

		err = Redis.SetPincodesToBrand(brand.ID, servicablePincodes)
		if err != nil {
			return fmt.Errorf("Unable to set pincodes to redis" + err.Error())
		}
	}

	_, err := db.BrandsCollection.InsertOne(context.Background(), brand)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (brand *Brand) Update() (*mongo.UpdateResult, error) {
	brand.UpdatedAt = time.Now()

	if len(brand.ServicableCities) > 0 {
		pincodes, err := GetPincodesForAListOfCities(brand.ServicableCities)
		if err != nil {
			return &mongo.UpdateResult{}, fmt.Errorf("Unable to read cities" + err.Error())
		}

		var servicablePincodes []string

		for _, pincode := range pincodes {
			servicablePincodes = append(servicablePincodes, pincode.Pincode)
		}

		err = Redis.DeleteKeyFromRedis(brand.ID)
		if err != nil {
			return &mongo.UpdateResult{}, fmt.Errorf("Unable to rem pincodes from redis" + err.Error())
		}

		err = Redis.SetPincodesToBrand(brand.ID, servicablePincodes)
		if err != nil {
			return &mongo.UpdateResult{}, fmt.Errorf("Unable to set pincodes to redis" + err.Error())
		}
	}

	result, err := db.BrandsCollection.UpdateOne(context.Background(), bson.M{"_id": brand.ID}, bson.M{"$set": brand})
	if err != nil {
		return result, err
	}
	return result, nil
}

func GetBrand(id string) (brand Brand, err error) {
	err = db.BrandsCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&brand)
	if err != nil {
		return
	}
	return
}

func GetBrandsList() ([]Brand, error) {
	var brands []Brand
	cur, err := db.BrandsCollection.Find(context.Background(), bson.D{})
	if err != nil {
		return brands, err
	}
	err = cur.All(context.Background(), &brands)
	if err != nil {
		return brands, err
	}

	return brands, nil
}
