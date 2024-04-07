package models

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"hermes/configs/Mysql"
	"hermes/db"
	"hermes/services/Temporal/TemporalJobs"
	"hermes/utils/whatsapp"
	"reflect"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm/clause"
)

type Shipping struct {
	Id                 string    `json:"id" bson:"_id"`
	CreatedAt          time.Time `json:"createdAt" bson:"createdAt" gorm:"column:createdAt"`
	UpdatedAt          time.Time `json:"updatedAt" bson:"updatedAt" gorm:"column:updatedAt"`
	AWB                string    `json:"awb" bson:"awb"`
	OrderId            int       `json:"orderId" bson:"orderId"`
	ShippingId         int       `json:"shippingId" bson:"shippingId"`
	LabelURL           string    `json:"labelUrl,omitempty" bson:"labelUrl"`
	InvoiceURL         string    `json:"invoiceUrl,omitempty" bson:"invoiceUrl"`
	ManifestURL        string    `json:"manifestUrl,omitempty" bson:"manifestUrl"`
	BrandID            string    `json:"brandID" bson:"brandID"`
	SellerID           string    `json:"sellerID" bson:"sellerID"`
	ShipmentCreated    bool      `json:"shipmentCreated" bson:"shipmentCreated"`
	ParentOrderID      string    `json:"parentOrderID" bson:"_parentOrderID"`
	Dispatched         bool      `json:"dispatched" bson:"dispatched"`
	Processed          bool      `json:"processed" bson:"processed"`
	Remitted           bool      `json:"remitted" bson:"remitted"`
	ParentRelation     int       `json:"relation" bson:"relation"`
	LabelDownloaded    bool      `json:"labelDownloaded" bson:"labelDownloaded"`
	InvoiceDownaded    bool      `json:"invoiceDownloaded" bson:"invoiceDownloaded"`
	ManifestDownloaded bool      `json:"manifestDownloaded" bson:"manifestDownloaded"`

	NDRWorkflowLastTriggeredOn *time.Time `json:"nDRWorkflowLastTriggeredOn" bson:"nDRWorkflowLastTriggeredOn" gorm:"column:nDRWorkflowLastTriggeredOn"`
	NDRWorkflowTriggeredTimes  TimeSlice  `json:"nDRWorkflowTriggeredTimes" bson:"nDRWorkflowTriggeredTimes" gorm:"column:nDRWorkflowTriggeredTimes"`

	State ShipmentStates `json:"state" bson:"state"`
}

type ShipmentStates struct {
	InTransit      ShipmentStateMilestone `json:"inTransit" bson:"inTransit"`
	Shipped        ShipmentStateMilestone `json:"shipped" bson:"shipped"`
	OutForDelivery ShipmentStateMilestone `json:"outForDelivery" bson:"outForDelivery"`
	Delivered      ShipmentStateMilestone `json:"delivered" bson:"delivered"`
}

type ShipmentStateMilestone struct {
	StateAchieved bool       `json:"stateAchieved" bson:"stateAchieved"`
	TimeStamp     *time.Time `json:"timestamp" bson:"timestamp"`
}

type TimeSlice []time.Time

func (TimeSlice) GormDataType() string {
	return "json"
}

func (data TimeSlice) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *TimeSlice) Scan(value interface{}) error {
	err := GormScan(reflect.ValueOf(data), value)
	if err != nil {
		return err
	}
	return nil
}

func (ShipmentStates) GormDataType() string {
	return "json"
}

func (data ShipmentStates) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}
func (data *ShipmentStates) Scan(value interface{}) error {

	err := GormScan(reflect.ValueOf(data), value)
	if err != nil {
		return err
	}
	return nil
}

type BackwardShipment struct {
	ID             string             `json:"_id" bson:"_id"` //id of the parent shipment
	ChildShipments ChildShipmentSlice `json:"childShipments" bson:"childShipments"`
}

type ChildShipmentSlice []ChildShipment

type ChildShipment struct {
	Type        int              `json:"type" bson:"type"`
	ShipmentID  string           `json:"shipmentID" bson:"shipmentID"`
	ProductInfo ProductInfoSlice `json:"productInfo" bson:"productInfo"`
	Status      string           `json:"status" bson:"status"`
}

type ProductInfoSlice []ProductInfoItem
type ProductInfoItem struct {
	ProductID string    `json:"productID" bson:"productID"`
	VariantID string    `json:"variantID" bson:"variantID"`
	Product   Product   `json:"product" bson:"product"`
	Variant   Variation `json:"variant" bson:"variant"`
	SKU       string    `json:"sku" bson:"sku"`
	Quantity  int       `json:"quantity" bson:"quantity"`
}

