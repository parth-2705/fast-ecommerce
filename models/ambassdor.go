package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	GCS "hermes/admin/services/gcs"
	"hermes/configs/Mysql"
	"hermes/utils/barcodes"
	"hermes/utils/data"
	"hermes/utils/images"
	"os"
	"time"
)

type Ambassdor struct {
	ID           string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
	Phone        string `gorm:"index;size:200"`
	Name         string
	UserID       string `gorm:"index;size:200"`
	City         string
	Interests    AmbassdorInterests
	ReferralCode string `gorm:"uniqueIndex;size:200"`
	TutorialSent bool
	DealsSent    int
	Referrals    ReferralRecords
	TestGroup    int `gorm:"default:0"`
}

type ReferralRecords map[string]map[string]time.Time

func (ReferralRecords) GormDataType() string {
	return "json"
}

func (data ReferralRecords) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *ReferralRecords) Scan(value interface{}) error {

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

type AmbassdorInterests []string

func (AmbassdorInterests) GormDataType() string {
	return "json"
}

func (data AmbassdorInterests) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *AmbassdorInterests) Scan(value interface{}) error {

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

func (user *User) MakeAmbassdor(city string, interests []string, referralCode string, testGroup int) (Ambassdor, error) {

	ambassdorName := user.Name
	if ambassdorName == "" {
		ambassdorName = "Ambassdor"
	}

	ambassdor := Ambassdor{
		ID:           data.GetUUIDString("ambassdor"),
		Phone:        user.Phone,
		Name:         user.Name,
		UserID:       user.ID,
		City:         city,
		Interests:    interests,
		ReferralCode: referralCode,
		TestGroup:    testGroup,
	}

	err := Mysql.DB.Create(&ambassdor).Error
	if err != nil {
		fmt.Printf("ambassdor create err: %v\n", err)
		return ambassdor, nil
	}

	return ambassdor, nil
}

func GetAmbassadorByID(ambassadorID string) (ambassdor Ambassdor, err error) {
	err = Mysql.DB.Model(&ambassdor).Where("id = ?", ambassadorID).First(&ambassdor).Error
	if err != nil {
		fmt.Printf("ambassdor get err: %v\n", err)
	}

	return
}

func GetAmbassdorByUserID(userID string) (ambassdor Ambassdor, err error) {

	err = Mysql.DB.Model(&ambassdor).Where("user_id = ?", userID).First(&ambassdor).Error
	if err != nil {
		fmt.Printf("ambassdor get err: %v\n", err)
	}

	return
}

func (ambassdor *Ambassdor) Update() (err error) {

	err = Mysql.DB.Save(&ambassdor).Error
	if err != nil {
		fmt.Printf("ambassdor update err: %v\n", err)
	}

	return
}

func (ambassdor *Ambassdor) IncreaseDealSentCount() (err error) {
	ambassdor.DealsSent = ambassdor.DealsSent + 1
	return ambassdor.Update()
}

func GetAmbassdorByReferralCode(referralCode string) (ambassdor Ambassdor, err error) {
	err = Mysql.DB.Model(&ambassdor).Where("referral_code = ?", referralCode).First(&ambassdor).Error
	if err != nil {
		fmt.Printf("Ambassdor find err: %v\n", err)
	}

	return
}

type ReferralRecord struct {
	ID             string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	AmbassadorID   string
	ReferredUserID string
	ProductID      string
	OrderID        string
}

func (ambassdor *Ambassdor) AddReferralRecord(referredUserID string, referredProductID string) (referralRecord ReferralRecord, err error) {
	// if ambassdor.Referrals == nil {
	// 	ambassdor.Referrals = make(ReferralRecords)
	// }

	// if ambassdor.Referrals[referredProductID] == nil {
	// 	ambassdor.Referrals[referredProductID] = make(map[string]time.Time)
	// }

	// ambassdor.Referrals[referredProductID][referredUserID] = time.Now()
	// err = ambassdor.Update()
	// if err != nil {
	// 	return
	// }

	referralRecord = ReferralRecord{
		ID:             data.GetUUIDString("RR"),
		AmbassadorID:   ambassdor.ID,
		ReferredUserID: referredUserID,
		ProductID:      referredProductID,
	}

	err = Mysql.DB.Model(&referralRecord).Create(&referralRecord).Error
	if err != nil {
		fmt.Printf("RR create err: %v\n", err)
		return
	}

	return
}

func GetReferralRecord(recordID string) (referralRecord ReferralRecord, err error) {

	err = Mysql.DB.Model(&referralRecord).Where("id = ?", recordID).First(&referralRecord).Error
	if err != nil {
		fmt.Printf("referral record get err: %v\n", err)
	}
	return
}

func (rr *ReferralRecord) MarkConverted(orderID string) error {
	rr.OrderID = orderID
	return rr.Update()
}

func (rr *ReferralRecord) Update() error {

	err := Mysql.DB.Save(rr).Error
	if err != nil {
		fmt.Printf("RR Save err: %v\n", err)
	}

	return err
}

func GetAllInternalAmbassdors() (internalAmbassdors []Ambassdor, err error) {

	internalUsers, err := GetAllInternalUsers()
	if err != nil {
		return
	}

	for _, internalUser := range internalUsers {

		ambassador, err := GetAmbassdorByUserID(internalUser.ID)
		if err != nil {
			continue
		}

		internalAmbassdors = append(internalAmbassdors, ambassador)

	}

	return
}

func GetAmbassadorsByTestGroup(testGroup int) (ambassadors []Ambassdor, err error) {

	err = Mysql.DB.Model(&ambassadors).Where("test_group = ?", testGroup).Find(&ambassadors).Error
	if err != nil {
		fmt.Printf("mysql get err: %v\n", err)
	}

	return
}

func (ambassdor *Ambassdor) GetDealImage(productID string, productName string) (imageLink string, err error) {

	referralCode := ambassdor.ReferralCode

	// Create the string to encode in QR Code
	qrStr, err := formDealText(productID, referralCode, productName)
	if err != nil {
		return
	}

	fileName := fmt.Sprintf("deal-%s", ambassdor.ID)

	// Make the QR Code
	err = barcodes.MakeQRCodeAndStoreOnDisk(qrStr, fileName)
	if err != nil {
		return
	}

	// SuperImpose QR over ad creative
	err = images.DealImageWithQR(fileName)
	if err != nil {
		return
	}

	// Upload to Bucket
	imageLink, err = GCS.UploafFileFromDiskToBucket(fileName)
	if err != nil {
		return
	}

	// Remove local image
	err = os.Remove(fileName)
	if err != nil {
		fmt.Printf("Deleting file err: %v\n", err)
	}

	return
}
