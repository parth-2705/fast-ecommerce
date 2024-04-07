package models

import (
	"context"
	"fmt"
	"hermes/db"
	"hermes/utils/data"
	"time"
)

type Contact struct {
	ID        string    `bson:"_id" json:"id"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
	FirstName string    `form:"first_name" binding:"required" bson:"first_name" json:"first_name"`
	LastName  string    `form:"last_name" binding:"required" bson:"last_name" json:"last_name"`
	Email     string    `form:"email" binding:"required" bson:"email" json:"email"`
	Mobile    string    `form:"mobile" binding:"required" bson:"mobile" json:"mobile"`
	Message   string    `form:"message" binding:"required" bson:"message" json:"message"`
}

func (contact Contact) SaveToDB() (err error) {
	contact.ID = data.GetUUIDString("contact")
	contact.CreatedAt = time.Now()
	contact.UpdatedAt = time.Now()
	_, err = db.ContactCollection.InsertOne(context.Background(), contact)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}
