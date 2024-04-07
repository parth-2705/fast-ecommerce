package models

import (
	"context"
	"fmt"
	"hermes/db"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Team struct {
	ID        string    `json:"id" bson:"_id"`
	Name      string    `json:"name" bson:"name"`
	DealID    string    `json:"dealID" bson:"dealID"`
	Capacity  int       `json:"capacity" bson:"capacity"`
	Strength  int       `json:"strength" bson:"strength"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

// Create
func CreateTeam(team Team) error {
	team.CreatedAt = time.Now()
	team.UpdatedAt = time.Now()
	_, err := db.TeamCollection.InsertOne(context.Background(), team)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return err
	}

	return nil
}

// Read
func GetTeamByID(teamID string) (Team, error) {
	var team Team
	err := db.TeamCollection.FindOne(context.Background(), bson.M{"_id": teamID}).Decode(&team)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return team, err
	}

	return team, nil
}

// Update
func UpdateTeam(team Team) error {
	team.UpdatedAt = time.Now()
	result, err := db.TeamCollection.ReplaceOne(context.Background(), bson.M{"_id": team.ID}, team)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return err
	}

	if result.UpsertedCount == 0 {
		return fmt.Errorf("no documents updated")
	}

	return nil
}

// Delete
func DeleteTeam(teamID string) error {
	_, err := db.TeamCollection.DeleteOne(context.Background(), bson.M{"_id": teamID})
	if err != nil {
		fmt.Printf("Error: %v", err)
		return err
	}

	return nil
}

// List
func GetAllTeamsByDealID(dealID string, limit int32) ([]Team, error) {
	var teams []Team
	var curr *mongo.Cursor
	var err error

	filters := bson.D{{Key: "$match", Value: bson.D{{Key: "dealID", Value: dealID}}}}
	aggregrateSearchObject := bson.A{}
	aggregrateSearchObject = append(aggregrateSearchObject, filters)

	aggregrateSearchObject = append(aggregrateSearchObject,
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "teamMembers"},
					{Key: "localField", Value: "_id"},
					{Key: "foreignField", Value: "teamID"},
					{Key: "as", Value: "members"},
				},
			},
		},
	// bson.D{
	// 	{Key: "$lookup",
	// 		Value: bson.D{
	// 			{Key: "from", Value: "profiles"},
	// 			{Key: "localField", Value: "userID"},
	// 			{Key: "foreignField", Value: "userID"},
	// 			{Key: "as", Value: "user"},
	// 		},
	// 	},
	// },
	// bson.D{
	// 	{Key: "$unwind",
	// 		Value: bson.D{
	// 			{Key: "path", Value: "$user"},
	// 			{Key: "preserveNullAndEmptyArrays", Value: true},
	// 		},
	// 	},
	// },
	)

	fmt.Println("aggregrateSearchObject:", aggregrateSearchObject)

	if limit <= 0 {
		fmt.Println("limit <=0 :", limit)
		curr, err = db.TeamCollection.Aggregate(context.Background(), aggregrateSearchObject)
	} else {
		opts := options.Aggregate().SetBatchSize(limit)
		curr, err = db.TeamCollection.Aggregate(context.Background(), aggregrateSearchObject, opts)
	}
	if err != nil {
		fmt.Printf("Error: %v", err)
		return teams, err
	}

	if err = curr.All(context.TODO(), &teams); err != nil {
		return teams, err
	}
	defer curr.Close(context.Background())

	return teams, nil
}