type CompleteBackwardShipment struct {
	ID            string             `json:"_id" bson:"_id"`
	ChildShipment ChildShipment      `json:"childShipments" bson:"childShipments"`
	Shipping      Shipping           `json:"shipping" bson:"shipping"`
	ParentOrder   Order              `json:"parentOrder" bson:"parentOrder"`
	Tracking      ShipRocketTracking `json:"tracking" bson:"tracking"`
}

type RemittableResp struct {
	ID            string   `json:"_id" bson:"_id"`
	RemittableIDs []string `json:"remittable" bson:"remittable"`
}

// type ParentStruct map[string]int

func (ChildShipmentSlice) GormDataType() string {
	return "json"
}

func (data ChildShipmentSlice) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *ChildShipmentSlice) Scan(value interface{}) error {
	err := GormScan(reflect.ValueOf(data), value)
	if err != nil {
		return err
	}
	return nil
}

func (ProductInfoSlice) GormDataType() string {
	return "json"
}

func (data ProductInfoSlice) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *ProductInfoSlice) Scan(value interface{}) error {
	err := GormScan(reflect.ValueOf(data), value)
	if err != nil {
		return err
	}
	return nil
}

const (
	Parent      int = 0
	Return      int = 1
	Replacement int = 2
	Missing     int = 3
	Exchange    int = 4
)

func GetAllBackwardShipments() (backShips []BackwardShipment, err error) {
	cur, err := db.BackwardShipmentCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return
	}
	err = cur.All(context.Background(), &backShips)
	return
}

func GetReturnShipmentByChildID(childID string) (backShip BackwardShipment, err error) {
	err = db.BackwardShipmentCollection.FindOne(context.Background(), bson.M{"childShipments.shipmentID": childID}).Decode(&backShip)
	return
}

func MarkStatusForBackwardShipment(ID string, status string) (err error) {
	var backShip BackwardShipment
	err = db.BackwardShipmentCollection.FindOne(context.Background(), bson.M{"childShipments.shipmentID": ID}).Decode(&backShip)
	if err != nil {
		return
	}
	for idx, val := range backShip.ChildShipments {
		if val.ShipmentID == ID {
			temp := val
			temp.Status = status
			backShip.ChildShipments[idx] = temp
			err = backShip.Update()
			if err != nil {
				return
			}
		}
	}
	return
}

func GetCompleteBackwardShipments(sellerID string) (backShips []CompleteBackwardShipment, err error) {
	prePipe := bson.A{
		bson.D{
			{"$unwind",
				bson.D{
					{"path", "$childShipments"},
					{"preserveNullAndEmptyArrays", false},
				},
			},
		},
		bson.D{
			{"$lookup",
				bson.D{
					{"from", "shipping"},
					{"localField", "childShipments.shipmentID"},
					{"foreignField", "_id"},
					{"as", "shipping"},
				},
			},
		},
		bson.D{
			{"$unwind",
				bson.D{
					{"path", "$shipping"},
					{"preserveNullAndEmptyArrays", false},
				},
			},
		}}

	postPipe := bson.A{
		bson.D{
			{"$lookup",
				bson.D{
					{"from", "orders"},
					{"localField", "_id"},
					{"foreignField", "_id"},
					{"as", "parentOrder"},
				},
			},
		},
		bson.D{
			{"$unwind",
				bson.D{
					{"path", "$parentOrder"},
					{"preserveNullAndEmptyArrays", false},
				},
			},
		},
		bson.D{
			{"$lookup",
				bson.D{
					{"from", "tracking"},
					{"localField", "shipping.awb"},
					{"foreignField", "awb"},
					{"as", "tracking"},
				},
			},
		},
		bson.D{
			{"$unwind",
				bson.D{
					{"path", "$tracking"},
					{"preserveNullAndEmptyArrays", true},
				},
			},
		},
	}
	temp := bson.A{}
	if sellerID != "" {
		temp = append(temp, bson.D{{Key: "$match", Value: bson.M{"shipping.sellerID": sellerID}}})
	}
	pipeline := append(prePipe, temp...)
	pipeline = append(pipeline, postPipe...)
	cur, err := db.BackwardShipmentCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return
	}
	err = cur.All(context.Background(), &backShips)
	return
}

