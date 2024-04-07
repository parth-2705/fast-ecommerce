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
	"hermes/utils/data"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type cartStatus int

const (
	Initiated cartStatus = iota
	OrderCreated
	OrderCompleted
	Dummy
)

type Cart struct {
	ID            string      `json:"id" bson:"_id"`
	CreatedAt     time.Time   `json:"createdAt" bson:"createdAt"`
	Source        PortalType  `json:"source" bson:"source"`
	UpdatedAt     time.Time   `json:"updatedAt" bson:"updatedAt"`
	UserAgentID   string      `json:"userAgentID" bson:"userAgentID"`
	UserID        string      `json:"userID" bson:"userID"`
	ProductID     string      `json:"productID" bson:"productID"`
	Store         string      `json:"store" bson:"store"`
	VariantID     string      `json:"variantID" bson:"variantID"`
	DealID        string      `json:"dealID" bson:"dealID"`
	CouponID      string      `json:"couponID" bson:"couponID"`
	Status        cartStatus  `json:"status" bson:"status"`
	Items         ItemList    `json:"items" bson:"items"`
	CartAmount    OrderAmount `json:"cartAmount" bson:"cartAmount"`
	OrderID       string      `json:"orderID" bson:"orderID"`
	ACRWorkflowID string      `json:"ACRWorkflowID" bson:"ACRWorkflowID"`
	RecoveredCart bool        `json:"recoveredCart" bson:"recoveredCart"`
}

func (Cart) GormDataType() string {
	return "json"
}

func (data Cart) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *Cart) Scan(value interface{}) error {

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

type Item struct {
	BrandID    string    `json:"brandID" bson:"brandID"`
	Brand      Brand     `json:"brand" bson:"brand"`
	ProductID  string    `json:"productID" bson:"productID"`
	Product    Product   `json:"product" bson:"product"`
	VariantID  string    `json:"variantID" bson:"variantID"`
	Variant    Variation `json:"variant" bson:"variant"`
	Quantity   int       `json:"quantity" bson:"quantity"`
	Commission float64   `json:"commission" bson:"commission"`
}

type ItemList []Item

func (ItemList) GormDataType() string {
	return "json"
}

func (data ItemList) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *ItemList) Scan(value interface{}) error {

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

func (cart Cart) CreateIndexes() error {
	indexModels := []mongo.IndexModel{}

	userIDModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "userID", Value: 1},
		},
	}
	indexModels = append(indexModels, userIDModel)

	productIDModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "productID", Value: 1},
		},
	}
	indexModels = append(indexModels, productIDModel)

	createdAtModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "createdAt", Value: -1},
		},
	}
	indexModels = append(indexModels, createdAtModel)

	updatedAtModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "updatedAt", Value: -1},
		},
	}
	indexModels = append(indexModels, updatedAtModel)

	statusModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "status", Value: -1},
		},
	}
	indexModels = append(indexModels, statusModel)

	indexName, err := db.CartCollection.Indexes().CreateMany(context.Background(), indexModels)
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

func CreateNewCart2(userID string, userAgentID string, items []Item, source PortalType) (cart Cart, err error) {
	cart.ID = data.GetUUIDString("cart")
	cart.CreatedAt = time.Now()
	cart.Source = source
	cart.UpdatedAt = cart.CreatedAt
	cart.UserAgentID = userAgentID
	cart.Status = Initiated
	cart.UserID = userID
	cart.ACRWorkflowID = fmt.Sprintf("ACR-%s", cart.ID)
	//iterate over items and add them to cart
	for _, item := range items {
		var variants []Variation
		var variant Variation
		item.Product, err = GetProduct(item.ProductID)
		if err != nil {
			fmt.Println("product get err:", err.Error())
			return
		}
		item.BrandID = item.Product.BrandID
		item.Brand, err = GetBrand(item.Product.BrandID)
		if err != nil {
			fmt.Println("brand get err:", err)
		}
		if item.VariantID == "" { //if variant id not sent
			variants, err = GetVariantsByProductID(item.ProductID)
			if err != nil {
				fmt.Println("variants get err:", err.Error())
				return
			}
			// add first variant
			item.VariantID = variants[0].ID
			item.Variant = variants[0]
		} else if item.Variant.ID == "" { // if variant empty but variant id present
			variant, err = GetVariant(item.VariantID)
			if err != nil {
				fmt.Println("variants get err:", err.Error())
				return
			}

			item.Variant = variant
		}
		cart.Items = append(cart.Items, item)
	}
	cart.ProductID = cart.Items[0].ProductID
	cart.VariantID = cart.Items[0].VariantID
	cart.CartAmount = cart.calculateCartAmount() // Use this line

	_, err = db.CartCollection.InsertOne(context.Background(), cart)
	if err != nil {
		fmt.Println(err)
		return
	}

	go Mysql.DB.Model(&cart).Create(&cart)

	// Create ACR Workflow
	TemporalJobs.CreateACRWorkflow(cart.ID, cart.ACRWorkflowID)

	return
}

