package models

import (
	"context"
	"errors"
	"fmt"
	"hermes/configs/Mysql"
	"hermes/db"
	"hermes/models/Logs"
	"hermes/services/Sentry"
	"hermes/services/Temporal/TemporalJobs"
	fb "hermes/services/fbAds"
	"hermes/services/sendgridEmail"
	"hermes/utils/data"
	"hermes/utils/whatsapp"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Order struct {
	ID        string    `json:"id" bson:"_id" gorm:"column:id"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt" gorm:"column:createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt" gorm:"column:updatedAt"`
	UserID    string    `json:"userId" bson:"userId" gorm:"column:userId"`
	User      User      `json:"user" bson:"user" gorm:"column:user"`
	Product   Product   `json:"product" bson:"product" gorm:"column:product"`
	Address   Address   `json:"address" bson:"address" gorm:"column:address"`
	// Store              Store       `json:"store" bson:"store" gorm:"column:store"`
	PaymentStatus      string         `json:"paymentStatus" bson:"paymentStatus" gorm:"column:paymentStatus"`
	Payment            Payment        `json:"payment" bson:"payment" gorm:"column:payment"`
	FulfillmentStatus  string         `json:"fulfillmentStatus" bson:"fulfillmentStatus" gorm:"column:fulfillmentStatus"`
	ShipmentStatus     string         `json:"shipmentStatus" bson:"shipmentStatus" gorm:"column:shipmentStatus"`
	CancellationReason string         `json:"cancellationReason" bson:"cancellationReason" gorm:"column:cancellationReason"`
	CancellationTime   *time.Time     `json:"cancellationTime" bson:"cancellationTime" gorm:"cancellationTime"`
	Brand              Brand          `json:"brand" bson:"brand" gorm:"column:brand"`
	Variant            Variation      `json:"variant" bson:"variant" gorm:"column:varaint"`
	VariantID          string         `json:"variantId" bson:"variantId" gorm:"column:varaintId"`
	DealID             string         `json:"dealID" bson:"dealID" gorm:"column:dealID"`
	Deal               Deal           `json:"deal" bson:"deal" gorm:"column:deal"`
	Coupon             Coupon         `json:"coupon" bson:"coupon" gorm:"column:coupon"`
	CartID             string         `json:"cartID" bson:"cartID" gorm:"column:cartID"`
	Cart               Cart           `json:"cart" bson:"cart" gorm:"column:cart"`
	OrderAmount        OrderAmount    `json:"orderAmount" bson:"orderAmount" gorm:"column:orderAmount"`
	AdditionalDiscount int64          `json:"additionalDiscount" bson:"additionalDiscount" gorm:"column:additionalDiscount"`
	Source             PortalType     `json:"source" bson:"source" gorm:"column:source"`
	ParentID           string         `json:"parentID" bson:"parentID" gorm:"column:parentID"`
	Session            Logs.Session   `json:"session" bson:"session" gorm:"column:session"`
	IPAddress          string         `json:"ip" bson:"ip" gorm:"column:ip"`
	UserAgent          string         `json:"userAgent" bson:"userAgent" gorm:"column:userAgent"`
	UTMParams          data.UTMParams `json:"utmParams" bson:"utmParams" gorm:"column:utmParams"`
	Commission         float64        `json:"commission" bson:"commission"`
	Remitted           bool           `json:"remitted" bson:"remitted"`
	Fulfillable        bool           `json:"fulfillable" bson:"fulfillable"`
	ReferralRecordID   string         `json:"referralRecordID" bson:"referralRecordID"`
}

type OrderWithTracking struct {
	Order    `bson:",inline"`
	Tracking ShipRocketTracking `json:"tracking" bson:"tracking"`
}

func (Order) TableName() string {
	return "Orders"
}

func (orderAmt *OrderAmount) calculatePaymentMethodDiscount(cartAmount float64) {
	if orderAmt.PaymentMethodDiscount.DiscountPercentage > 0 {
		orderAmt.PaymentMethodDiscount.DiscountAmount = float64(orderAmt.PaymentMethodDiscount.DiscountPercentage) * cartAmount / 100
		if orderAmt.PaymentMethodDiscount.DiscountAmount > 30 {
			orderAmt.PaymentMethodDiscount.DiscountAmount = 30
		}
		orderAmt.TotalAmount = cartAmount - orderAmt.PaymentMethodDiscount.DiscountAmount
	} else if orderAmt.PaymentMethodDiscount.DiscountAmount > 0 {
		orderAmt.TotalAmount = cartAmount - orderAmt.PaymentMethodDiscount.DiscountAmount
	}
	//handling for negative amount
	if orderAmt.TotalAmount < 0 {
		orderAmt.PaymentMethodDiscount.DiscountAmount += orderAmt.TotalAmount
		orderAmt.TotalAmount = 0
	}
	orderAmt.TotalAmount = math.Round(orderAmt.TotalAmount)
}

func (amount *OrderAmount) AddPaymentMethodDiscountAndCalculateDiscount(method string, cartAmount float64) {
	if method != "COD" && method != string(WALLET) {
		amount.PaymentMethodDiscount.DiscountPercentage = 5
		amount.calculatePaymentMethodDiscount(cartAmount)
	} else {
		amount.PaymentMethodDiscount.DiscountPercentage = 0
		amount.PaymentMethodDiscount.DiscountAmount = 0
		amount.TotalAmount = cartAmount
	}
}

func MarkOrdersAsRemitted(IDs []string) (err error) {
	_, err = db.OrderCollection.UpdateMany(context.Background(), bson.M{"_id": bson.M{"$in": IDs}}, bson.M{"$set": bson.M{"remitted": true}})
	return
}

func (Order Order) CreateIndexes() error {

	indexModels := []mongo.IndexModel{}
	userIDModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "userId", Value: 1},
		},
	}
	indexModels = append(indexModels, userIDModel)

	createdAtModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "createdAt", Value: -1},
		},
	}
	indexModels = append(indexModels, createdAtModel)

	paymentStatusModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "paymentStatus", Value: 1},
		},
	}
	indexModels = append(indexModels, paymentStatusModel)

	sellerIDModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "product.sellerID", Value: 1},
		},
	}
	indexModels = append(indexModels, sellerIDModel)

	brandIDModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "product.brandID", Value: 1},
		},
	}
	indexModels = append(indexModels, brandIDModel)

	// awbTextIndex := mongo.IndexModel{
	// 	Keys: bson.M{
	// 		"awb": "text",
	// 	},
	// }
	// indexModels = append(indexModels, awbTextIndex)

	indexName, err := db.OrderCollection.Indexes().CreateMany(context.Background(), indexModels)
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

