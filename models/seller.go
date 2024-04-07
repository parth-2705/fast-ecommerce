package models

import (
	"context"
	"fmt"
	"hermes/db"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Seller struct {
	ID               string    `json:"id" bson:"_id"`
	CreatedAt        time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt" bson:"updatedAt"`
	Name             string    `json:"name" bson:"name" form:"name"`
	Phone            string    `json:"phone" bson:"phone" form:"phone"`
	Email            string    `json:"email" bson:"email" form:"email"`
	PinCode          string    `json:"pinCode" bson:"pinCode" form:"pinCode"`
	HouseArea        string    `json:"houseArea" bson:"houseArea" form:"houseArea"`
	StreetName       string    `json:"streetName" bson:"streetName" form:"streetName"`
	City             string    `json:"city" bson:"city" form:"city"`
	State            string    `json:"state" bson:"state" form:"state"`
	ProfileCompleted bool      `json:"profileCompleted" bson:"profileCompleted" form:"profileCompleted"`
	GSTIN            string    `json:"GSTIN" bson:"GSTIN" form:"GSTIN"`
	PAN              string    `json:"PAN" bson:"PAN" form:"PAN"`
}

func (seller Seller) CreateIndexes() error {
	indexModels := []mongo.IndexModel{}

	phoneModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "phone", Value: 1},
		},
		// Options: options.Index().SetUnique(true),
	}
	indexModels = append(indexModels, phoneModel)

	indexName, err := db.CartCollection.Indexes().CreateMany(context.Background(), indexModels)
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

func (seller *Seller) Create() error {
	seller.CreatedAt = time.Now()
	seller.UpdatedAt = time.Now()

	_, err := db.SellerCollection.InsertOne(context.Background(), seller)
	if err != nil {
		fmt.Println(err)
		return err
	}

	var sellerMember SellerMember = SellerMember{
		ID:       seller.ID,
		Phone:    seller.Phone,
		Name:     seller.Name,
		SellerID: seller.ID,
	}
	sellerMember.Create()

	return nil
}

func (seller Seller) Update() error {
	seller.UpdatedAt = time.Now()
	_, err := db.SellerCollection.ReplaceOne(context.Background(), bson.M{"_id": seller.ID}, seller)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return err
	}
	return nil
}

func DoesSellerExists(phone string) bool {
	var seller Seller
	err := db.SellerCollection.FindOne(context.Background(), bson.M{"phone": phone}).Decode(&seller)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func GetSellerByPhone(phone string) (Seller, error) {
	var seller Seller
	err := db.SellerCollection.FindOne(context.Background(), bson.M{"phone": phone}).Decode(&seller)
	if err != nil {
		fmt.Println(fmt.Errorf("error: %s", err.Error()))
		return seller, err
	}
	return seller, nil
}

func GetSellerByID(id string) (Seller, error) {
	var seller Seller
	err := db.SellerCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&seller)
	if err != nil {
		fmt.Println(fmt.Errorf("error in GetSellerByID: %s", err.Error()))
		return seller, err
	}
	return seller, nil
}

func GetSellerMapByID(id string) (map[string]interface{}, error) {
	var seller map[string]interface{}
	err := db.SellerCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&seller)
	if err != nil {
		fmt.Println(fmt.Errorf("error: %s", err.Error()))
		return seller, err
	}
	return seller, nil
}

func GetSellersList() ([]Seller, error) {
	var sellers []Seller

	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cur, err := db.SellerCollection.Find(context.Background(), bson.D{}, opts)
	if err != nil {
		return sellers, err
	}
	err = cur.All(context.Background(), &sellers)
	if err != nil {
		return sellers, err
	}

	return sellers, nil
}
