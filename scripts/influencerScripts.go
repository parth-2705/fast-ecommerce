package scripts

import (
	"context"
	"fmt"
	"hermes/db"
	"hermes/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func addUserIDToInfluencerModel() (err error) {
	var influencers []models.Influencer
	curr, err := db.InfluencerCollection.Find(context.Background(), bson.M{})

	if err != nil {
		fmt.Println(err)
		return
	}

	err = curr.All(context.Background(), &influencers)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, influencer := range influencers {

		if len(influencer.UserID) == 0 {
			user, err := models.GetUser(influencer.Phone)
			if err != nil {

				if err == mongo.ErrNoDocuments {
					fmt.Println(err.Error())
					err = nil
				} else {
					panic(err)
				}
			}

			influencer.UserID = user.ID
			err = influencer.Update()
			if err != nil {
				panic(err)
			}
		}
	}
	return
}
