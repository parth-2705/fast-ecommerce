package scripts

import (
	"context"
	"encoding/base64"
	"fmt"
	"hermes/db"
	"hermes/models"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func encodeDescriptionsForProduct() error {
	var products []models.Product
	cur, err := db.ProductCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}
	err = cur.All(context.Background(), &products)
	if err != nil {
		return err
	}
	for _, val := range products {
		temp := val
		temp.DescriptionEncoded = base64.StdEncoding.EncodeToString([]byte(val.Description))
		temp.Name = strings.TrimSpace(val.Name)
		_, err = db.ProductCollection.ReplaceOne(context.Background(), bson.M{"_id": temp.ID}, temp)
		if err != nil {
			return err
		}
	}
	return nil
}

func AddDeliveryTimeToProducts() (err error) {
	var orders []models.Order
	var updateMap map[string]struct{} = make(map[string]struct{})
	cur, err := db.OrderCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return
	}
	err = cur.All(context.Background(), &orders)
	if err != nil {
		return
	}
	for _, val := range orders {
		var temp models.ShipRocketTracking
		if val.ShipmentStatus == "DELIVERED" || val.ShipmentStatus == "Delivered" {
			fmt.Println("id:", val.ID)
			shipment, err := models.GetShipmentByID(val.ID)
			if err == mongo.ErrNoDocuments {
				continue
			}
			if err != nil {
				return err
			}
			err = db.TrackingCollection.FindOne(context.Background(), bson.M{"awb": shipment.AWB}).Decode(&temp)
			if err == mongo.ErrNoDocuments {
				continue
			}
			if err != nil && err != mongo.ErrNoDocuments {
				return err
			}
			err = models.UpdateAverageDeliveryTimeByOrder(val, temp.CurrentTimestamp)
			if err == mongo.ErrNoDocuments {
				continue
			}
			if err != nil {
				return err
			}
			for _, cartItem := range val.Cart.Items {
				updateMap[cartItem.ProductID] = struct{}{}
			}
		}
	}

	var products []models.Product
	cur, err = db.ProductCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return
	}
	err = cur.All(context.Background(), &products)
	if err != nil {
		return
	}
	for _, val := range products {
		if _, ok := updateMap[val.ID]; !ok {
			fmt.Println("product id:", val.ID)
			_, err = db.ProductCollection.UpdateOne(context.Background(), bson.M{"_id": val.ID}, bson.M{"$set": bson.M{"deliveryTime": models.DeliveryTime{
				AverageDeliveryTime: 3,
				DeliveryCompleted:   0,
				DeliveryMap:         map[int]int{},
			}}})
			if err != nil {
				return
			}
			updateMap[val.ID] = struct{}{}
		}
	}
	return
}
