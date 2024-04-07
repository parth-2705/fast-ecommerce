package models

import (
	"context"
	"hermes/db"
	"hermes/utils/data"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type Commission struct {
	ID         string    `json:"_id" bson:"_id"`
	CreatedAt  time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt" bson:"updatedAt"`
	BrandID    string    `json:"brandID" bson:"brandID"`
	Commission float64   `json:"commission" bson:"commission"`
	From       time.Time `json:"from" bson:"from"`
}

type CommissionHistory struct {
	ID      string                  `json:"_id" bson:"_id"`
	History []CommissionHistoryItem `json:"history" bson:"history"`
}

type CommissionHistoryItem struct {
	UpdatedAt  time.Time `json:"updatedAt" bson:"updatedAt"`
	Commission float64   `json:"commission" bson:"commission"`
	From       time.Time `json:"from" bson:"from"`
}

func (commission Commission) Create() (err error) {
	commission.ID = data.GetUUIDString("comm")
	_, err = db.CommissionCollection.InsertOne(context.Background(), commission)
	if err != nil {
		return
	}
	var tempCommHistory CommissionHistory
	tempCommHistory.ID = commission.ID
	tempCommHistory.History = []CommissionHistoryItem{{
		UpdatedAt:  commission.UpdatedAt,
		From:       commission.From,
		Commission: commission.Commission,
	}}
	_, err = db.CommissionHistoryCollection.ReplaceOne(context.Background(), bson.M{"_id": commission.ID}, commission)
	return
}

func (commission Commission) Update() (err error) {
	var tempCommHistory CommissionHistory
	err = db.CommissionHistoryCollection.FindOne(context.Background(), bson.M{"_id": commission.ID}).Decode(&tempCommHistory)
	if err != nil {
		return
	}
	tempCommHistory.History = append(tempCommHistory.History, CommissionHistoryItem{
		UpdatedAt:  commission.UpdatedAt,
		From:       commission.From,
		Commission: commission.Commission,
	})
	_, err = db.CommissionCollection.ReplaceOne(context.Background(), bson.M{"_id": commission.ID}, commission)
	if err != nil {
		return
	}
	_, err = db.CommissionHistoryCollection.ReplaceOne(context.Background(), bson.M{"_id": commission.ID}, tempCommHistory)
	return
}
