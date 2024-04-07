package models

import (
	"context"
	"hermes/db"
	"hermes/utils/data"

	"go.mongodb.org/mongo-driver/bson"
)

type Admin struct {
	ID    string `json:"id" bson:"_id"`
	Email string `json:"email" bson:"email"`
}

func CreateAdmin(email string) (admin Admin, err error) {
	admin.Email = email
	admin.ID = data.GetUUIDString("admin")
	_, err = db.AdminsCollection.InsertOne(context.Background(), &admin)
	return admin, err
}

func GetAdmin(email string) (admin Admin, err error) {
	err = db.AdminsCollection.FindOne(context.Background(), bson.M{"email": email}).Decode(&admin)
	return admin, err
}

func GetAdminbyID(ID string) (admin Admin, err error) {
	err = db.AdminsCollection.FindOne(context.Background(), bson.M{"_id": ID}).Decode(&admin)
	return admin, err
}
