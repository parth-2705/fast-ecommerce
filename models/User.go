package models

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	GCS "hermes/admin/services/gcs"
	"hermes/configs/Mysql"
	"hermes/db"
	"hermes/utils/barcodes"
	"hermes/utils/data"
	"hermes/utils/images"
	"hermes/utils/whatsapp"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID                        string    `json:"id" bson:"_id"`
	CreatedAt                 time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt                 time.Time `json:"updatedAt" bson:"updatedAt"`
	Name                      string    `json:"name" bson:"name"`
	Phone                     string    `json:"phone" bson:"phone"`
	Email                     string    `json:"email" bson:"email"`
	ProfileImage              Image     `json:"profileImage" bson:"profileImage"`
	Internal                  bool      `json:"internal" bson:"internal"`
	MarketingCommDisabled     bool      `json:"marketingCommDisabled" bson:"marketingCommDisabled"`
	RepeatUser                bool      `json:"repeatUser" bson:"repeatUser"`
	ReferredByUser            string    `json:"referredByUser" bson:"referredByUser"`
	HasJoinedReferralProgram  bool      `json:"hasJoinedReferralProgram" bson:"hasJoinedReferralProgram"`
	AmbassdorInvitesSentCount int       `json:"-" bson:"ambassdorInvitesSentCount"`
}

func (User User) CreateIndexes() error {

	name := "phoneNumber"
	t := true
	options := options.Index()
	options.Unique = &t
	options.Name = &name

	indexModel := mongo.IndexModel{
		Keys: bson.M{
			"phone": 1, // 1 for ascending, -1 for descending
		},
		Options: options,
	}

	_, err := db.UserCollection.Indexes().CreateOne(context.Background(), indexModel)
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

func (user User) JoinReferralProgram(name string, city string, interests []string, testGroup int) error {

	if user.HasJoinedReferralProgram {
		return nil
	}

	var referralCode string
	for {
		// Generate a unique referral code for this user
		referralCode = GenerateReferralCode()
		if isReferralCodeUnique(referralCode) {
			break
		}
	}

	user.Name = name

	// Create a session for mongod transaction
	session, err := db.MongoClient.StartSession()
	if err != nil {
		return err
	}

	defer session.EndSession(context.Background())

	// Starting transaction
	err = session.StartTransaction()
	if err != nil {
		return err
	}

	// Assign the referral code
	err = AssignReferralCode(user.ID, referralCode)
	if err != nil {
		session.AbortTransaction(context.Background())
		return err
	}

	// Mark that user has joined the program
	_, err = db.UserCollection.UpdateOne(context.Background(), bson.M{"_id": user.ID}, bson.M{"$set": bson.M{"hasJoinedReferralProgram": true}})
	if err != nil {
		session.AbortTransaction(context.Background())
		return err
	}

	// Commit the transaction
	err = session.CommitTransaction(context.Background())
	if err != nil {
		session.AbortTransaction(context.Background())
		return err
	}

	_, err = user.MakeAmbassdor(city, interests, referralCode, testGroup)
	if err != nil {
		return err
	}
	return nil
}

func (User) GormDataType() string {
	return "json"
}

func (data User) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *User) Scan(value interface{}) error {
	err := GormScan(reflect.ValueOf(data), value)
	if err != nil {
		return err
	}
	return nil
}

type UserToUserAgentMapping struct {
	UserID      string    `json:"userID" bson:"userID"`
	UserAgentID string    `json:"userAgentID" bson:"userAgentID"`
	LoginTime   time.Time `json:"loginTime" bson:"loginTime"`
	LogoutTime  time.Time `json:"logouTime" bson:"logoutTime"`
}

func CheckIfOTPCanbeSent(mobileNumber string) bool {

	var otpLog OTPLog

	err := db.OTPLogCollection.FindOne(context.Background(), bson.M{"mobileNumber": mobileNumber}, options.FindOne().SetSort(bson.D{{"sendTime", -1}})).Decode(&otpLog)
	if err != nil {
		fmt.Println(err)
		return true
	}

	return time.Since(otpLog.SendTime) > (30 * time.Second)

}

func CreateOTPLog(mobileNumber string, userAgentID string, referrer string, dirtyInput string) (otpLog OTPLog, err error) {

	var product Product
	if strings.HasPrefix(referrer, "/product") {
		urlPath := strings.Split(referrer, "/")
		if len(urlPath) >= 3 {
			product, _ = GetCompleteProduct(urlPath[2])
		}
	}

	otpLog = OTPLog{
		ID:           data.GetUUIDString("otpLog"),
		MobileNumber: mobileNumber,
		SendTime:     time.Now(),
		UserAgentID:  userAgentID,
		Referrer:     referrer,
		DirtyInput:   dirtyInput,
		Product:      product,
	}

	_, err = db.OTPLogCollection.InsertOne(context.Background(), otpLog)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = Mysql.DB.Create(&otpLog).Error
	if err != nil {
		fmt.Println(err)
		return
	}

	return
}

