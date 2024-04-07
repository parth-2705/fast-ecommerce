package scripts

import (
	"context"
	"hermes/db"
	"hermes/models"

	"go.mongodb.org/mongo-driver/bson"
)

func unremitShipmentsFromTransaction(id string) (err error) {
	if id == "all" {
		_, err = db.ShippingCollection.UpdateMany(context.Background(), bson.M{}, bson.M{"$set": bson.M{"remitted": false}})
		return
	}
	transaction, err := models.GetTransactionByID(id)
	if err != nil {
		return
	}
	_, err = db.ShippingCollection.UpdateMany(context.Background(), bson.M{"_id": bson.M{"$in": transaction[0].ShipmentsRemitted}}, bson.M{"$set": bson.M{"remitted": false}})
	if err != nil {
		return
	}
	_, err = db.OrderCollection.UpdateMany(context.Background(), bson.M{"_id": bson.M{"$in": transaction[0].ShipmentsRemitted}}, bson.M{"$set": bson.M{"remitted": false}})
	return
}

func remitShipmentsFromTransaction(id string) (err error) {
	transaction, err := models.GetTransactionByID(id)
	if err != nil {
		return
	}
	_, err = db.ShippingCollection.UpdateMany(context.Background(), bson.M{"_id": bson.M{"$in": transaction[0].ShipmentsRemitted}}, bson.M{"$set": bson.M{"remitted": true}})
	if err != nil {
		return
	}
	_, err = db.OrderCollection.UpdateMany(context.Background(), bson.M{"_id": bson.M{"$in": transaction[0].ShipmentsRemitted}}, bson.M{"$set": bson.M{"remitted": true}})
	return
}

func addProductAndVariantToBackwardShipments() (err error) {
	var temp []models.BackwardShipment
	cur, err := db.BackwardShipmentCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return
	}
	err = cur.All(context.Background(), &temp)
	if err != nil {
		return
	}
	for _, val := range temp {
		tempVal := val
		for idx, childShipment := range val.ChildShipments {
			tempChildShipment := childShipment
			for index, productInfoItem := range childShipment.ProductInfo {
				var prod models.Product
				var vari models.Variation
				tempProductInfoItem := productInfoItem
				prod, err = models.GetCompleteProduct(tempProductInfoItem.ProductID)
				if err != nil {
					return
				}
				vari, err = models.GetVariant(tempProductInfoItem.VariantID)
				if err != nil {
					return
				}
				tempProductInfoItem.Product = prod
				tempProductInfoItem.Variant = vari
				tempChildShipment.ProductInfo[index] = tempProductInfoItem
			}
			tempVal.ChildShipments[idx] = tempChildShipment
		}
		db.BackwardShipmentCollection.ReplaceOne(context.Background(), bson.M{"_id": tempVal.ID}, tempVal)
	}
	return
}
