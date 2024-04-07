package models

import (
	"context"
	"fmt"
	"hermes/db"
	"strings"

	"github.com/getsentry/sentry-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ShipRocketTracking struct {
	Awb              string `json:"awb" bson:"awb"`
	CourierName      string `json:"courier_name" bson:"couriername"`
	CurrentStatus    string `json:"current_status" bson:"currentstatus"`
	CurrentStatusID  any    `json:"current_status_id" bson:"currentstatusid"`
	ShipmentStatus   string `json:"shipment_status" bson:"shipmentstatus"`
	ShipmentStatusID any    `json:"shipment_status_id" bson:"shipmentstatusid"`
	CurrentTimestamp string `json:"current_timestamp" bson:"currenttimestamp"`
	OrderID          string `json:"order_id" bson:"orderid"`
	SrOrderID        int    `json:"sr_order_id" bson:"srorderid"`
	Etd              string `json:"etd" bson:"etd"`
	Scans            []struct {
		Location      string `json:"location" bson:"location"`
		Date          string `json:"date" bson:"date"`
		Activity      string `json:"activity" bson:"activity"`
		Status        string `json:"status" bson:"status"`
		SrStatusLabel string `json:"sr-status-label" bson:"srstatuslabel"`
	} `json:"scans" bson:"scans"`
	IsReturn  any `json:"is_return" bson:"isreturn"`
	ChannelID any `json:"channel_id" bson:"channelid"`
}

func (ShipRocketTracking ShipRocketTracking) CreateIndexes() error {
	awbModel := mongo.IndexModel{
		Keys: bson.M{
			"awb": 1, // 1 for ascending, -1 for descending
		},
	}
	_, err := db.TrackingCollection.Indexes().CreateOne(context.Background(), awbModel)
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

func (event ShipRocketTracking) MarkOrderStatusForThisEvent() (err error) {

	// Find Shipments
	shipments, err := GetShipmentsByAWB(event.Awb)
	if err != nil {
		return err
	}

	// Iterate Through Shipment
	for _, shipment := range shipments {
		// Get Order for Each Shipment
		if shipment.ParentRelation != 0 {
			err = MarkStatusForBackwardShipment(shipment.Id, event.ShipmentStatus)
		} else {
			order, err := GetOrder(shipment.Id)
			if err != nil {
				return err
			}

			// Mark Order Shipment Status According to Shiprocket Shipment Status Codes : https://apidocs.shiprocket.in/#0f9a75fd-6d23-453c-a3d7-85857e8c8759
			// Folks at Shiprocket are very smart, they randomly send shipment Status as Integer or String, hence a check on type is required
			switch val := event.ShipmentStatusID.(type) {
			case string:
				switch val {
				case "6":
					err = order.MarkOrderAsShipped(event.Etd, event.ShipmentStatus)
				case "7":
					err = order.MarkOrderDelivered2(event.ShipmentStatus, event.CurrentTimestamp)
				case "19":
					err = order.MarkOrderAsOutForPickup(event.ShipmentStatus)
				case "42":
					err = order.MarkOrderAsPickedUp(event.Etd, event.ShipmentStatus)
				default:
					err = order.SetShipmentStatus(event.ShipmentStatus)
				}
			case float64:
				switch val {
				case 6:
					err = order.MarkOrderAsShipped(event.Etd, event.ShipmentStatus)
				case 7:
					err = order.MarkOrderDelivered2(event.ShipmentStatus, event.CurrentTimestamp)
				case 19:
					err = order.MarkOrderAsOutForPickup(event.ShipmentStatus)
				case 42:
					err = order.MarkOrderAsPickedUp(event.Etd, event.ShipmentStatus)
				default:
					err = order.SetShipmentStatus(event.ShipmentStatus)
				}
			default:
				sentry.CaptureException(fmt.Errorf("could not match type StatusID for Tracking Event: %s", shipment.Id))
				err = order.SetShipmentStatus(event.ShipmentStatus)
			}
			if err != nil {
				return err
			}

			// Mark shipment as dispatched
			switch val := event.ShipmentStatusID.(type) {
			case string:
				switch val {
				case "7":
					err = shipment.TriggerDeliveryWorkflows(order)
				case "17":
					err = shipment.MarkAsOFD(event.ShipmentStatus)
				case "18":
					err = shipment.MarkAsInTransit(event.Etd, event.ShipmentStatus)
				case "42":
					err = shipment.MarkDispatched()
				case "21":
					err = shipment.TriggerNDRHandling()
				}
			case float64:
				switch val {
				case 7:
					err = shipment.TriggerDeliveryWorkflows(order)
				case 17:
					err = shipment.MarkAsOFD(event.ShipmentStatus)
				case 18:
					err = shipment.MarkAsInTransit(event.Etd, event.ShipmentStatus)
				case 42:
					err = shipment.MarkDispatched()
				case 21:
					err = shipment.TriggerNDRHandling()
				}
			default:
				sentry.CaptureException(fmt.Errorf("could not match type StatusID for Tracking Event: %s", shipment.Id))
				err = order.SetShipmentStatus(event.ShipmentStatus)
			}

			if err != nil {
				return err
			}
		}
		if err != nil {
			return err
		}
	}

	return
}