func UpdateSuccessStatusOfOTPLog(OtpLogId string) (err error) {
	_, err = db.OTPLogCollection.UpdateOne(context.Background(), bson.D{{"_id", OtpLogId}}, bson.D{{"$set", bson.D{{"success", true}}}})
	if err != nil {
		fmt.Println(err)
	}

	err = Mysql.DB.Table("OTPLogs").Where("id = ?", OtpLogId).UpdateColumn("success", true).Error
	if err != nil {
		fmt.Println(err)
	}

	return err
}

func (user User) Create() (newUser User, err error) {
	user.ID = data.GetUUIDString("user")
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	_, err = db.UserCollection.InsertOne(context.Background(), user)
	if err != nil {
		fmt.Println(err)
		return newUser, err
	}
	return user, nil
}

func (user User) CreateWithReferralCode(code string) (newUser User, err error) {

	if len(code) == 0 {
		return newUser, fmt.Errorf("referral code is empty")
	}

	user.ID = data.GetUUIDString("user")
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Find the user profile of referring user
	var referredByUserProfile Profile
	err = db.ProfileCollection.FindOne(context.Background(), bson.M{"referralCode": code}).Decode(&referredByUserProfile)
	if err != nil {
		// If error occurs return
		if err != mongo.ErrNoDocuments {
			return newUser, err
		}

		// If no referring user is found then continue with normal flow
		err = nil
	}

	user.ReferredByUser = referredByUserProfile.UserID
	_, err = db.UserCollection.InsertOne(context.Background(), user)
	if err != nil {
		fmt.Println(err)
		return newUser, err
	}
	return user, nil
}

func (user *User) MarkAsRepeatUser() (err error) {
	user.RepeatUser = true
	return user.Update()
}

