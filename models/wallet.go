package models

import (
	"context"
	"fmt"
	"hermes/configs/Mysql"
	"hermes/db"
	"hermes/utils/data"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Wallet struct {
	ID        string    `json:"id" bson:"_id"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
	UserID    string    `json:"userID" bson:"userID"` // unique
	Balance   float64   `json:"balance" bson:"balance"`
	Blocked   bool      `json:"blocked"`
}

type walletTransactionType int

const (
	CREDIT walletTransactionType = iota
	DEBIT
)

type WalletTransaction struct {
	ID              string                `json:"id" bson:"_id"`
	CreatedAt       time.Time             `json:"createdAt" bson:"createdAt"`
	WalletID        string                `json:"walletID" bason:"_id" gorm:"index;size:200"` 
	TransactionType walletTransactionType `json:"transactionType" bson:"transactionType"`
	Amount          float64               `json:"transactionAmount" bson:"transactionAmount"`
	CurrentBalance  float64               `json:"balance" bson:"balance"`
}

func (w Wallet) CreateTransactionLog(transactionType walletTransactionType, amount float64) (WalletTransaction, error) {

	log := WalletTransaction{
		ID:              data.GetUUIDString("WT"),
		WalletID:        w.ID,
		TransactionType: transactionType,
		Amount:          amount,
		CurrentBalance:  w.Balance,
	}

	err := Mysql.DB.Model(&log).Create(&log).Error
	if err != nil {
		fmt.Printf("Wallet Transaction err: %v\n", err)
	}

	return log, err
}

func (w Wallet) CreateIndexes() error {

	indexModels := []mongo.IndexModel{}

	t := true
	options := options.Index()
	options.Unique = &t

	userIDModel := mongo.IndexModel{
		Keys: bson.M{
			"userID": 1,
		},
		Options: options,
	}
	indexModels = append(indexModels, userIDModel)

	_, err := db.WalletCollection.Indexes().CreateMany(context.Background(), indexModels)
	if err != nil {
		// handle error
		if strings.Contains(err.Error(), "Index with name") && strings.Contains(err.Error(), "already exists") {
			fmt.Println("Indexes already exist")
		} else {
			// Handle other errors
			fmt.Println("Error creating indexes:", err)
			return err
		}
	}
	return nil
}

// Function to add balance to the wallet
func (w *Wallet) AddBalance(amount float64) error {

	// Create a session for mongod transaction
	session, err := db.MongoClient.StartSession()
	if err != nil {
		return err
	}

	defer session.EndSession(context.Background())

	// Starting transaction
	err = session.StartTransaction()
	if err != nil {
		return err
	}

	w.Balance += amount
	w.UpdatedAt = time.Now()

	// Update the wallet in the MongoDB collection
	err = w.Update()
	if err != nil {
		log.Printf("failed to update wallet in the collection: %v", err)
		return err
	}

	// Create TransactionLog
	w.CreateTransactionLog(CREDIT, amount)

	// Commit the transaction
	err = session.CommitTransaction(context.Background())
	if err != nil {
		session.AbortTransaction(context.Background())
		return err
	}

	return nil
}

// Function to deduct balance from wallet
func (w *Wallet) DeductAmount(amount float64) error {

	if w.Balance < amount {
		return fmt.Errorf("not enough balance in Wallet")
	}

	// Create a session for mongod transaction
	session, err := db.MongoClient.StartSession()
	if err != nil {
		return err
	}

	defer session.EndSession(context.Background())

	// Starting transaction
	err = session.StartTransaction()
	if err != nil {
		return err
	}

	w.Balance -= amount
	w.UpdatedAt = time.Now()

	// Update the wallet in the MongoDB collection
	err = w.Update()
	if err != nil {
		log.Printf("failed to update wallet in the collection: %v", err)
		return err
	}

	// Create TransactionLog
	w.CreateTransactionLog(DEBIT, amount)

	// Commit the transaction
	err = session.CommitTransaction(context.Background())
	if err != nil {
		session.AbortTransaction(context.Background())
		return err
	}

	return nil
}

func (profile *Profile) CreateWallet() error {
	if len(profile.UserID) == 0 {
		return fmt.Errorf("userID empty")
	}

	var w Wallet
	w.ID = data.GetUUIDString("wallet")
	w.CreatedAt = time.Now()
	w.UpdatedAt = time.Now()
	w.UserID = profile.UserID
	w.Balance = 0
	_, err := db.WalletCollection.InsertOne(context.Background(), w)
	if err != nil {
		fmt.Println(err)
		return err
	}

	profile.WalletID = w.ID

	return nil
}

func (w *Wallet) Update() error {
	w.UpdatedAt = time.Now()
	_, err := db.WalletCollection.UpdateOne(context.Background(), bson.M{"_id": w.ID}, bson.M{"$set": bson.M{
		"balance":   w.Balance,
		"updatedAt": w.UpdatedAt,
	}})
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func GetUserWallet(userID string) (Wallet, error) {
	var wallet Wallet
	err := db.WalletCollection.FindOne(context.Background(), bson.M{"userID": userID}).Decode(&wallet)
	if err != nil {
		return wallet, err
	}
	return wallet, nil
}
