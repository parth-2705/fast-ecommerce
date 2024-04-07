package shiprocket

import (
	"context"
	"encoding/json"
	"fmt"
	"hermes/db"
	"hermes/models"
	"hermes/services/Sentry"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tryamigo/themis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DistinctKeys struct {
	ID       string `json:"_id" bson:"_id"`
	OrderIDs []int  `json:"orderIDs" bson:"orderIDs"`
}

type Scan struct {
	Location      string `json:"location,omitempty"`
	Date          string `json:"date,omitempty"`
	Activity      string `json:"activity,omitempty"`
	Status        string `json:"status,omitempty"`
	SrStatusLabel string `json:"sr-status-label,omitempty"`
}

type BreakPoint struct {
	Completed bool   `json:"completed,omitempty"`
	Status    string `json:"status,omitempty"`
	Date      string `json:"date,omitempty"`
	Activity  string `json:"activity,omitempty"`
}

var completeMap = []string{
	"ARRIVED",
	"OUT FOR DELIVERY",
	"IN TRANSIT",
	"SHIPPED",
	"ORDERED",
}

func TrackingEventReceived(c *gin.Context) {

	c.Writer.Header().Set("Content-Type", "application/json")
	var eventsDump map[string]interface{}
	if err := c.ShouldBind(&eventsDump); err != nil {
		Sentry.SendErrorToSentry(c, err, nil)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	_, err := db.EventsDumpCollection.InsertOne(context.Background(), eventsDump)
	if err != nil {
		Sentry.SendErrorToSentry(c, err, nil)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	marshalledEventsDump, err := json.Marshal(&eventsDump)
	if err != nil {
		Sentry.SendErrorToSentry(c, err, nil)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var trackingEvent models.ShipRocketTracking

	err = json.Unmarshal(marshalledEventsDump, &trackingEvent)
	if err != nil {
		Sentry.SendErrorToSentry(c, err, nil)
		Sentry.SentryCaptureException(err)
		c.JSON(http.StatusOK, gin.H{
			"error": err.Error(),
		})
		return
	}

	opts := options.Update().SetUpsert(true)
	_, err = db.TrackingCollection.UpdateOne(context.Background(), bson.M{"awb": trackingEvent.Awb}, bson.M{"$set": trackingEvent}, opts)
	if err != nil {
		Sentry.SendErrorToSentry(c, err, nil)
		c.JSON(http.StatusOK, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = trackingEvent.MarkOrderStatusForThisEvent()
	if err != nil {
		Sentry.SendErrorToSentry(c, err, nil)
		c.JSON(http.StatusOK, gin.H{
			"error": err.Error(),
		})
		return
	}

	go UpdateShiprocketOrderUtil(trackingEvent.SrOrderID)

	c.JSON(http.StatusOK, gin.H{
		"response": "Success",
	})
}

func GetTrackingForAWB(awb string) (tracking models.ShipRocketTracking, err error) {
	err = db.TrackingCollection.FindOne(context.Background(), bson.M{"awb": awb}).Decode(&tracking)
	return
}

func remove(slice []BreakPoint, s int) []BreakPoint {
	return append(slice[:s], slice[s+1:]...)
}

func MakeTrackingData(tracking models.ShipRocketTracking, createdAt time.Time) (breakpoints []BreakPoint, err error) {

	breakpoints = []BreakPoint{
		{Completed: false, Status: "ARRIVING", Date: "", Activity: ""},
		{Completed: false, Status: "OUT FOR DELIVERY", Date: "", Activity: ""},
		{Completed: false, Status: "IN TRANSIT", Date: "", Activity: ""},
		{Completed: false, Status: "SHIPPING", Date: "", Activity: ""},
		{Completed: true, Status: "ORDERED", Date: createdAt.Format("2006-01-02 15:04:05"), Activity: ""},
	}

	if tracking.CurrentStatus == "DELIVERED" {
		MarkBreakpointAsComplete(breakpoints, 3)
		MarkBreakpointAsComplete(breakpoints, 2)
		MarkBreakpointAsComplete(breakpoints, 1)
		MarkBreakpointAsComplete(breakpoints, 0)
	}

	for _, val := range tracking.Scans {
		if val.SrStatusLabel == "OUT FOR DELIVERY" {
			MarkBreakpointAsComplete(breakpoints, 1)
			MarkBreakpointAsComplete(breakpoints, 2)
			MarkBreakpointAsComplete(breakpoints, 3)
			breakpoints[1].Date = val.Date
			breakpoints[1].Activity = val.Activity
		} else if val.SrStatusLabel == "IN TRANSIT" {
			MarkBreakpointAsComplete(breakpoints, 2)
			MarkBreakpointAsComplete(breakpoints, 3)
			breakpoints[2].Activity = val.Activity
			breakpoints[2].Date = val.Date
		} else if val.SrStatusLabel == "SHIPPED" {
			MarkBreakpointAsComplete(breakpoints, 3)
			breakpoints[3].Date = val.Date
			breakpoints[3].Activity = val.Activity
		} else if val.SrStatusLabel == "RTO INITIATED" && val.Status != "created" {
			breakpoints[0].Status = "RETURNED"
			breakpoints[0].Completed = true

			if !breakpoints[3].Completed {
				breakpoints = remove(breakpoints, 3)
			}
			if !breakpoints[2].Completed {
				breakpoints = remove(breakpoints, 2)
			}
			if !breakpoints[1].Completed {
				breakpoints = remove(breakpoints, 1)
			}

		} else if val.SrStatusLabel == "DELIVERED" {
			MarkBreakpointAsComplete(breakpoints, 3)
			MarkBreakpointAsComplete(breakpoints, 2)
			MarkBreakpointAsComplete(breakpoints, 1)
			MarkBreakpointAsComplete(breakpoints, 0)
			breakpoints[0].Activity = val.Activity
			breakpoints[0].Date = val.Date

		}
	}
	return
}

func MarkBreakpointAsComplete(breakpoints []BreakPoint, index int) []BreakPoint {
	temp := breakpoints
	if len(temp) <= index {
		return temp
	}
	temp[index].Completed = true
	temp[index].Status = completeMap[index]
	return temp
}

func UpdateShiprocketOrderUtil(shiprocketOrderID int) {
	err := UpdateShiprocketOrder(shiprocketOrderID)
	if err != nil {
		Sentry.SentryCaptureException(err)
	}
}

func UpdateShiprocketOrder(shiprocketOrderID int) (err error) {
	var shipping models.Shipping
	var temp map[string]models.FullShiprocketOrder
	var tempData models.FullShiprocketOrder

	err = db.ShippingCollection.FindOne(context.Background(), bson.M{"orderId": shiprocketOrderID}).Decode(&shipping)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			Sentry.SentryCaptureException(fmt.Errorf("could not find order by orderId: %d", shiprocketOrderID))
			return nil
		} else {
			Sentry.SentryCaptureException(fmt.Errorf("error in finding shipping by order id: %s", err.Error()))
			return
		}
	}

	token, err := GetShipRocketToken()
	if err != nil {
		Sentry.SentryCaptureException(fmt.Errorf("error in getting shiprocket token: %s", err.Error()))
		return err
	}

	headers := [][]string{{"Authorization", "Bearer " + token}}
	url := "https://apiv2.shiprocket.in/v1/external/orders/show/" + fmt.Sprint(shipping.OrderId)
	statusCode, responseBody, err := themis.HitAPIEndpoint2(url, http.MethodGet, []byte{}, headers, [][]string{})
	if err != nil {
		return err
	}
	if statusCode >= 400 {
		return fmt.Errorf("error: %s with status code: %d for orderID: %d", string(responseBody), statusCode, shipping.OrderId)
	}

	err = json.Unmarshal(responseBody, &temp)
	if err != nil {
		return fmt.Errorf("Error in unmarshalling "+err.Error()+" for object:%s", string(responseBody))
	}
	tempData = temp["data"]
	tempData.ID = fmt.Sprint(shipping.OrderId)
	if tempData.DeliveredDate == "" {
		tempData.DeliveredDateTimeStamp, _ = time.Parse("2006-01-02", "9999-01-01")
	} else {
		tempData.DeliveredDateTimeStamp, err = time.Parse("02-01-2006 15:04:05", tempData.DeliveredDate)
	}
	if err != nil {
		return err
	}

	if tempData.Shipments.RtoDeliveredDate == "" {
		tempData.RTODeliveredDateTimeStamp, _ = time.Parse("2006-01-02", "9999-01-01")
	} else {
		tempData.RTODeliveredDateTimeStamp, err = time.Parse("2006-01-02 15:04:05", tempData.Shipments.RtoDeliveredDate)
	}
	if err != nil {
		return err
	}
	_, err = db.ShiprocketOrderCollection.UpdateOne(context.Background(), bson.M{"srid": shiprocketOrderID}, bson.M{"$set": tempData}, options.Update().SetUpsert(true))
	if err != nil {
		Sentry.SentryCaptureException(err)
		return err
	}

	if len(tempData.AwbData.Awb) > 0 && tempData.AwbData.Awb != shipping.AWB {
		Sentry.SentryCaptureException(fmt.Errorf("different awb case found previous %s new %s", shipping.AWB, tempData.AwbData.Awb))

		_, err = db.ShippingCollection.UpdateMany(context.Background(), bson.M{"awb": shipping.AWB}, bson.M{"$set": bson.M{
			"awb": tempData.AwbData.Awb,
		}})

		if err != nil {
			Sentry.SentryCaptureException(err)
			return err
		}
	}

	err = models.MigrateShiprocketOrderToMySQL(tempData)
	if err != nil {
		Sentry.SentryCaptureException(err)
		return err
	}
	return
}

func PopulateShiprocketOrders() (err error) {
	var shippings []DistinctKeys
	var shiprocketOrders []interface{}
	token, err := GetShipRocketToken()
	if err != nil {
		return err
	}
	cur, err := db.ShippingCollection.Aggregate(context.Background(), bson.A{
		bson.D{
			{Key: "$group",
				Value: bson.D{
					{Key: "_id", Value: ""},
					{Key: "orderIDs", Value: bson.D{{Key: "$addToSet", Value: "$orderId"}}},
				},
			},
		},
	})
	if err != nil {
		return
	}
	err = cur.All(context.Background(), &shippings)
	if err != nil {
		return
	}

	headers := [][]string{{"Authorization", "Bearer " + token}}
	for idx, shippin := range shippings[0].OrderIDs {
		var temp map[string]models.FullShiprocketOrder
		url := "https://apiv2.shiprocket.in/v1/external/orders/show/" + fmt.Sprint(shippin)
		fmt.Println("id:", idx, shippin)
		statusCode, responseBody, err1 := themis.HitAPIEndpoint2(url, http.MethodGet, []byte{}, headers, [][]string{})
		if err1 != nil {
			return err
		}
		if statusCode >= 400 {
			return fmt.Errorf("erorr: %s with status code: %d", string(responseBody), statusCode)
		}
		json.Unmarshal(responseBody, &temp)
		tempData := temp["data"]
		tempData.ID = fmt.Sprint(shippin)
		if tempData.DeliveredDate == "" {
			tempData.DeliveredDateTimeStamp, _ = time.Parse("2006-01-02", "9999-01-01")
		} else {
			tempData.DeliveredDateTimeStamp, err = time.Parse("02-01-2006 15:04:05", tempData.DeliveredDate)
		}
		if err != nil {
			return err
		}
		if tempData.Shipments.RtoDeliveredDate == "" {
			tempData.RTODeliveredDateTimeStamp, _ = time.Parse("2006-01-02", "9999-01-01")
		} else {
			tempData.RTODeliveredDateTimeStamp, err = time.Parse("2006-01-02 15:04:05", tempData.Shipments.RtoDeliveredDate)
		}
		if err != nil {
			return err
		}
		shiprocketOrders = append(shiprocketOrders, tempData)
		time.Sleep(75 * time.Millisecond)
	}
	_, err = db.ShiprocketOrderCollection.InsertMany(context.Background(), shiprocketOrders)
	return
}

func AddRTODeliveredTimestamp() (err error) {
	var srOrders []models.FullShiprocketOrder
	cur, err := db.ShiprocketOrderCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return
	}
	err = cur.All(context.Background(), &srOrders)
	if err != nil {
		return
	}
	for _, val := range srOrders {
		fmt.Println("id: ", val.ID)
		if val.Shipments.RtoDeliveredDate == "" {
			val.RTODeliveredDateTimeStamp, _ = time.Parse("2006-01-02", "9999-01-01")
		} else {
			val.RTODeliveredDateTimeStamp, err = time.Parse("2006-01-02 15:04:05", val.Shipments.RtoDeliveredDate)
		}
		if err != nil {
			return err
		}
		_, err = db.ShiprocketOrderCollection.ReplaceOne(context.Background(), bson.M{"_id": val.ID}, val)
		if err != nil {
			return
		}
	}
	return
}