func MigrateOrderToMySQL(order Order) (err error) {

	defer sentry.Recover()

	err = Mysql.DB.Create(&order).Error
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func (order Order) MarkOrderAsCompleted(payment Payment) (err error) {
	order.PaymentStatus = "Paid"
	order.Payment = payment
	err = order.Update()
	if err != nil {
		return err
	}

	// Mark Coupon if used as expired
	if order.Coupon.ID != "" {
		err = order.Coupon.MarkAsUsed(order.UserID)
		if err != nil {
			return err
		}
	}

	cart, err := GetCart(order.CartID)
	if err != nil {
		return
	}

	cart.OrderCompleted()

	user, err := GetUserByID(order.UserID)
	if err != nil {
		return err
	}

	user.MarkAsRepeatUser()

	wallet, err := user.GetWallet()
	if err != nil {
		return err
	}

	go order.AddToRefferalRecord()

	// Kill all currently running ACR Workflows for this User
	go user.KillAllACRWorkflows()

	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	go sendgridEmail.SendOrderPlacedEmail(ctx, sendgridEmail.EmailNotifier{
		UserPhone:      user.Phone,
		ProductID:      order.Product.ID,
		ProductName:    order.Product.Name,
		ProductURL:     fmt.Sprintf("https://roovo.in/product/%s", order.Product.ID),
		BrandID:        order.Brand.ID,
		BrandName:      order.Brand.Name,
		UserID:         user.ID,
		OrderID:        order.ID,
		PaymentMethod:  order.Payment.Method,
		Price:          strconv.FormatFloat(order.Cart.CartAmount.TotalAmount, 'f', 2, 64),
		AddressID:      order.Address.ID,
		ContactName:    order.Address.Name,
		Address:        order.Address.GetAddessString(),
		Quantity:       order.Cart.Items[0].Quantity,
		ProductPrice:   order.Cart.Items[0].Variant.Price.SellingPrice,
		CouponDiscount: order.OrderAmount.Coupon.DiscountAmount,
		CouponID:       order.OrderAmount.Coupon.CouponID,
		RPBalance:      wallet.Balance,
		ReferredBy:     user.ReferredByUser,
	})

	if order.Source == Whatsapp {
		// fb.SendChatPurchaseEvent(order.OrderAmount.TotalAmount, order.ID, order.User.Phone, order.Address.PinCode)
		go whatsapp.MarkOrderCompletedOnWhatsapp(user.Phone)
	}
	return

}

func (order Order) MarkOrderAsFulfillable() (err error) {
	if order.Fulfillable {
		return
	}
	if len(order.CancellationReason) > 0 {
		return errors.New("Order cancelled")
	}
	user, err := GetUserByID(order.UserID)
	if err != nil {
		return err
	}
	err = order.SetFullfillable(true)
	if err != nil {
		return
	}

	// a function that runs in a goroutine to send whatsapp message
	go whatsapp.SendOrderConfirmationMessage2(user.Phone, order.Address.Name, order.Product.Name, order.ID)
	go MigrateOrderToMySQL(order)
	if user.Internal {

		go func() {

			defer sentry.Recover()

			conf, err := GetTemplateConfig(order.Cart.Items[0].ProductID)
			if err != nil {
				return
			}

			TemporalJobs.CreateReOrderMessageWorkflow(order.ID, user.Internal, conf.ProductID, conf.SleepTime)
		}()
	}

	if order.Source == Website {
		fb.SendPurchaseEvent(order.OrderAmount.TotalAmount, order.ID, order.Session.UserAgent, order.Session.IPAddress, order.Session.FBC, order.Session.FBP, order.User.Phone, order.Address.PinCode)
	}
	return
}

func (order Order) MarkOrderAsCODUnconfirmed() (err error) {
	name := order.Address.Name
	if name == "" {
		name = "there"
	}
	err = whatsapp.SendCODConfirmationTemplate("https://storage.googleapis.com/roovo-images/rawImages/"+string(order.Cart.Items[0].Product.Thumbnail), order.Address.Phone, name, order.Cart.Items[0].Product.Name, order.Cart.Items[0].Quantity, int(order.OrderAmount.TotalAmount), order.ID)
	return
}

func (order Order) Create() (err error) {
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	_, err = db.OrderCollection.InsertOne(context.Background(), order)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func (order Order) CreateWithOutTimeUpdate() (err error) {
	_, err = db.OrderCollection.InsertOne(context.Background(), order)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func (order Order) Update() (err error) {
	order.UpdatedAt = time.Now()
	_, err = db.OrderCollection.UpdateOne(context.Background(), bson.M{"_id": order.ID}, bson.M{"$set": order})
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func GetOrder(id string) (order Order, err error) {
	err = db.OrderCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&order)
	return
}

func GetOrderByUserAndID(userId string, id string) (order Order, err error) {
	err = db.OrderCollection.FindOne(context.Background(), bson.M{"_id": id, "userId": userId}).Decode(&order)
	return
}

func GetAllOrders(userID string) (orders []OrderWithTracking, err error) {
	ordersCursor, err := db.OrderCollection.Aggregate(context.Background(), bson.A{bson.D{{Key: "$match", Value: bson.M{"userId": userID}}}, bson.D{{Key: "$lookup", Value: bson.M{"from": "tracking", "foreignField": "orderid", "localField": "_id", "as": "tracking"}}}, bson.D{{Key: "$unwind", Value: bson.M{"path": "$tracking", "preserveNullAndEmptyArrays": true}}}, bson.D{{Key: "$sort", Value: bson.M{"createdAt": -1}}}})
	if err != nil {
		fmt.Println(err)
		return
	}

	defer ordersCursor.Close(context.Background())

	err = ordersCursor.All(context.Background(), &orders)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func GetAllOrdersInDB() (orders []Order, err error) {
	ordersCursor, err := db.OrderCollection.Aggregate(context.Background(), bson.A{})
	if err != nil {
		fmt.Println(err)
		return
	}

	defer ordersCursor.Close(context.Background())

	err = ordersCursor.All(context.Background(), &orders)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func GetOrderCount(userID string) (orderCount int64, err error) {
	orderCount, err = db.OrderCollection.CountDocuments(context.Background(), bson.M{"userId": userID, "paymentStatus": "Paid"})
	return
}

type OrderCountStruct struct {
	ID    string `bson:"_id"`
	Total int64  `bson:"total"`
}

func GetTotalOrderValue(userID string) (orderValue int64, err error) {
	aggregateSearchObject := bson.A{bson.D{{Key: "$match", Value: bson.D{{Key: "userId", Value: userID}, {Key: "paymentStatus", Value: "Paid"}}}}}
	pipeline := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "total", Value: bson.D{
				{Key: "$sum", Value: "$orderAmount.totalAmount"},
			}},
			{Key: "_id", Value: "bruh"},
		}},
	}
	// pipeline := bson.D{{Key: "$group", Value: bson.M{"total": bson.M{"$sum": "$orderAmount.totalAmount"}, "_id": "sumGroup"}}}

	aggregateSearchObject = append(aggregateSearchObject, pipeline)
	cursor, err := db.OrderCollection.Aggregate(context.Background(), aggregateSearchObject)
	if err != nil {
		fmt.Println("err in order summary: ", err)
	}
	fmt.Println("cursor", cursor)
	var orderCountStruct OrderCountStruct
	cursor.Next(context.Background())
	err = cursor.Decode(&orderCountStruct)
	fmt.Println("orderCountStruct", orderCountStruct)
	orderValue = orderCountStruct.Total
	return
}

func (order Order) MarkOrderDelivered(status string) (err error) {

	user, err := GetUserByID(order.UserID)
	if err != nil {
		return err
	}

	go whatsapp.SendOrderDeliveredMessage(user.Phone, order.Address.Name, order.Product.Name, order.Product.ID)

	TemporalJobs.CreateProductReviewMessageWorkflow(order.ID, user.Internal)

	err = order.SetShipmentStatus(status)
	return
}

func (order Order) SetFullfillable(fulfillable bool) (err error) {
	order.Fulfillable = fulfillable
	err = order.Update()
	return
}

func (order Order) MarkOrderDelivered2(status string, currentTimestamp string) (err error) {

	user, err := GetUserByID(order.UserID)
	if err != nil {
		return err
	}

	err = order.SetShipmentStatus(status)
	if err != nil {
		return err
	}

	err = UpdateAverageDeliveryTimeByOrder(order, currentTimestamp)
	if err != nil {
		return err
	}
	go func() {

		defer sentry.Recover()

		whatsapp.SendOrderDeliveredMessage2(user.Phone, order.Address.Name, order.Product.Name, order.Product.ID)
		TemporalJobs.CreateProductReviewMessageWorkflow(order.ID, user.Internal)

		conf, err := GetTemplateConfig(order.Cart.Items[0].ProductID)
		if err != nil {
			return
		}

		TemporalJobs.CreateReOrderMessageWorkflow(order.ID, user.Internal, conf.ProductID, conf.SleepTime)
	}()

	return
}

func (order Order) MarkOrderAsOutForDelivery(status string) (err error) {
	user, err := GetUserByID(order.UserID)
	if err != nil {
		return err
	}

	go whatsapp.SendOrderOutForDeliveryMessage(user.Phone, order.Address.Name, order.Product.Name)

	err = order.SetShipmentStatus(status)
	return
}

func (order Order) MarkOrderAsInTransit(etd string, status string) (err error) {

	user, err := GetUserByID(order.UserID)
	if err != nil {
		return err
	}

	customerName, _ := parseUserInfo(user)

	loc, _ := time.LoadLocation("Asia/Calcutta")
	deliveryDate, _ := time.ParseInLocation("2006-01-02 15:04:05", etd, loc)
	diffTime := time.Until(deliveryDate)
	deliveryDays := int((diffTime/time.Hour)/24) + 1

	err = whatsapp.SendOrderInTransitMessage(user.Phone, customerName, order.Cart.Items[0].Product.Name, strconv.FormatInt(int64(deliveryDays), 10), order.ID)
	if err != nil {
		return err
	}

	err = order.SetShipmentStatus(status)
	return err
}

func (order Order) MarkOrderAsOutForPickup(status string) (err error) {
	// user, err := GetUserByID(order.UserID)
	// if err != nil {
	// 	return err
	// }

	// customerName, _ := parseUserInfo(user)

	// err = whatsapp.SendOrderOutForPickupMessage(user.Phone, customerName, order.Cart.Items[0].Product.Name, order.ID)
	// if err != nil {
	// 	return err
	// }
	err = order.SetShipmentStatus(status)
	return err
}

func parseUserInfo(user User) (customerName string, sendMessage bool) {

	if user.Phone == "" {
		return
	}

	customerName = "there"
	if user.Name != "" {
		customerName = user.Name
	}

	if user.MarketingCommDisabled {
		return
	}

	sendMessage = true
	return

}

func (order Order) MarkOrderAsShipped(etd string, status string) (err error) {
	user, err := GetUserByID(order.UserID)
	if err != nil {
		return err
	}

	loc, _ := time.LoadLocation("Asia/Calcutta")
	filterStartDate, _ := time.ParseInLocation("2006-01-02 15:04:05", etd, loc)

	go whatsapp.SendOrderShippedMessage(user.Phone, order.Address.Name, order.Product.Name, filterStartDate.Format("2006/01/02"), order.ID)

	err = order.SetShipmentStatus(status)
	return
}

func (order Order) MarkOrderAsPickedUp(etd string, status string) (err error) {
	user, err := GetUserByID(order.UserID)
	if err != nil {
		return err
	}

	loc, _ := time.LoadLocation("Asia/Calcutta")
	filterStartDate, _ := time.ParseInLocation("2006-01-02 15:04:05", etd, loc)

	go whatsapp.SendOrderPickedUpMessage(user.Phone, order.Address.Name, order.Product.Name, filterStartDate.Format("2006/01/02"), order.ID)

	err = order.SetShipmentStatus(status)
	return
}

func (order Order) SetShipmentStatus(shipmentStatus string) (err error) {
	order.ShipmentStatus = shipmentStatus
	err = order.Update()
	return
}

func UpdateAverageDeliveryTimeByOrder(order Order, currentTimestamp string) (err error) {
	currentTime, err := time.Parse("02 01 2006 15:04:05", currentTimestamp)
	if err != nil {
		return
	}
	timeDiff := currentTime.Sub(order.CreatedAt).Hours() / 24
	days := math.Round(timeDiff)
	for _, val := range order.Cart.Items {
		err = UpdateAverageDeliveryTimeForProduct(val.ProductID, int(days))
		if err != nil {
			return
		}
	}
	return
}

func (order Order) IsEligibleForCashback() (bool, error) {

	// Check if Payment is remitted
	remittanceStatus, err := IsOrderPaymentRemitted(order)
	if err != nil {
		fmt.Println("err: IsOrderPaymentRemitted", err.Error())
		Sentry.SentryCaptureException(err)
		return false, err
	}

	// Check if their is any return for this order or not
	hasReturn, err := OrderHasReturnOrder(order)
	if err != nil {
		fmt.Println("err: OrderHasReturnOrder", err.Error())
		Sentry.SentryCaptureException(err)
		return false, err
	}

	fmt.Println("remittanceStatus", remittanceStatus)
	fmt.Println("hasReturn", hasReturn)

	if remittanceStatus || !hasReturn {
		return true, nil
	}
	return false, nil
}

func (order Order) AddToRefferalRecord() (ReferralRecord, error) {

	defer sentry.Recover()

	record, err := GetReferralRecord(order.ReferralRecordID)
	if err != nil {
		return record, err
	}

	err = record.MarkConverted(order.ID)
	return record, err
}
