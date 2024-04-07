package models

import (
	"context"
	"fmt"
	"hermes/db"
	"hermes/utils/data"
	utils "hermes/utils/queries"
	"html/template"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type ApplicationWithCampaign struct {
	InfluencerCampaignApplication `bson:",inline"`
	Campaign                      Campaign `json:"campaign" bson:"campaign"`
}

type Campaign struct {
	ID           string        `json:"_id" bson:"_id"`
	CreatedAt    time.Time     `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time     `json:"updatedAt" bson:"updatedAt"`
	Banner       string        `json:"banner" bson:"banner"`
	Title        string        `json:"title" bson:"title"`
	BrandID      string        `json:"brandID" bson:"brandID"`
	Brand        Brand         `json:"brand" bson:"brand"`
	SellerID     string        `json:"sellerID" bson:"sellerID"`
	Products     []string      `json:"products" bson:"products"`
	ProductArray []Product     `json:"productArray" bson:"productArray"`
	Platform     Platform      `json:"platform" bson:"platform"`
	Locations    []string      `json:"locations" bson:"locations"`
	Genders      GenderChoice  `json:"genders" bson:"genders"`
	Deliverables Deliverable   `json:"deliverables" bson:"deliverables"`
	MinAge       int           `json:"minAge" bson:"minAge"`
	MaxAge       int           `json:"maxAge" bson:"maxAge"`
	IsActive     bool          `json:"isActive" bson:"isActive"`
	Description  template.HTML `json:"description" bson:"description"`
	Commission   int           `json:"commission" bson:"commission"`
	Discount     int           `json:"discount" bson:"discount"`
}

type Deliverable struct {
	InstagramDeliverable InstagramDeliverable `json:"instagram" bson:"instagram"`
	YoutubeDeliverable   YoutubeDeliverable   `json:"youtube" bson:"youtube"`
}

type InstagramDeliverable struct {
	Hashtags []string `json:"hashtags" bson:"hashtags"`
	Accounts []string `json:"accounts" bson:"accounts"`
	Title    string   `json:"title" bson:"title"`
	Post     int      `json:"post" bson:"post"`
	Reel     int      `json:"reel" bson:"reel"`
	Story    int      `json:"story" bson:"story"`
}

type YoutubeDeliverable struct {
	Hashtags []string `json:"hashtags" bson:"hashtags"`
	Title    string   `json:"title" bson:"title"`
	Video    int      `json:"video" bson:"video"`
	Short    int      `json:"short" bson:"short"`
}

type Platform struct {
	Instagram bool `json:"instagram" bson:"instagram"`
	YouTube   bool `json:"youtube" bson:"youtube"`
	Snapchat  bool `json:"snapchat" bson:"snapchat"`
}

type GenderChoice struct {
	Male   bool `json:"male" bson:"male"`
	Female bool `json:"female" bson:"female"`
	Others bool `json:"others" bson:"others"`
}

func (campaign *Campaign) Create() (err error) {
	campaign.ID = data.GetUUIDString("campaign")
	campaign.CreatedAt = time.Now()
	campaign.UpdatedAt = time.Now()
	_, err = db.CampaignCollection.InsertOne(context.Background(), campaign)
	return
}

func (campaign *Campaign) Update() (err error) {
	campaign.UpdatedAt = time.Now()
	_, err = db.CampaignCollection.UpdateOne(context.Background(), bson.M{"_id": campaign.ID}, bson.M{"$set": campaign})
	return
}

func GetAllCampaigns() (campaigns []Campaign, err error) {
	pipeline := utils.GetAllCampaigns()
	cur, err := db.CampaignCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return
	}
	err = cur.All(context.Background(), &campaigns)
	return
}

func GetCampaignsBySeller(sellerID string) (campaigns []Campaign, err error) {
	cur, err := db.CampaignCollection.Find(context.Background(), bson.M{"sellerID": sellerID})
	if err != nil {
		return
	}
	err = cur.All(context.Background(), &campaigns)
	return
}

func GetCampaignsByBrand(brandID string) (campaigns []Campaign, err error) {
	cur, err := db.CampaignCollection.Find(context.Background(), bson.M{"brandID": brandID})
	if err != nil {
		return
	}
	err = cur.All(context.Background(), &campaigns)
	return
}

func GetCampaignByID(id string) (campaign Campaign, err error) {
	err = db.CampaignCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&campaign)
	if err != nil {
		return
	}
	return
}

func GetCampaignByIDWithProductsAndBrand(id string) (campaign Campaign, err error) {
	var response []Campaign
	pipeline := bson.A{bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: id}}}}}
	pipeline = append(pipeline, utils.GetAllCampaigns()...)

	cur, err := db.CampaignCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return Campaign{}, err
	}

	err = cur.All(context.Background(), &response)
	if err != nil {
		return Campaign{}, err
	}

	if len(response) == 0 {
		return Campaign{}, fmt.Errorf("no campaigns found")
	}

	return response[0], nil
}

func GetApplicationWithFullCampaign(campaignID string, influencerID string) (application ApplicationWithCampaign, err error) {

	var response []ApplicationWithCampaign

	pipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.D{{Key: "campaignID", Value: campaignID}}}},
	}
	pipeline = append(pipeline, utils.QueryGetAllAppliedCampaignsByInfluencer(influencerID)...)
	cur, err := db.InfluencerCampaignApplicationCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return ApplicationWithCampaign{}, err
	}

	err = cur.All(context.Background(), &response)
	if err != nil {
		return ApplicationWithCampaign{}, err
	}

	if len(response) == 0 {
		return ApplicationWithCampaign{}, fmt.Errorf("no applications found")
	}

	return response[0], nil
}

func (campaign *Campaign) CreateApplication(influencerID string) (InfluencerCampaignApplication, error) {

	var application InfluencerCampaignApplication
	if influencerID == "" {
		return application, fmt.Errorf("influencer ID is empty")
	}

	if len(campaign.Products) == 0 {
		return application, fmt.Errorf("no products in campaign")
	}

	application, err := CreateInfluencerCampaignApplication(campaign.ID, influencerID, campaign.Products[0])
	if err != nil {
		return application, err
	}

	return application, nil
}
