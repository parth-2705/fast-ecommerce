package models

import (
	"context"
	"fmt"
	"hermes/db"
	"hermes/utils/data"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type SellerMember struct {
	ID        string    `json:"_id" bson:"_id"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
	Name      string    `json:"name" bson:"name"`
	Phone     string    `json:"phone" bson:"phone"`
	SellerID  string    `json:"sellerID" bson:"sellerID"`
}

func DoesSellerMemberExist(phone string) bool {
	var sellerMember SellerMember
	err := db.SellerMembersCollection.FindOne(context.Background(), bson.M{"phone": phone}).Decode(&sellerMember)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func (sellerMember SellerMember) Create() (SellerMember, error) {
	sellerMember.ID = data.GetUUIDStringWithoutPrefix()
	sellerMember.CreatedAt = time.Now()
	sellerMember.UpdatedAt = time.Now()

	_, err := db.SellerMembersCollection.InsertOne(context.Background(), sellerMember)
	if err != nil {
		fmt.Println(err)
		return sellerMember, err
	}

	return sellerMember, nil
}

func (sellerMember SellerMember) Update() error {
	sellerMember.UpdatedAt = time.Now()
	_, err := db.SellerMembersCollection.ReplaceOne(context.Background(), bson.M{"_id": sellerMember.ID}, sellerMember)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return err
	}

	return nil
}

func GetSellerMemberByPhone(phone string) (SellerMember, error) {
	var sellerMember SellerMember
	err := db.SellerMembersCollection.FindOne(context.Background(), bson.M{"phone": phone}).Decode(&sellerMember)
	if err != nil {
		fmt.Println(fmt.Errorf("error: %s", err.Error()))
		return sellerMember, err
	}
	return sellerMember, nil
}

func GetSellerMemberByID(id string) (SellerMember, error) {
	var sellerMember SellerMember
	err := db.SellerMembersCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&sellerMember)
	if err != nil {
		fmt.Println(fmt.Errorf("error: %s", err.Error()))
		return sellerMember, err
	}
	return sellerMember, nil
}