func CreateNewCart(userAgentID string, product *Product, variant *Variation, brand *Brand, dealID string, userID string, quantity int, source PortalType) (cart Cart, err error) {
	cart.ID = data.GetUUIDString("cart")
	cart.CreatedAt = time.Now()
	cart.Source = source
	cart.UpdatedAt = cart.CreatedAt
	cart.UserAgentID = userAgentID
	cart.ProductID = product.ID
	cart.VariantID = variant.ID
	cart.DealID = dealID
	cart.Status = Initiated
	cart.UserID = userID
	cart.ACRWorkflowID = fmt.Sprintf("ACR-%s", cart.ID)

	firstItem := Item{
		Product:   *product,
		ProductID: product.ID,
		VariantID: variant.ID,
		Quantity:  quantity,
		Variant:   *variant,
	}

	cart.Items = append(cart.Items, firstItem)

	cart.CartAmount = cart.calculateCartAmount() // Use this line

	_, err = db.CartCollection.InsertOne(context.Background(), cart)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create ACR Workflow
	TemporalJobs.CreateACRWorkflow(cart.ID, cart.ACRWorkflowID)

	return
}

// func CreateNewCartFromQuantityInfo(userID string ,quantityInfo map[string]shiprocket.QuantityInfo) (cart *Cart, err error) {

// 	return
// }

func (cart *Cart) calculateCartAmount() OrderAmount {
	cartAmount := ProductPrice{}
	for _, item := range cart.Items {
		cartAmount.Add(item.Variant.Price, item.Quantity)
	}

	cart.CartAmount.ProductPrice = cartAmount
	cart.CartAmount.CalculateTotalPrice()

	return cart.CartAmount
}

func (cart *Cart) CalculateCartAmount() OrderAmount {
	return cart.calculateCartAmount()
}

func GetCart(cartID string) (cart Cart, err error) {

	err = db.CartCollection.FindOne(context.Background(), bson.M{"_id": cartID}).Decode(&cart)
	if err != nil {
		fmt.Println(err)
		return
	}

	return
}

func (cart Cart) OrderCreated(userID string, orderID string, recoveredCart bool) (err error) {

	cart.UpdatedAt = time.Now()
	cart.UserID = userID
	cart.OrderID = orderID
	cart.Status = OrderCreated
	cart.RecoveredCart = recoveredCart
	err = cart.Update()
	if err != nil {
		fmt.Println(err)
		return
	}
	return

}

func (cart Cart) OrderCompleted() (err error) {

	// // If wallet balance is Used, deduct it
	// err = cart.DeductWalletBalance()
	// if err != nil {
	// 	return
	// }

	cart.UpdatedAt = time.Now()
	cart.Status = OrderCompleted
	err = cart.Update()
	if err != nil {
		fmt.Println(err)
		return
	}
	return

}

// func (cart Cart) DeductWalletBalance() (err error) {

// 	wallet, err := GetUserWallet(cart.UserID)
// 	if err != nil {
// 		return err
// 	}

// 	err = wallet.DeductAmount(cart.CartAmount.WalletBalancedUsed)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

func (cart *Cart) ApplyCoupon(coupon Coupon) (applicable bool, reason string, err error) {
	// Check if Coupon is not Expired

	validity, reason1 := coupon.IsValid(cart.UserID)
	if !validity {
		return false, reason1, nil
	}

	for _, item := range cart.Items {
		applicable, reason, err = coupon.IsCouponApplicable(item.Product.ID)
		if err != nil {
			return
		}

		if applicable {
			cart.CouponID = coupon.ID
			cart.CartAmount.Coupon.CouponID = coupon.ID
			cart.CartAmount.Coupon.DiscountAmount = coupon.CalculateDiscount(item.Variant.Price.SellingPrice)
			cart.CartAmount.CalculateTotalPrice()

			err = cart.Update()
			if err != nil {
				return false, "Could not Apply Coupon to Cart", err
			}
			reason = "Coupon Applied"

			return
		}

	}

	return false, "Coupon Does not match Cart", nil

	// applicable, reason, err = coupon.IsCouponApplicable(cart.ProductID)
	// if err != nil {
	// 	return
	// }

	// if applicable {
	// 	cart.CouponID = coupon.ID
	// 	cart.CartAmount.Coupon.CouponID = coupon.ID
	// 	cart.CartAmount.Coupon.DiscountAmount = coupon.CalculateDiscount(cart.CartAmount.ProductPrice.SellingPrice)
	// 	cart.CartAmount.CalculateTotalPrice()

	// 	err = cart.Update()
	// 	if err != nil {
	// 		return false, "Could not Apply Coupon to Cart", err
	// 	}
	// 	reason = "Coupon Applied"
	// }

	// return
}