func (shipping Shipping) CreateIndexes() error {
	awbModel := mongo.IndexModel{
		Keys: bson.M{
			"awb": 1, // 1 for ascending, -1 for descending
		},
	}
	_, err := db.ShippingCollection.Indexes().CreateOne(context.Background(), awbModel)
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

func (backShip BackwardShipment) Create() (err error) {
	_, err = db.BackwardShipmentCollection.InsertOne(context.Background(), backShip)
	return
}

func (backShip BackwardShipment) Update() (err error) {
	_, err = db.BackwardShipmentCollection.UpdateOne(context.Background(), bson.M{"_id": backShip.ID}, bson.M{"$set": backShip})
	return
}

func (shipping Shipping) Insert() error {
	shipping.CreatedAt = time.Now()
	shipping.UpdatedAt = time.Now()

	_, err := db.ShippingCollection.InsertOne(context.Background(), shipping)
	if err != nil {
		return err
	}

	err = MigrateShippingToMySQLLogs(shipping)
	if err != nil {
		return err
	}

	return nil
}

func (shipping Shipping) TriggerNDRHandling() error {

	if len(shipping.AWB) == 0 {
		return fmt.Errorf("no awb present")
	}

	if shipping.NDRWorkflowLastTriggeredOn != nil {

		// Find the duration since last triggered NDR message
		duration := time.Since(*shipping.NDRWorkflowLastTriggeredOn)

		// Check if the duration is less than 24 hours
		if duration < 24*time.Hour {
			return nil
		}
	}

	order, err := GetOrder(shipping.Id)
	if err != nil {
		return fmt.Errorf("error while fetching order: " + err.Error())
	}

	if len(order.Cart.Items) == 0 {
		return fmt.Errorf("no cart items found")
	}

	var paymentMethod string
	if order.Payment.Method == "COD" {
		paymentMethod = "Cash on Delivery"
	} else {
		paymentMethod = "Prepaid"
	}

	// trigger the NDR workflow with AWB
	err = whatsapp.SendNDRMessage(order.Address.Name, order.Cart.Items[0].Product.Name, paymentMethod, order.Address.Phone, shipping.AWB)
	if err != nil {
		return err
	}

	timeNow := time.Now()
	shipping.NDRWorkflowLastTriggeredOn = &timeNow
	shipping.NDRWorkflowTriggeredTimes = append(shipping.NDRWorkflowTriggeredTimes, time.Now())

	_, err = db.ShippingCollection.UpdateMany(context.Background(), bson.M{"awb": shipping.AWB}, bson.M{"$set": bson.M{"nDRWorkflowLastTriggeredOn": shipping.NDRWorkflowLastTriggeredOn, "nDRWorkflowTriggeredTimes": shipping.NDRWorkflowTriggeredTimes, "updatedAt": time.Now()}})
	if err != nil {
		return err
	}

	return nil
}

func (shipping Shipping) UpdateFields(toUpdateFields bson.M) error {
	_, err := db.ShippingCollection.UpdateMany(context.Background(), bson.M{"shippingId": shipping.ShippingId}, bson.M{"$set": toUpdateFields})
	if err != nil {
		return err
	}

	return nil
}

func (shipping Shipping) UpdateFieldsByOrderID(toUpdateFields bson.M) error {
	_, err := db.ShippingCollection.UpdateMany(context.Background(), bson.M{"_id": shipping.Id}, bson.M{"$set": toUpdateFields})
	if err != nil {
		return err
	}

	return nil
}

func GetShipmentByID(id string) (Shipping, error) {
	var shipping Shipping
	err := db.ShippingCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&shipping)
	if err != nil {
		return shipping, err
	}
	return shipping, nil
}

func GetShipmentByShipmentID(shipmentID int) (Shipping, error) {
	var shipping Shipping
	err := db.ShippingCollection.FindOne(context.Background(), bson.M{"shippingId": shipmentID}).Decode(&shipping)
	if err != nil {
		return shipping, err
	}
	return shipping, nil
}

// Get first Shipment with given AWB
func GetShipmentByAWB(awb string) (shipments Shipping, err error) {
	err = db.ShippingCollection.FindOne(context.Background(), bson.M{"awb": awb}).Decode(&shipments)
	if err != nil {
		fmt.Println(err)
		return
	}

	return
}

// Get all Shipments with given AWB
func GetShipmentsByAWB(awb string) (shipments []Shipping, err error) {
	cursor, err := db.ShippingCollection.Find(context.Background(), bson.M{"awb": awb})
	if err != nil {
		fmt.Println(err)
		return
	}

	err = cursor.All(context.Background(), &shipments)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	return
}

func MarkAsTrue(orderID string, key string) (err error) {
	var shipping Shipping
	shipping.Id = orderID

	err = shipping.UpdateFieldsByOrderID(bson.M{key: true})
	return
}

func MarkAsDispatched(orderID string) (err error) {
	var shipping Shipping
	shipping.Id = orderID

	err = shipping.UpdateFieldsByOrderID(bson.M{"dispatched": true})
	return
}

func (shipping Shipping) MarkDispatched() (err error) {
	err = shipping.UpdateFieldsByOrderID(bson.M{"dispatched": true, "shipmentCreated": true})
	return
}

