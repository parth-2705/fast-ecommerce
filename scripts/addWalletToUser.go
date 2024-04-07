package scripts

import (
	"context"
	"fmt"
	"hermes/db"
	"hermes/models"

	"go.mongodb.org/mongo-driver/bson"
)

func addWalletToUser() error {

	var profiles []models.Profile

	cursor, _ := db.ProfileCollection.Find(context.TODO(), bson.D{})
	err := cursor.All(context.Background(), &profiles)
	if err != nil {
		return err
	}

	fmt.Printf("len(profiles): %v\n", len(profiles))

	for idx, profile := range profiles {
		fmt.Println("idx: ", idx)

		if profile.WalletID != "" {
			continue
		}

		// Create a new wallet and assign it to the profile
		err = profile.CreateWallet()
		if err != nil {
			panic(err)
		}

		err = profile.Update()
		if err != nil {
			panic(err)
		}
	}

	return nil
}
