package models

import (
	"context"
	"errors"
	"fmt"
	"hermes/db"
	"hermes/utils/data"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type QuickReply struct {
	ID      string `json:"id" bson:"_id"`
	Message string `json:"message"`
	Name    string `json:"name"`
}

func (QuickReply QuickReply) CreateIndexes() error {
	indexModels := []mongo.IndexModel{}

	IDModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "_id", Value: 1},
		},
	}
	indexModels = append(indexModels, IDModel)

	nameModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "name", Value: 1},
		},
	}
	indexModels = append(indexModels, nameModel)

	indexName, err := db.QuickRepliesCollection.Indexes().CreateMany(context.Background(), indexModels)
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

func (QuickReply QuickReply) CreateQuickReply() (err error) {
	QuickReply.ID = data.GetUUIDString("reply")
	_, err = db.QuickRepliesCollection.InsertOne(context.Background(), QuickReply)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func (QuickReply QuickReply) EditQuickReply() (err error) {
	_, err = db.QuickRepliesCollection.ReplaceOne(context.Background(), bson.M{"_id": QuickReply.ID}, QuickReply)
	if err != nil {
		fmt.Println(err)
		return 
	}
	return
}

func (QuickReply QuickReply) DeleteQuickReply() (err error) {
	delResult, err := db.QuickRepliesCollection.DeleteOne(context.Background(), bson.M{"_id": QuickReply.ID})
	if err != nil {
		fmt.Println(err)
		return err
	}
	if delResult.DeletedCount == 0 {
		return errors.New("Unable to delete")
	}
	return
}

func GetAllQuickReplies() (quickReplies []QuickReply, err error){
	repliesCursor, err := db.QuickRepliesCollection.Find(context.Background(), bson.M{})
	if err != nil {
		fmt.Println(err)
		return
	}

	err = repliesCursor.All(context.Background(), &quickReplies)
	if err != nil {
		fmt.Println(err)
	}
	return
}

func GetQuickReplyByName(name string) (quickReply QuickReply, err error) {
	err = db.QuickRepliesCollection.FindOne(context.Background(), bson.M{"name": name}).Decode(&quickReply)
	return
}

func GetQuickReplyByID(ID string) (quickReply QuickReply, err error) {
	err = db.QuickRepliesCollection.FindOne(context.Background(), bson.M{"_id": ID}).Decode(&quickReply)
	return
}