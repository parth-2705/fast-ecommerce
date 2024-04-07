package models

import (
	"context"
	"fmt"
	"hermes/db"

	"go.mongodb.org/mongo-driver/bson"
)

type Notify struct {
	UserID    string `json:"userID" bson:"_userID"`
	ProductID string `json:"productID" bson:"_productID"`
}

func GetIfItemIsInNotifylistOfUser(userID string, productID string) bool {
	var notify Notify
	err := db.NotifyCollection.FindOne(context.Background(), bson.M{"_userID": userID, "_productID": productID}).Decode(&notify)
	if err != nil {
		fmt.Println("error", err)
		return false
	}
	return true
}

func AddProductToNotify(notify Notify) error {
	_, err := db.NotifyCollection.InsertOne(context.Background(), notify)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func RemoveProductFromNotify(notify Notify) error {
	_, err := db.NotifyCollection.DeleteOne(context.Background(), notify)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
