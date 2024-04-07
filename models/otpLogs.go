package models

import (
	"context"
	"fmt"
	"hermes/db"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type OTPLog struct {
	ID           string    `json:"id" bson:"_id" gorm:"column:id"`
	MobileNumber string    `json:"mobileNumber" bson:"mobileNumber" gorm:"column:mobileNumber"`
	SendTime     time.Time `json:"sendTime" bson:"sendTime" gorm:"column:sendTime"`
	Success      bool      `json:"success" bson:"success" gorm:"column:success"`
	UserAgentID  string    `json:"userAgentID" bson:"userAgentID" gorm:"column:userAgentID"`
	Referrer     string    `json:"referrer" bson:"referrer" gorm:"column:referrer"`
	DirtyInput   string    `json:"dirtyInput" bson:"dirtyInput" gorm:"column:dirtyInput"`
	Product      Product   `json:"product" bson:"product" gorm:"column:product"`
}

func (OTPLog) TableName() string {
	return "OTPLogs"
}

func (OTPLog OTPLog) CreateIndexes() error {
	// Create an index on sendTime field
	sendTimeIndex := mongo.IndexModel{
		Keys: bson.M{
			"sendTime": -1, // 1 for ascending order, -1 for descending order
		},
	}

	// Create an index on MobileNumber field
	phoneIndex := mongo.IndexModel{
		Keys: bson.M{
			"mobileNumber": 1, // 1 for ascending order, -1 for descending order
		},
	}

	// Create both indexes using CreateMany
	_, err := db.OTPLogCollection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{sendTimeIndex, phoneIndex})
	if err != nil {
		// Handle error
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

type AdminLogs struct {
	OTPLog   `bson:",inline"`
	Internal bool `json:"internal" bson:"internal"`
}

func GetOTPLogsListWithUser() ([]AdminLogs, error) {

	var logs []AdminLogs

	pipeline := bson.A{
		bson.D{{Key: "$sort", Value: bson.D{{Key: "sendTime", Value: -1}}}},
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "users"},
				{Key: "localField", Value: "mobileNumber"},
				{Key: "foreignField", Value: "phone"},
				{Key: "as", Value: "user"},
			}},
		},
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$user"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		bson.D{{Key: "$addFields", Value: bson.D{{Key: "internal", Value: "$user.internal"}}}},
	}

	cur, err := db.OTPLogCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return logs, err
	}
	err = cur.All(context.Background(), &logs)
	if err != nil {
		return logs, err
	}

	return logs, nil
}
