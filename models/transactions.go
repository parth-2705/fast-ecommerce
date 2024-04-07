package models

import (
	"context"
	"hermes/db"
	"hermes/utils/data"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type Transaction struct {
	ID                  string    `json:"_id" bson:"_id"`
	CreatedAt           time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt" bson:"updatedAt"`
	UTR                 string    `json:"utr" bson:"utr"`
	SellerID            string    `json:"sellerID" bson:"sellerID"`
	ShipmentsRemitted   []string  `json:"shipmentsRemitted" bson:"shipmentsRemitted"`
	TransactionAmount   float64   `json:"transactionAmount" bson:"transactionAmount"`
	TransactionOverview Charges   `json:"transactionOverview" bson:"transactionOverview"`
	RemitReflected      bool      `json:"remitReflected" bson:"remitReflected"`
	StartTime           time.Time `json:"startTime" bson:"startTime"`
	EndTime             time.Time `json:"endTime" bson:"endTime"`
}
type Charges struct {
	TotalTax                float64 `json:"total_tax" bson:"total_tax"`
	TotalTransactionCharges float64 `json:"total_transaction_charges" bson:"total_transaction_charges"`
	TotalCODCharges         float64 `json:"total_cod_charges" bson:"total_cod_charges"`
	TotalForwardAmount      float64 `json:"total_fwd_amount" bson:"total_fwd_amount"`
	TotalRTOAmount          float64 `json:"total_rto_amount" bson:"total_rto_amount"`
	TotalOrderAmount        float64 `json:"total_order_amount" bson:"total_order_amount"`
	TotalPayableAmount      float64 `json:"total_payable_amount" bson:"total_payable_amount"`
	TotalCommission         float64 `json:"total_commission" bson:"total_commission"`
}

func (transaction *Transaction) Create() (err error) {
	transaction.ID = data.GetUUIDString("transaction")
	transaction.CreatedAt = time.Now()
	transaction.UpdatedAt = time.Now()
	_, err = db.TransactionCollection.InsertOne(context.Background(), transaction)
	return
}

func (transaction *Transaction) Update() (err error) {
	transaction.UpdatedAt = time.Now()
	_, err = db.TransactionCollection.UpdateOne(context.Background(), bson.M{"_id": transaction.ID}, transaction)
	return
}

func GetAllTransactions() (transactions []Transaction, err error) {
	cur, err := db.TransactionCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return
	}
	err = cur.All(context.Background(), &transactions)
	return
}

func GetTransactionByID(id string) (transactions []Transaction, err error) {
	cur, err := db.TransactionCollection.Find(context.Background(), bson.M{"_id": id})
	if err != nil {
		return
	}
	err = cur.All(context.Background(), &transactions)
	return
}

func GetTransactionsBySeller(sellerID string) (transactions []Transaction, err error) {
	cur, err := db.TransactionCollection.Find(context.Background(), bson.M{"sellerID": sellerID})
	if err != nil {
		return
	}
	err = cur.All(context.Background(), &transactions)
	return
}