func (cart *Cart) RemoveCoupon() (err error) {
	cart.CouponID = ""
	cart.CartAmount.Coupon.CouponID = ""
	cart.CartAmount.Coupon.DiscountAmount = 0
	cart.CartAmount.CalculateTotalPrice()
	err = cart.Update()
	return
}

func (cart *Cart) MarkAsRecovered() (err error) {
	cart.RecoveredCart = true
	err = cart.Update()
	return
}

func (cart *Cart) Update() (err error) {
	cart.UpdatedAt = time.Now()
	_, err = db.CartCollection.UpdateOne(context.Background(), bson.M{"_id": cart.ID}, bson.M{"$set": cart})
	if err != nil {
		fmt.Println(err)
		return
	}

	go Mysql.DB.Model(&cart).Save(&cart)

	return
}

func (cart *Cart) ChangeQuantityForanItem(variantID string, newQuantity int) error {
	for i := range cart.Items {
		if cart.Items[i].Variant.ID == variantID {
			cart.Items[i].Quantity = newQuantity
		}
	}

	cart.calculateCartAmount()
	return cart.Update()
}

func AssociateCartWithUser(cartID string, userID string) error {
	cart, err := GetCart(cartID)
	if err != nil {
		return err
	}

	if cart.UserID != "" {
		return nil
	}

	cart.UserID = userID
	err = cart.Update()
	return err
}

func (cart Cart) WasSimilarCartWasCompleted() bool {

	filter := bson.D{
		{Key: "userID", Value: cart.UserID},                      // Cart was of this User
		{Key: "status", Value: 2},                                // Cart was completed
		{Key: "createdAt", Value: bson.M{"$gt": cart.CreatedAt}}, // Cart was created after this cart
		// {Key: "items.productID", Value: cart.Items[0].ProductID},  // Cart has the same product
	}
	options := options.FindOneOptions{
		Sort: bson.M{"createdAt": -1},
	}

	err := db.CartCollection.FindOne(context.Background(), filter, &options).Err()

	return err != nil // If error is receivied

}

func (cart Cart) KillACRWorkflow(reason string) (err error) {
	return TemporalJobs.DeleteTemporalWorkflow2(cart.ACRWorkflowID, reason)
}

func (cart Cart) ACRAllowed() (bool, string) {

	// Get User from DB
	user, err := GetUserByID(cart.UserID)
	if err != nil {
		return false, err.Error()
	}

	// if user.WasACartCompletedInLastXDays(1) {
	// 	return false, "ACR Cooldown Period"
	// }

	if !user.LastAbandonedCart(cart.ID) {
		return false, "not Last Abondoned Cart"
	}

	order, err := user.GetLastCompletedOrder()
	if err != nil {
		return true, ""
	}

	if time.Since(order.CreatedAt) < 24*time.Hour {
		return false, "ACR Cooldown period"
	}

	orderedProduct := order.Cart.Items[0].ProductID

	productInThisCart := cart.Items[0].ProductID

	if orderedProduct != productInThisCart {
		return true, ""
	}

	if order.ShipmentStatus != "DELIVERED" {
		return false, "Previous Order with this Product not yet delivered"
	}

	return true, ""
}

// func (cart *Cart) UseWalletBalance() error {

// 	wallet, err := GetUserWallet(cart.UserID)
// 	if err != nil {
// 		return err
// 	}

// 	cart.CartAmount.WalletBalancedUsed = math.Min(cart.CartAmount.TotalAmount, wallet.Balance)
// 	cart.calculateCartAmount()

// 	return nil
// }

// func (cart *Cart) RemoveWalletBalance() error {

// 	cart.CartAmount.WalletBalancedUsed = 0
// 	cart.calculateCartAmount()

// 	// Remove Lock on Wallet Balance
// 	return nil
// }
