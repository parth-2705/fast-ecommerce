package models

import (
	"context"
	"fmt"
	"hermes/db"
	"hermes/utils/data"
	"math/rand"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Profile struct {
	ID        string    `json:"id" bson:"_id"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
	// userId will have and unique index
	UserID                string `json:"userID" bson:"userID"` // unique
	StripeCustomerID      string `json:"stripeCustomerID" bson:"stripeCustomerID"`
	DefaultPaymentMethod  string `json:"defaultPaymentMethod" bson:"defaultPaymentMethod"`
	LastUsedPaymentMethod string `json:"lastUsedPaymentMethod" bson:"lastUsedPaymentMethod"`
	WishlistID            string `json:"wishlistID" bson:"wishlistID"`
	WalletID              string `json:"walletID" bson:"walletID"`
	WhatsappEnabled       bool   `json:"whatsappEnabled" bson:"whatsappEnabled"`
	StreamUserToken       string `json:"streamUserToken" bson:"streamUserToken"`
	ReferralCode          string `json:"referralCode" bson:"referralCode"`
}

func (profile Profile) CreateIndexes() error {

	indexModels := []mongo.IndexModel{}

	// t1 := true
	// name1 := "referralCode-1"
	// options1 := options.Index()
	// options1.Unique = &t1
	// options1.Name = &name1
	referralCodeModel := mongo.IndexModel{
		Keys: bson.M{
			"referralCode": 1,
		},
		// Options: options1,
	}
	indexModels = append(indexModels, referralCodeModel)

	userIDModel := mongo.IndexModel{
		Keys: bson.M{
			"userID": 1,
		},
	}
	indexModels = append(indexModels, userIDModel)

	_, err := db.ProfileCollection.Indexes().CreateMany(context.Background(), indexModels)
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

// Generate a unique referral code for this user
func GenerateReferralCode() string {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	code := make([]byte, 6)
	rand.Seed(time.Now().UnixNano())
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}

// Function to check if a generated referral code is unique
func isReferralCodeUnique(code string) bool {
	var profile Profile
	err := db.ProfileCollection.FindOne(context.Background(), bson.M{"referralCode": code}).Decode(&profile)
	if err != nil {
		return err == mongo.ErrNoDocuments
	}
	return false
}

func AssignReferralCode(userID string, code string) error {
	_, err := db.ProfileCollection.UpdateOne(context.Background(), bson.M{"userID": userID}, bson.M{"$set": bson.M{"referralCode": code}})
	if err != nil {
		return err
	}
	return nil
}

func (user User) CreateProfile() (Profile, error) {

	profile := Profile{
		ID:                   data.GetUUIDString("profile"),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
		UserID:               user.ID,
		DefaultPaymentMethod: "COD",
		WishlistID:           data.GetUUIDString("wishlist"),
		WhatsappEnabled:      true,
	}

	err := profile.CreateWallet()
	if err != nil {
		return profile, err
	}

	_, err = db.ProfileCollection.InsertOne(context.Background(), profile)
	if err != nil {
		fmt.Println(err)
		return profile, err
	}

	return profile, nil
}

func (profile Profile) Create() (err error) {
	if profile.DefaultPaymentMethod == "" {
		profile.DefaultPaymentMethod = "COD"
	}
	profile.ID = data.GetUUIDString("profile")
	profile.WishlistID = data.GetUUIDString("wishlist")
	profile.CreatedAt = time.Now()

	// Create a new wallet and assign it to the profile
	err = profile.CreateWallet()
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = db.ProfileCollection.InsertOne(context.Background(), profile)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func (profile Profile) CreateIfNotExists() (err error) {

	if len(profile.UserID) == 0 {
		return fmt.Errorf("user id empty")
	}

	if profile.DefaultPaymentMethod == "" {
		profile.DefaultPaymentMethod = "COD"
	}
	profile.ID = data.GetUUIDString("profile")
	profile.WishlistID = data.GetUUIDString("wishlist")
	profile.CreatedAt = time.Now()
	profile.UpdatedAt = time.Now()
	profile.WhatsappEnabled = true

	var newProfile Profile
	err = db.ProfileCollection.FindOne(context.Background(), bson.M{"userID": profile.UserID}).Decode(&newProfile)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			// Create a new wallet and assign it to the profile
			err = profile.CreateWallet()
			if err != nil {
				fmt.Println(err)
				return
			}

			_, err = db.ProfileCollection.InsertOne(context.Background(), profile)
			if err != nil {
				return
			}
		} else {
			fmt.Println(err)
			return
		}
	}
	return
}

func (profile Profile) Update() (err error) {
	profile.UpdatedAt = time.Now()
	_, err = db.ProfileCollection.UpdateOne(context.Background(), bson.M{"userID": profile.UserID}, bson.M{"$set": profile})
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func (profile Profile) MyReferredProfiles() (int64, error) {

	referredProfiles, err := db.UserCollection.CountDocuments(context.Background(), bson.M{"referredByUser": profile.UserID})
	if err != nil {
		fmt.Println(err)
		return referredProfiles, err
	}
	return referredProfiles, err
}
