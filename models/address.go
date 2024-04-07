package models

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"hermes/db"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Address struct {
	ID         string    `json:"id" bson:"_id" gorm:"column:id"`
	CreatedAt  time.Time `json:"createdAt" bson:"createdAt" gorm:"column:createdAt"`
	UpdatedAt  time.Time `json:"updatesAt" bson:"updatedAt" gorm:"column:updatedAt"`
	UserID     string    `json:"userID" bson:"userID" gorm:"column:userID"`
	Name       string    `json:"name" bson:"name" form:"name" gorm:"column:name"`
	Phone      string    `json:"phone" bson:"phone" form:"phone" gorm:"column:phone"`
	PinCode    string    `json:"pincode" bson:"pincode" form:"pinCode" gorm:"column:pinCode"`
	HouseArea  string    `json:"houseArea" bson:"houseArea" form:"houseArea" gorm:"column:houseArea"`
	StreetName string    `json:"streetName" bson:"streetName" form:"area" gorm:"column:area"`
	City       string    `json:"city" bson:"city" form:"city" gorm:"column:city"`
	State      string    `json:"state" bson:"state" form:"state" gorm:"column:state"`
	IsDefault  bool      `json:"isDefault" bson:"isDefault" form:"isDefault" gorm:"column:isDefault"`
}

func (Address) GormDataType() string {
	return "json"
}

func (data Address) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *Address) Scan(value interface{}) error {

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

func (address Address) CreateIndexes() error {
	userIDModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "userID", Value: 1},
		},
	}

	indexName, err := db.AddressCollection.Indexes().CreateOne(context.Background(), userIDModel)
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

func (address Address) GetAddessString() (addressString string) {
	addressString = fmt.Sprintf("House: %s, City: %s, State: %s, Pincode: %s", address.HouseArea, address.City, address.State, address.PinCode)
	return
}

func (address Address) SaveToDB() (err error) {
	address.CreatedAt = time.Now()
	address.UpdatedAt = time.Now()
	_, err = db.AddressCollection.InsertOne(context.Background(), address)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func (address Address) UpdateInDB() (err error) {
	address.UpdatedAt = time.Now()
	_, err = db.AddressCollection.ReplaceOne(context.Background(), bson.M{"_id": address.ID}, address)
	if err != nil && err != mongo.ErrNoDocuments {
		fmt.Println(err)
		return
	}
	return
}

func GetAddress(addressId string) (Address, error) {
	// Find the product in the database
	var address Address
	err := db.AddressCollection.FindOne(context.Background(), bson.M{"_id": addressId}).Decode(&address)
	if err != nil {
		fmt.Println(err)
		return address, err
	}

	return address, nil
}

func DeleteAddress(addressId string, userId string) error {
	// Delete the product in the database
	delResult, err := db.AddressCollection.DeleteOne(context.Background(), bson.M{"_id": addressId, "userID": userId})
	if err != nil {
		fmt.Println(err)
		return err
	}
	if delResult.DeletedCount == 0 {
		return fmt.Errorf("couldn't find address with given ID")
	}
	return nil
}

func UpdateDefaultAddress(addressId string, userID string) error {

	_, err := db.AddressCollection.UpdateMany(context.Background(), bson.M{"isDefault": true, "userID": userID}, bson.D{{"$set", bson.D{{"isDefault", false}}}})
	if err != nil {
		return err
	}

	_, err = db.AddressCollection.UpdateOne(context.Background(), bson.M{"_id": addressId, "userID": userID}, bson.D{{"$set", bson.D{{"isDefault", true}}}})
	if err != nil {
		return err
	}

	return nil

}