func (user *User) Update() (err error) {
	user.UpdatedAt = time.Now()
	_, err = db.UserCollection.ReplaceOne(context.Background(), bson.M{"_id": user.ID}, user)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (user User) GetProfile() (profile Profile, err error) {
	err = db.ProfileCollection.FindOne(context.Background(), bson.M{"userID": user.ID}).Decode(&profile)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func (user User) GetAddresses() (addresses []Address, err error) {
	addressCursor, err := db.AddressCollection.Find(context.Background(), bson.M{"userID": user.ID}, options.Find())
	if err != nil {
		fmt.Println(err)
		return
	}

	defer addressCursor.Close(context.Background())

	if err = addressCursor.All(context.TODO(), &addresses); err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	return
}

func (user User) GetDefaultAddress() (defaultAddress Address) {
	db.AddressCollection.FindOne(context.Background(), bson.M{"isDefault": true, "userID": user.ID}).Decode(&defaultAddress)
	return
}

func (user User) FindDefaultAddress() (defaultAddress Address, err error) {
	err = db.AddressCollection.FindOne(context.Background(), bson.M{"isDefault": true, "userID": user.ID}).Decode(&defaultAddress)
	return
}

func (user User) GetNonDefaultAddresses() (addresses []Address, err error) {
	addressCursor, err := db.AddressCollection.Find(context.Background(), bson.D{{"userID", user.ID}, {"isDefault", bson.D{{"$not", bson.D{{"$eq", true}}}}}}, options.Find())
	if err != nil {
		fmt.Println(err)
		return
	}

	defer addressCursor.Close(context.Background())

	if err = addressCursor.All(context.TODO(), &addresses); err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	return
}

func (user User) UpdateUserName(username string) (err error) {
	user.UpdatedAt = time.Now()
	user.Name = username
	_, err = db.UserCollection.ReplaceOne(context.Background(), bson.M{"_id": user.ID}, user)
	return
}

func UserExists(phone string) bool {
	var user User
	err := db.UserCollection.FindOne(context.Background(), bson.M{"phone": phone}).Decode(&user)
	if err != nil {
		fmt.Println(err)
		return false
	}
	if user.ID == "" {
		return false
	}
	return true
}

func GetUser(phone string) (user User, err error) {
	err = db.UserCollection.FindOne(context.Background(), bson.M{"phone": phone}).Decode(&user)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func GetUserByEmail(email string) (user User, err error) {
	err = db.UserCollection.FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func ToggleUserInternalExternalStatus(user User) (err error) {
	_, err = db.UserCollection.UpdateOne(context.Background(), bson.M{"email": user.Email}, bson.M{"$set": bson.M{
		"internal": user.Internal,
	}})
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func (user User) GetAllOrders() ([]Order, error) {

	cursor, err := db.OrderCollection.Find(context.Background(), bson.M{"userId": user.ID})
	if err != nil {
		return nil, err
	}

	defer cursor.Close(context.Background())

	var orders []Order

	err = cursor.All(context.Background(), &orders)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func GetAllUsers() (users []User, err error) {
	cursor, err := db.UserCollection.Find(context.Background(), bson.M{})
	if err != nil {
		fmt.Println(err)
		return
	}

	defer cursor.Close(context.Background())

	if err = cursor.All(context.TODO(), &users); err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	return
}

func GetUserByID(id string) (user User, err error) {
	err = db.UserCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&user)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

// func to update default payment method
func (profile Profile) UpdateDefaultPaymentMethod(paymentMethod string) (err error) {
	_, err = db.ProfileCollection.UpdateOne(context.Background(), bson.M{"userID": profile.UserID}, bson.M{"$set": bson.M{"defaultPaymentMethod": paymentMethod}})
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

// func to update last used payment method
func (profile Profile) UpdateLastUsedPaymentMethod(paymentMethod string) (err error) {
	_, err = db.ProfileCollection.UpdateOne(context.Background(), bson.M{"userID": profile.UserID}, bson.M{"$set": bson.M{"lastUsedPaymentMethod": paymentMethod}})
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func CreateUserToUserAgentMapping(userID string, userAgentID string) (UserToUserAgentMapping, error) {

	mapping := UserToUserAgentMapping{
		UserID:      userID,
		UserAgentID: userAgentID,
		LoginTime:   time.Now(),
	}

	_, err := db.UserUserAgentMappingCollection.InsertOne(context.Background(), mapping)
	if err != nil {
		fmt.Println(err)
		return UserToUserAgentMapping{}, err
	}

	return mapping, nil
}

func UpdateUserUserAgenMappingWithLogoutTime(userID string, userAgentID string) (err error) {

	filter := bson.D{
		{Key: "$and",
			Value: bson.A{
				bson.D{{Key: "userID", Value: userID}},
				bson.D{{Key: "userAgentID", Value: userAgentID}},
			}},
	}

	_, err = db.UserUserAgentMappingCollection.UpdateOne(context.Background(), filter, bson.M{"$set": bson.M{"logoutTime": time.Now()}})
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func (user *User) TurnOffMarketingComm() error {
	user.MarketingCommDisabled = true
	return user.Update()
}

func (user User) LastAbandonedCart(cartID string) bool {
	var lastAbandonedCart Cart

	filter := bson.D{
		{Key: "userID", Value: user.ID},
		{Key: "status", Value: bson.D{{Key: "$lte", Value: 1}}},
	}
	options := options.FindOneOptions{
		Sort: bson.M{"createdAt": -1},
	}

	err := db.CartCollection.FindOne(context.Background(), filter, &options).Decode(&lastAbandonedCart)
	if err != nil {
		return false
	}

	return lastAbandonedCart.ID == cartID

}

func (user User) KillAllACRWorkflows() error {

	defer sentry.Recover()

	op := options.Find()
	op.Sort = bson.M{"createdAt": -1}
	cursor, err := db.CartCollection.Find(context.Background(), bson.M{"userID": user.ID, "createdAt": bson.M{"$gt": time.Now().AddDate(0, 0, -3)}}, op)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return err
	}

	var carts []Cart
	err = cursor.All(context.Background(), &carts)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return err
	}

	for _, cart := range carts {
		cart.KillACRWorkflow("User Placed Order")
	}

	return nil
}

func (user User) WasACartCompletedInLastXDays(x int) bool {

	filter := bson.D{
		{Key: "userID", Value: user.ID},                                        // Cart was of this User
		{Key: "status", Value: OrderCompleted},                                 // Cart was completed
		{Key: "updatedAt", Value: bson.M{"$gt": time.Now().AddDate(0, 0, -x)}}, // Cart was completed in last x days
	}
	options := options.FindOneOptions{
		Sort: bson.M{"createdAt": -1},
	}

	err := db.OrderCollection.FindOne(context.Background(), filter, &options).Err()

	return err != nil
}

func (user User) GetLastCompletedOrder() (order Order, err error) {

	filter := bson.D{
		{Key: "userId", Value: user.ID},
		{Key: "paymentStatus", Value: "Paid"},
	}

	options := options.FindOneOptions{
		Sort: bson.M{"createdAt": -1},
	}

	err = db.OrderCollection.FindOne(context.Background(), filter, &options).Decode(&order)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}

	return
}

func (user User) GetWallet() (Wallet, error) {

	return GetUserWallet(user.ID)

}

func GetAllAmbassdors() (ambassdors []User, err error) {

	cursor, _ := db.UserCollection.Find(context.Background(), bson.M{"hasJoinedReferralProgram": true})
	err = cursor.All(context.Background(), &ambassdors)
	if err != nil {
		fmt.Printf("Ambassdors list err: %v\n", err)
	}

	return
}

func GetAllInternalUsers() (users []User, err error) {
	cursor, _ := db.UserCollection.Find(context.Background(), bson.M{"internal": true})
	err = cursor.All(context.Background(), &users)
	if err != nil {
		fmt.Printf("Ambassdors To Be List Err: %v\n", err)
	}

	return

}

func GetUsersByPhoneNumbers(phoneNumbers []string) (users []User, err error) {
	orCondition := bson.A{bson.D{{Key: "phone", Value: ""}}}

	for _, phNo := range phoneNumbers {

		orCondition = append(orCondition, bson.D{{Key: "phone", Value: phNo}})

	}

	cursor, _ := db.UserCollection.Find(context.Background(), bson.D{{Key: "$or", Value: orCondition}})
	err = cursor.All(context.Background(), &users)
	if err != nil {
		fmt.Printf("err getting Users: %v\n", err)
	}

	return
}

func GetAllAmbassdorsTobe() (ambassdorsToBe []User, err error) {

	// return GetAllInternalUsers()

	cursor, _ := db.UserCollection.Find(context.Background(), bson.M{"hasJoinedReferralProgram": false, "marketingCommDisabled": false})
	err = cursor.All(context.Background(), &ambassdorsToBe)
	if err != nil {
		fmt.Printf("Ambassdors To Be List Err: %v\n", err)
	}

	return
}

func (user *User) SendAmbassdorProgramInvite() (err error) {

	if user.HasJoinedReferralProgram || user.MarketingCommDisabled {
		return nil
	}

	err = whatsapp.SendAmbassdorRecruitmentMessage(user.Phone)
	if err != nil {
		return err
	}

	user.AmbassdorInvitesSentCount++
	user.Update()

	return nil
}

func formDealText(productID string, referalCode string, productStrStub string) (string, error) {

	// // Get Product Information
	// product, err := getCompleteProduct(productID)
	// if err != nil {
	// 	return "", err
	// }

	// // Truncate productName
	// var productName string
	// if len(product.Name) <= 30 {
	// 	productName = product.Name
	// } else {
	// 	productName = product.Name[:30] + "..."
	// }

	messageString := fmt.Sprintf("Hey, I would like to know more about %s.\r\nhttps://roovo.in/product/%s?code=%s", productStrStub, productID, referalCode)
	// messageString := fmt.Sprintf("https://roovo.in/product/%s?referalCode=%s", product.ID, referalCode)

	urlEncodedMsgStr := url.QueryEscape(messageString) // Escape this string as this goes as query parameter to the QR Code encoded URL
	// urlEncodedMsgStr = "yoyoyo"

	qrStr := fmt.Sprintf("https://wa.me/919910606373?text=%s", urlEncodedMsgStr)

	return qrStr, nil
}

func (ambassdor *User) GetDealImage(productID string, productName string) (imageLink string, err error) {

	if !ambassdor.HasJoinedReferralProgram {
		return "", fmt.Errorf("not an Ambassdor")
	}

	profile, err := ambassdor.GetProfile()
	if err != nil {
		return
	}
	referralCode := profile.ReferralCode

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

func CreateUser(phone string, customerName string) (User, error) {

	user := User{
		ID:        data.GetUUIDString("user"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      customerName,
		Phone:     phone,
	}

	_, err := user.CreateProfile()
	if err != nil {
		return user, err
	}

	_, err = db.UserCollection.InsertOne(context.Background(), user)
	if err != nil {
		fmt.Printf("user create err: %v\n", err)
		return user, err
	}

	return user, nil
}

func CSVToAmbassdor(records [][]string, testGroup int) error {

	validCSV := true

	for i := range records {

		if len(records[i]) != 3 {
			validCSV = false
			break
		}

		for j := range records[i] {
			if records[i][j] == "" {
				validCSV = false
				break
			}
		}

		if !validCSV {
			return fmt.Errorf("invalid CSV")
		}

	}

	for i := range records {

		phoneNumber := records[i][1]
		if string(phoneNumber[0]) != "+" {
			phoneNumber = "+" + phoneNumber
		}

		customerName := records[i][0]
		city := records[i][2]

		user, err := GetUser(phoneNumber)
		if err != nil {
			user, err = CreateUser(phoneNumber, customerName)
			if err != nil {
				continue
			}
		}

		err = user.JoinReferralProgram(customerName, city, []string{"", "", ""}, testGroup)
		if err != nil {
			continue
		}
	}

	return nil
}
