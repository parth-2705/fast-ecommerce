package models

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"hermes/db"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Deal struct {
	ID           string    `json:"id" bson:"_id"`
	CreatedAt    time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt" bson:"updatedAt"`
	ProductID    string    `json:"productID" bson:"_productID"`
	StartsAt     time.Time `json:"startsAt" bson:"startsAt"`
	EndsAt       time.Time `json:"endsAt" bson:"endsAt"`
	DealPrice    float64   `json:"dealPrice" bson:"dealPrice"`
	AdminPrice   float64   `json:"adminPrice" bson:"adminPrice"`
	MemberPrice  float64   `json:"memberPrice" bson:"memberPrice"`
	Quantity     int       `json:"quantity" bson:"quantity"` //max number of deals to be given
	DealsAvailed int       `json:"dealsAvailed" bson:"dealsAvailed"`
	Active       bool      `json:"active" bson:"active"`
	IsTeamDeal   bool      `json:"isTeamDeal" bson:"isTeamDeal"`
	TeamCapacity int       `json:"teamCapacity" bson:"teamCapacity"`
}

func (Deal) GormDataType() string {
	return "json"
}

func (data Deal) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *Deal) Scan(value interface{}) error {

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

// Has Deal Expired
func DealExpired(deal Deal) bool {
	duration := deal.EndsAt.Sub(deal.StartsAt)
	return duration > 0
}

// Has Deal Expired
func (deal Deal) DealExpired() bool {
	duration := deal.EndsAt.Sub(deal.StartsAt)
	return duration > 0
}

// Create New Deal
func CreateDeal(deal Deal) error {
	deal.CreatedAt = time.Now()
	deal.UpdatedAt = time.Now()
	_, err := db.DealCollection.InsertOne(context.Background(), deal)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func GetDealByProductID(productID string) (Deal, error) {
	var deal Deal
	err := db.DealCollection.FindOne(context.Background(), bson.M{"_productID": productID}).Decode(&deal)
	return deal, err
}

func GetAllDeals() ([]Deal, error) {
	var deals []Deal
	cur, err := db.DealCollection.Find(context.Background(), bson.M{}, options.Find())
	if err != nil {
		return deals, err
	}
	err = cur.All(context.Background(), &deals)
	if err != nil {
		return deals, err
	}
	return deals, nil
}

func ReplaceDeal(newDeal Deal) error {
	_, err := db.DealCollection.ReplaceOne(context.Background(), bson.M{"productID": newDeal.ID}, newDeal)
	return err
}

func DealActiveStatusUpdate(dealID string, status bool) error {
	_, err := db.DealCollection.UpdateOne(context.Background(), bson.M{"dealID": dealID}, bson.M{"$set": bson.M{"active": status}})
	return err
}

func GetDealByID(dealID string) (Deal, error) {
	var deal Deal
	err := db.DealCollection.FindOne(context.Background(), bson.M{"_id": dealID}).Decode(&deal)
	return deal, err
}