func UnmarkDispatch(orderID string) (err error) {
	var shipping Shipping
	shipping.Id = orderID

	err = shipping.UpdateFieldsByOrderID(bson.M{"dispatched": false})
	return
}

func MarkAsProcessed(orderID string) (err error) {
	var shipping Shipping
	shipping.Id = orderID
	err = shipping.UpdateFieldsByOrderID(bson.M{"processed": true})
	return
}

func UnmarkProcessed(orderID string) (err error) {
	var shipping Shipping
	shipping.Id = orderID
	err = shipping.UpdateFieldsByOrderID(bson.M{"processed": false, "dispatched": false})
	return
}

func MigrateShippingToMySQLLogs(shipping Shipping) (err error) {
	err = Mysql.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&shipping).Error

	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func (shipment *Shipping) MarkAsOFD(orderShipmentStatus string) error {
	order, err := GetOrder(shipment.Id)
	if err != nil {
		return err
	}

	if shipment.State.OutForDelivery.StateAchieved && todaysDate(shipment.State.OutForDelivery.TimeStamp) {
		return nil
	}

	currTime := time.Now()
	shipment.State.OutForDelivery = ShipmentStateMilestone{
		StateAchieved: true,
		TimeStamp:     &currTime,
	}

	err = order.MarkOrderAsOutForDelivery(orderShipmentStatus)
	if err != nil {
		return err
	}

	return shipment.Update()
}

func (shipment *Shipping) TriggerDeliveryWorkflows(order Order) error {
	// Workflow to add 5% of the order value to the referring user wallet and add 1% of the order value to the ordering user wallet
	err := TemporalJobs.CreateRoovoPeCreditWorkflow(order.ID, order.User.Internal)
	if err != nil {
		return err
	}

	return nil
}

func todaysDate(ts *time.Time) bool {
	if ts == nil {
		return false
	}

	return ts.Day() == time.Now().Day()
}

func (shipment *Shipping) MarkAsInTransit(etd string, orderShipmentStatus string) error {
	order, err := GetOrder(shipment.Id)
	if err != nil {
		return err
	}

	if shipment.State.InTransit.StateAchieved {
		return nil
	}

	currTime := time.Now()
	shipment.State.InTransit = ShipmentStateMilestone{
		StateAchieved: true,
		TimeStamp:     &currTime,
	}

	err = order.MarkOrderAsInTransit(etd, orderShipmentStatus)
	if err != nil {
		return err
	}

	return shipment.Update()
}

func (shipment *Shipping) Update() (err error) {
	shipment.UpdatedAt = time.Now()
	_, err = db.ShippingCollection.ReplaceOne(context.Background(), bson.M{"_id": shipment.Id}, shipment)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}

	return
}

func MarkShipmentsAsRemitted(IDs []string) (err error) {
	_, err = db.ShippingCollection.UpdateMany(context.Background(), bson.M{"_id": bson.M{"$in": IDs}}, bson.M{"$set": bson.M{"remitted": true}})
	return
}

// func GetRemittableShipmentIDsForDateRange(sellerID string, startDate string, endDate string) (IDs []string, err error) {

// 	loc, _ := time.LoadLocation("Asia/Calcutta")
// 	startDateTime, _ := time.ParseInLocation("2006-01-02", startDate, loc)
// 	endDateTime, _ := time.ParseInLocation("2006-01-02", endDate, loc)

// 	pipeline := utils.RemittableShipmentsQuery(sellerID, startDateTime, endDateTime)
// 	cur, err := db.ShippingCollection.Aggregate(context.Background(), pipeline)
// 	if err != nil {
// 		return
// 	}
// 	var temp []RemittableResp
// 	err = cur.All(context.Background(), &temp)
// 	if err != nil {
// 		return
// 	}
// 	if len(temp) == 0 {
// 		return []string{}, fmt.Errorf("no remittable orders")
// 	}
// 	return temp[0].RemittableIDs, nil
// }

func OrderHasReturnOrder(order Order) (bool, error) {
	var shipping []Shipping

	pipeline := bson.A{
		bson.D{
			{Key: "$match",
				Value: bson.D{
					{Key: "relation", Value: 1},
					{Key: "$expr",
						Value: bson.D{
							{Key: "$not",
								Value: bson.D{
									{Key: "$eq",
										Value: bson.A{
											"$_id",
											"$_parentOrderID",
										},
									},
								},
							},
						},
					},
					{Key: "$expr",
						Value: bson.D{
							{Key: "$eq",
								Value: bson.A{
									order.ID,
									"$_parentOrderID",
								},
							},
						},
					},
				},
			},
		},
	}

	curr, err := db.ShippingCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return true, err
	}

	err = curr.All(context.Background(), &shipping)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return true, err
	}

	if len(shipping) == 0 {
		return false, nil
	}

	return true, nil
}
