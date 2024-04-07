package models

import (
	"context"
	"fmt"
	"hermes/db"
	"hermes/utils/data"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type OAuth struct {
	ID           string             `json:"id" bson:"_id"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updatedAt"`
	Phone        string             `json:"phone" bson:"phone"`
	InfluencerID string             `json:"influencerID" bson:"influencerID"`
	Info         *map[string]string `json:"info" bson:"info"`
	OauthSuccess bool               `json:"oauthSuccess" bson:"oauthSuccess"`
	Redirect     string             `json:"redirect" bson:"redirect"`
}

func (oAuth *OAuth) Create() error {
	oAuth.ID = data.GetUUIDString("oAuth")
	oAuth.CreatedAt = time.Now()
	oAuth.UpdatedAt = time.Now()

	_, err := db.OAuthStateCollection.InsertOne(context.Background(), oAuth)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (oAuth *OAuth) MarkOAuthAsCompleted() error {
	_, err := db.OAuthStateCollection.UpdateOne(context.Background(), bson.M{"_id": oAuth.ID}, bson.M{"$set": bson.M{"oauthSuccess": true}})
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func GetOauthByID(oauthID string) (OAuth, error) {
	var oauth OAuth
	err := db.OAuthStateCollection.FindOne(context.Background(), bson.M{"_id": oauthID}).Decode(&oauth)
	if err != nil {
		fmt.Println(err)
		return oauth, err
	}
	return oauth, nil
}
