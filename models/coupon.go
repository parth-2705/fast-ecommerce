package models

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"hermes/configs/Mysql"
	"hermes/db"
	"hermes/services/Temporal/TemporalJobs"
	"hermes/utils"
	"hermes/utils/data"
	"math"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

type Coupon struct {
	Discount           Discount       `json:"discount" bson:"discount"`
	CreatedAt          time.Time      `json:"createdAt" bson:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt" bson:"updatedAt"`
	ID                 string         `json:"id" bson:"_id"`
	Code               string         `json:"code" bson:"_code"`
	Valid              bool           `json:"valid" bson:"valid"`
	MaxLife            int            `json:"maxLife" bson:"maxLife"`
	Applicable         DiscountEntity `json:"applicable" bson:"applicable"`
	ApplicableID       string         `json:"applicableID" bson:"applicableID"`
	ApplicableIDs      []string       `json:"applicableIDs" bson:"applicableIDs"`
	Usability          UsabilityType  `json:"usability" bson:"usability"`
	IsInfluencerCoupon bool           `json:"isInfluencerCoupon" bson:"isInfluencerCoupon"`
}

func CreateCoupon(MaxLife int, discount Discount, code string, applicable DiscountEntity, applicableID []string, usability UsabilityType) (coupon Coupon, err error) {
	coupon.Discount = discount
	coupon.Code = code
	coupon.MaxLife = MaxLife
	coupon.Applicable = applicable
	coupon.Usability = usability
	coupon.ApplicableIDs = applicableID
	coupon.ID = data.GetUUIDStringWithoutPrefix()
	coupon.CreatedAt = time.Now()
	coupon.Valid = true
	if coupon.MaxLife != 0 {
		TemporalJobs.CreateCouponExpiryWorkflow(coupon.ID, 24*time.Hour*time.Duration(coupon.MaxLife))
	}
	_, err = db.CouponCollection.InsertOne(context.Background(), coupon)
	return
}

func CreateDiscount(value float64, cap float64, Type DiscountType) (discount Discount) {
	discount.Cap = cap
	discount.Value = value
	discount.Type = Type
	return
}

func (coupon Coupon) CreateIndexes() error {
	t := true
	options := options.Index()
	options.Unique = &t
	indexModels := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "_code", Value: 1},
			},
			Options: options,
		},
	}

	indexName, err := db.CouponCollection.Indexes().CreateMany(context.Background(), indexModels)
	if err != nil {
		if strings.Contains(err.Error(), "Index with name") && strings.Contains(err.Error(), "already exists") {
			fmt.Println("Indexes already exist")
		} else {
			// Handle other errors
			fmt.Println("Error creating indexes:", err)
			return err
		}
	} else {
		fmt.Println("Created index:", indexName)
	}

	return nil
}

func (Coupon) GormDataType() string {
	return "json"
}

