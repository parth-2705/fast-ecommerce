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

type TeamMember struct {
	ID        string    `json:"id" bson:"_id"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
	TeamID    string    `json:"teamID" bson:"teamID"`
	IsAdmin   bool      `json:"isAdmin" bson:"isAdmin"`
	UserID    string    `json:"userID" bson:"userID"`
	User      User      `json:"user" bson:"user"`
}

// Create
func CreateTeamMember(member TeamMember) error {
	member.CreatedAt = time.Now()
	member.UpdatedAt = time.Now()
	_, err := db.TeamMemberCollection.InsertOne(context.Background(), member)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return err
	}

	return nil
}

// Read
func GetTeamMembersByTeamID(teamID string, limit int32) ([]TeamMember, error) {
	var members []TeamMember
	var curr *mongo.Cursor
	var err error

	filters := bson.M{"$match": bson.M{"teamID": teamID}}
	aggregrateSearchObject := bson.A{}
	aggregrateSearchObject = append(aggregrateSearchObject, filters)

	aggregrateSearchObject = append(aggregrateSearchObject,
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "users"},
					{Key: "localField", Value: "userID"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "user"},
				},
			},
		},
		bson.D{
			{Key: "$unwind",
				Value: bson.D{
					{Key: "path", Value: "$user"},
					{Key: "preserveNullAndEmptyArrays", Value: true},
				},
			},
		},
	)

	if limit <= 0 {
		curr, err = db.TeamMemberCollection.Aggregate(context.Background(), aggregrateSearchObject)
	} else {
		opts := options.Aggregate().SetBatchSize(limit)
		curr, err = db.TeamMemberCollection.Aggregate(context.Background(), aggregrateSearchObject, opts)
	}

	if err != nil {
		fmt.Printf("Error: %v", err)
		return members, err
	}

	if err = curr.All(context.TODO(), &members); err != nil {
		return members, err
	}
	defer curr.Close(context.Background())
	return members, nil
}

// Update
func UpdateTeamMember(member TeamMember) error {
	member.UpdatedAt = time.Now()
	_, err := db.TeamMemberCollection.ReplaceOne(context.Background(), bson.M{"_id": member.ID}, member)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return err
	}

	return nil
}

// Delete
func DeleteTeamMember(teamMemberID string) error {
	_, err := db.TeamMemberCollection.DeleteOne(context.Background(), bson.M{"_id": teamMemberID})
	if err != nil {
		fmt.Printf("Error: %v", err)
		return err
	}

	return nil
}

// Check if current user is a member of this team or not
func IsUserTeamMember(user User, teamMembers []TeamMember) bool {
	if len(user.ID) == 0 {
		return false
	}

	for _, member := range teamMembers {
		if member.User.ID == user.ID {
			return true
		}
	}

	return false
}
