package models

import (
	"context"
	"fmt"
	"hermes/db"
	"hermes/utils/data"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type InfluencerCampaignApplication struct {
	ID           string     `json:"id" bson:"_id"`
	CreatedAt    time.Time  `json:"createdAt" bson:"createdeAt"`
	UpdatedAt    time.Time  `json:"updatedAt" bson:"updatedAt"`
	SubmittedAt  *time.Time `json:"submittedAt" bson:"submittedAt"`
	ApprovedAt   *time.Time `json:"approvedAt" bson:"approvedAt"`
	CamapignID   string     `json:"campaignID" bson:"campaignID"`
	InfluencerID string     `json:"influencerID" bson:"influencerID"`
	ProductID    string     `json:"productID" bson:"productID"`
	Approved     bool       `json:"approved" bson:"approved"`
	Address      Address    `json:"address" bson:"address"`
	Coupon       Coupon     `json:"coupon" bson:"coupon"`
}

func (w InfluencerCampaignApplication) CreateIndexes() error {

	indexModels := []mongo.IndexModel{}

	t := true
	options := options.Index()
	options.Unique = &t

	CodeModel := mongo.IndexModel{
		Keys: bson.M{
			"coupon.code": 1,
		},
		Options: options,
	}
	indexModels = append(indexModels, CodeModel)

	_, err := db.InfluencerCampaignApplicationCollection.Indexes().CreateMany(context.Background(), indexModels)
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

func CreateInfluencerCampaignApplication(campaignID string, influencerID string, productID string) (application InfluencerCampaignApplication, err error) {

	err = db.InfluencerCampaignApplicationCollection.FindOne(context.Background(), bson.M{"campaignID": campaignID, "influencerID": influencerID}).Decode(&application)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			application.ID = data.GetUUIDString("influ-camp-appl")
			application.CamapignID = campaignID
			application.InfluencerID = influencerID
			application.CreatedAt = time.Now()
			application.UpdatedAt = time.Now()
			application.ProductID = productID

			_, err = db.InfluencerCampaignApplicationCollection.InsertOne(context.Background(), application)
			return application, err
		} else {
			return application, err
		}
	}
	return
}

func (application *InfluencerCampaignApplication) UpdateProductSelection(productID string) error {

	application.ProductID = productID
	_, err := db.InfluencerCampaignApplicationCollection.UpdateOne(context.Background(), bson.M{"_id": application.ID}, bson.M{"$set": bson.M{
		"productID": application.ProductID,
	}})

	return err
}

func (application *InfluencerCampaignApplication) UpdateAddressSelection(address Address) error {

	application.Address = address
	_, err := db.InfluencerCampaignApplicationCollection.UpdateOne(context.Background(), bson.M{"_id": application.ID}, bson.M{"$set": bson.M{
		"address": application.Address,
	}})

	return err
}

func (application *InfluencerCampaignApplication) SubmitApplication(discount Discount) error {
	t := time.Now()
	application.SubmittedAt = &t
	application.Coupon.Discount = discount
	_, err := db.InfluencerCampaignApplicationCollection.UpdateOne(context.Background(), bson.M{"_id": application.ID}, bson.M{"$set": bson.M{
		"submittedAt": application.SubmittedAt,
		"coupon":      application.Coupon,
	}})

	return err
}

func (application InfluencerCampaignApplication) Approve() (err error) {
	application.Approved = true
	t := time.Now()
	application.ApprovedAt = &t
	coupon, err := createCoupon(application)
	if err != nil {
		return
	}
	application.Coupon = coupon
	db.InfluencerCampaignApplicationCollection.ReplaceOne(context.Background(), bson.M{"_id": application.ID}, application)
	return
}

func createCoupon(application InfluencerCampaignApplication) (coupon Coupon, err error) {
	influencer, err := GetInfluencerByID(application.InfluencerID)
	if err != nil {
		return
	}
	name := influencer.Instagram.Name
	couponCode := ""
	if len(name) < 5 {
		couponCode = name
	} else {
		couponCode = name[0:5]
	}
	for {
		uniqueNumber := 1
		couponCode += fmt.Sprint(uniqueNumber)
		coupon, _ := GetCouponByCode(couponCode)
		if coupon.ID != "" {
			break
		} else {
			uniqueNumber++
		}
	}
	coupon, err = CreateCoupon(30, application.Coupon.Discount, couponCode, ProductDiscount, []string{application.ProductID}, SingleUsePerUser)
	return
}

func GetInfluencerCampaignApplicaton(influencerID string, campaignID string) (application InfluencerCampaignApplication, err error) {
	err = db.InfluencerCampaignApplicationCollection.FindOne(context.Background(), bson.M{"campaignID": campaignID, "influencerID": influencerID}).Decode(&application)
	return
}

func GetInfluencerApplicationByID(applicationID string) (application InfluencerCampaignApplication, err error) {
	err = db.InfluencerCampaignApplicationCollection.FindOne(context.Background(), bson.M{"_id": applicationID}).Decode(&application)
	return
}

func GetInfluencercampaignApplicatonByCode(code string) (application InfluencerCampaignApplication, err error) {
	err = db.InfluencerCampaignApplicationCollection.FindOne(context.Background(), bson.M{"coupon.code": code}).Decode(&application)
	return
}