func (data Coupon) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *Coupon) Scan(value interface{}) error {

	if value == nil {
		return nil
	}

	var byteSlice []byte
	switch v := value.(type) {
	case []byte:
		if len(v) > 0 {
			byteSlice = make([]byte, len(v))
			copy(byteSlice, v)
		}
	case string:
		byteSlice = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	err := json.Unmarshal(byteSlice, &data)
	return err
}

type Discount struct {
	Type  DiscountType `json:"type"`
	Value float64      `json:"value"`
	Cap   float64      `json:"cap"`
}

type DiscountType int

const (
	Flat DiscountType = iota
	Percent
)

type DiscountEntity int

const (
	ProductDiscount DiscountEntity = iota
	CategoryDiscount
	UserDiscount
	BrandDiscount
)

type UsabilityType int

const (
	SingleUse UsabilityType = iota
	SingleUsePerUser
)

func GetCouponByID(ID string) (coupon Coupon, err error) {
	err = db.CouponCollection.FindOne(context.Background(), bson.M{"_id": ID}).Decode(&coupon)
	if err != nil {
		return
	}
	return
}

func GetCouponByCode(code string) (coupon Coupon, err error) {
	err = db.CouponCollection.FindOne(context.Background(), bson.M{"_code": code}).Decode(&coupon)
	if err != nil {
		return
	}
	return
}

// type CouponApplicabilityCheckFunction func(string, string) (bool, string, error)

// var CouponApplicabilityMap = map[DiscountEntity]map[UsabilityType]CouponApplicabilityCheckFunction{
// 	ProductDiscount: {
// 		SingleUse: productBasedSingleUseCouponApplicability,
// 		// SingleUsePerUser
// 	},
// }

// func productBasedSingleUseCouponApplicability(applicableID string, productID string) (applicable bool, reason string, err error) {
// 	if productID == applicableID {
// 		return true, "Valid", nil
// 	} else {
// 		return false, "Coupon not Valid on this Product", nil
// 	}
// }

// func produtBasedSingleUserPerUser

func (coupon Coupon) IsCouponApplicable(productID string) (applicable bool, reason string, err error) {
	switch coupon.Applicable {
	case ProductDiscount:
		if utils.Includes(coupon.ApplicableIDs, productID) {
			return true, "Valid", nil
		} else {
			return false, "Coupon not Valid on this Product", nil
		}
	default:
		return false, "Coupon Invalid", nil
	}
}

func (coupon Coupon) CalculateDiscount(cartValue float64) (discount float64) {
	switch coupon.Discount.Type {
	case Flat:
		discount = math.Min(cartValue, coupon.Discount.Value)
	case Percent:
		discount = math.Min(coupon.Discount.Cap, math.Floor((cartValue*coupon.Discount.Value)/100))
	}

	return
}

func (coupon *Coupon) MarkAsUsed(userID string) (err error) {
	switch coupon.Usability {
	case SingleUse:
		coupon.Valid = false
		return coupon.Update()
	case SingleUsePerUser:
		_, err = createCouponUseLog(userID, coupon.ID)
		return err
	}

	return fmt.Errorf("unknown Column Type")
}

func createCouponUseLog(userID string, couponID string) (CouponUseRecord, error) {
	var entry = CouponUseRecord{
		ID:       data.GetUUIDString("couponUse"),
		UserID:   userID,
		CouponID: couponID,
	}

	err := Mysql.DB.Create(&entry).Error

	if err != nil {
		fmt.Printf("err: %v\n", err)
		return entry, err
	}

	return entry, nil
}

func IsValidCoupon(couponID string, userID string) (validity bool) {
	coupon, err := GetCouponByID(couponID)
	if err != nil {
		return false
	}

	validity, _ = coupon.IsValid(userID)
	return validity
}

func (coupon *Coupon) Update() (err error) {
	coupon.UpdatedAt = time.Now()
	_, err = db.CouponCollection.UpdateOne(context.Background(), bson.M{"_id": coupon.ID}, bson.M{"$set": coupon})
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func (coupon Coupon) IsValid(userID string) (valid bool, reason string) {

	switch coupon.Usability {
	case SingleUse:
		return coupon.Valid, "Coupon Already Used"
	case SingleUsePerUser:
		couponAlreadyUsed, err := coupon.CheckIfUserhasAlreadyUsedThisCoupon(userID)
		if err != nil {
			return false, "Error Checking Coupon validity"
		}

		return !couponAlreadyUsed, "Coupon Already Used"
	}

	return
}

func (coupon Coupon) Expire() (err error) {
	coupon.Valid = false
	err = coupon.Update()
	return
}

func (coupon Coupon) CheckIfUserhasAlreadyUsedThisCoupon(userID string) (hasUsed bool, err error) {

	// Check if mapping of this User to Coupon exists in Used Coupons Table

	var entry CouponUseRecord

	err = Mysql.DB.Model(&entry).Where("userID = ? AND couponID = ?", userID, coupon.ID).First(&entry).Error

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}

	if err != nil {
		return true, err
	}

	return true, nil
}

type CouponUseRecord struct {
	ID       string `gorm:"column:id"`
	UserID   string `gorm:"column:userID;index:userCouponCombo"`
	CouponID string `gorm:"column:couponID;index:userCouponCombo"`
}

func GetAllCoupons() (coupons []Coupon, err error) {

	cursor, _ := db.CouponCollection.Find(context.Background(), bson.D{})
	err = cursor.All(context.Background(), &coupons)
	if err != nil {
		fmt.Printf("all couopns get err: %v\n", err)
	}

	return
}
