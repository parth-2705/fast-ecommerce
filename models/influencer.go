package models

import (
	"context"
	"fmt"
	"hermes/db"
	"hermes/utils/data"
	utils "hermes/utils/queries"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Gender string

const (
	Other Gender = "other"
	Men   Gender = "men"
	Women Gender = "women"
)

type Influencer struct {
	ID           string    `json:"id" bson:"_id"`
	CreatedAt    time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt" bson:"updatedAt"`
	Phone        string    `json:"phone" bson:"phone"`
	UserID       string    `json:"userID" bson:"userID"`
	Name         string    `json:"name" bson:"name"`
	Email        string    `json:"email" bson:"email"`
	ProfileImage Image     `json:"profileImage" bson:"profileImage"`
	Gender       Gender    `json:"gender" bson:"gender"`
	DOB          time.Time `json:"dob" bson:"dob"`
	Address      Address   `json:"address" bson:"address"`
	Instagram    Instagram `json:"instagram" bson:"instagram"`
	Snapchat     Snapchat  `json:"snapchat" bson:"snapchat"`
	YouTube      YouTube   `json:"youtube" bson:"youtube"`
}

func (influencer *Influencer) Create(userID string) error {

	if len(influencer.Phone) == 0 {
		return fmt.Errorf("phone number is empty")
	}

	if influencer.InfluencerExistsByPhone() {
		return nil
	}

	if len(userID) == 0 {
		return fmt.Errorf("user ID is empty")
	}

	influencer.UserID = userID

	influencer.ID = data.GetUUIDString("influencer")
	influencer.CreatedAt = time.Now()
	influencer.UpdatedAt = time.Now()

	opts := options.Update().SetUpsert(true)
	_, err := db.InfluencerCollection.UpdateOne(context.Background(), bson.M{"phone": influencer.Phone}, bson.M{"$set": influencer}, opts)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (influencer *Influencer) Update() error {
	influencer.UpdatedAt = time.Now()

	_, err := db.InfluencerCollection.UpdateOne(context.Background(), bson.M{"_id": influencer.ID}, bson.M{"$set": influencer})
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func GetInfluencerByID(influencerID string) (Influencer, error) {
	var influencer Influencer
	err := db.InfluencerCollection.FindOne(context.Background(), bson.M{"_id": influencerID}).Decode(&influencer)
	if err != nil {
		fmt.Println(err)
		return influencer, err
	}
	return influencer, nil
}

func (influencer *Influencer) InfluencerExistsByPhone() bool {
	err := db.InfluencerCollection.FindOne(context.Background(), bson.M{"phone": influencer.Phone}).Decode(influencer)
	if err != nil {
		return err != mongo.ErrNoDocuments
	}

	if len(influencer.ID) == 0 {
		return false
	}

	return true
}

func (influencer *Influencer) IsConnected() bool {

	if influencer.Instagram.IsConnected && influencer.YouTube.IsConnected {
		return true
	}

	return false
}

func (influencer *Influencer) HasMinimumConnections() bool {

	if influencer.Instagram.IsConnected || influencer.Instagram.Approved {
		return true
	}

	if influencer.YouTube.IsConnected || influencer.YouTube.Approved {
		return true
	}

	return false
}

func (influencer *Influencer) GetMyCampaigns() ([]Campaign, error) {

	type MyCampaigns struct {
		Campaign Campaign `json:"campaign" bson:"campaign"`
	}

	var response []MyCampaigns

	cur, err := db.InfluencerCampaignApplicationCollection.Aggregate(context.Background(), utils.QueryGetAllAppliedCampaignsByInfluencer(influencer.ID))
	if err != nil {
		return []Campaign{}, err
	}

	err = cur.All(context.Background(), &response)
	if err != nil {
		return []Campaign{}, err
	}

	var campaings []Campaign

	for _, app := range response {
		campaings = append(campaings, app.Campaign)
	}

	return campaings, nil
}

func (influencer *Influencer) HasApplied(campaignID string) (bool, error) {

	var application InfluencerCampaignApplication
	err := db.InfluencerCampaignApplicationCollection.FindOne(context.Background(), bson.M{"campaignID": campaignID, "influencerID": influencer.ID}).Decode(&application)
	if err != nil {

		if err == mongo.ErrNoDocuments {
			return false, nil
		}

		return false, err
	}

	if application.SubmittedAt != nil {
		return true, nil
	}

	return false, nil
}
