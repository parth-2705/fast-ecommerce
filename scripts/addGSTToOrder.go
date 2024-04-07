package scripts

import (
	"context"
	"fmt"
	"hermes/configs/Redis"
	"hermes/db"
	"hermes/models"
	"hermes/utils/data"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func addGSTToOrder() (err error) {
	var orders []models.Order
	curr, err := db.OrderCollection.Find(context.Background(), bson.M{})
	if err != nil {
		panic(err)
	}

	err = curr.All(context.Background(), &orders)
	if err != nil {
		panic(err)
	}

	fmt.Println("orders: ", len(orders))

	for idx, order := range orders {
		fmt.Println("No: ", idx)

		var product models.Product
		err := db.ProductCollection.FindOne(context.Background(), bson.M{"_id": order.Product.ID}).Decode(&product)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				fmt.Println("product err: ", err, order.ID)
				continue
			} else {
				panic(err)
			}
		}

		var variant models.Variation
		err = db.VariantsCollection.FindOne(context.Background(), bson.M{"_id": order.Variant.ID}).Decode(&variant)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				fmt.Println("variant err: ", err, order.ID)
				continue
			} else {
				panic(err)
			}
		}

		order.Product.GST = product.GST
		order.Product.EANCode = product.EANCode
		order.Product.Cess = product.Cess

		order.Variant.Barcode = variant.Barcode

		_, err = db.OrderCollection.UpdateOne(context.Background(), bson.M{"_id": order.ID}, bson.M{"$set": order})
		if err != nil {
			panic(err)
		}
	}

	return nil
}

func updateAWBCodeForShipping() (err error) {

	var shiprocketOrders []models.FullShiprocketOrder
	curr, err := db.ShiprocketOrderCollection.Find(context.Background(), bson.M{})
	if err != nil {
		panic(err)
	}

	err = curr.All(context.Background(), &shiprocketOrders)
	if err != nil {
		panic(err)
	}

	for _, shiprocketOrder := range shiprocketOrders {
		// fmt.Println("idx: ", idx)

		var shipping models.Shipping
		err = db.ShippingCollection.FindOne(context.Background(), bson.M{"orderId": shiprocketOrder.SRID}).Decode(&shipping)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				fmt.Println(err, shiprocketOrder.SRID)
			} else {
				panic(err)
			}
		}

		if len(shiprocketOrder.AwbData.Awb) > 0 && shiprocketOrder.AwbData.Awb != shipping.AWB {
			shipping.AWB = shiprocketOrder.AwbData.Awb
			_, err = db.ShippingCharges.UpdateOne(context.Background(), bson.M{"orderId": shiprocketOrder.SRID}, bson.M{"$set": shipping})
			if err != nil {
				panic(err)
			}
		}

	}

	return nil
}

func addUserObjectToOrders() (err error) {
	var orders []models.Order
	curr, err := db.OrderCollection.Find(context.Background(), bson.M{})
	if err != nil {
		panic(err)
	}

	err = curr.All(context.Background(), &orders)
	if err != nil {
		panic(err)
	}

	fmt.Println("orders: ", len(orders))

	for _, order := range orders {
		// fmt.Println("No: ", idx)

		if len(order.User.ID) == 0 {
			user, err := models.GetUserByID(order.UserID)

			if err != nil {
				fmt.Println(order.ID, len(order.User.ID), order.UserID)

				if err == mongo.ErrNoDocuments {
					continue
				} else {
					panic(err)
				}

			}

			order.User = user

			_, err = db.OrderCollection.UpdateOne(context.Background(), bson.M{"_id": order.ID}, bson.M{"$set": order})
			if err != nil {
				panic(err)
			}
		}
	}

	return nil
}

func deleteOrderIfUserNotFound() (err error) {
	var orders []models.Order
	curr, err := db.OrderCollection.Find(context.Background(), bson.M{})
	if err != nil {
		panic(err)
	}

	err = curr.All(context.Background(), &orders)
	if err != nil {
		panic(err)
	}

	fmt.Println("orders: ", len(orders))

	for idx, order := range orders {
		if len(order.User.ID) == 0 {
			_, err := models.GetUserByID(order.UserID)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					fmt.Println("No: ", idx)
					_, err = db.OrderCollection.DeleteOne(context.Background(), bson.M{"_id": order.ID})
					if err != nil {
						panic(err)
					}
				} else {
					panic(err)
				}

			}
		}
	}

	return nil
}

func deleteProductFromRedisBySellerID(sellerID string) (err error) {

	if len(sellerID) == 0 {
		fmt.Printf("empty seller ID")
		return nil
	}

	var products []models.Product
	sellerID = strings.TrimSpace(sellerID)
	curr, err := db.ProductCollection.Find(context.Background(), bson.M{"sellerID": sellerID})
	if err != nil {
		panic(err)
	}

	err = curr.All(context.Background(), &products)
	if err != nil {
		panic(err)
	}

	fmt.Println("len: ", len(products))
	for _, product := range products {
		err := Redis.DeleteProductCacheByID(product.ID)
		if err != nil {
			panic(err)
		}
	}

	return nil
}

func CancelAllOrdersOfASeller(sellerID string, reason string) (err error) {

	sellerID = strings.TrimSpace(sellerID)
	if len(sellerID) == 0 {
		fmt.Printf("empty seller ID")
		return nil
	}

	reason = strings.TrimSpace(reason)
	if len(reason) == 0 {
		fmt.Printf("empty reason")
		return nil
	}

	var orders []models.Order
	curr, err := db.OrderCollection.Find(context.Background(), bson.M{"product.sellerID": sellerID})
	if err != nil {
		panic(err)
	}

	err = curr.All(context.Background(), &orders)
	if err != nil {
		panic(err)
	}

	fmt.Println("len: ", len(orders))
	for _, order := range orders {
		order.FulfillmentStatus = "Cancelled"
		order.CancellationReason = reason
		order.UpdatedAt = time.Now()

		err := order.Update()
		if err != nil {
			panic(err)
		}
	}

	return nil
}

func CreateDuplicateOrderBySeller(sellerID string, newSellerID string) (err error) {

	sellerID = strings.TrimSpace(sellerID)
	if len(sellerID) == 0 {
		fmt.Printf("empty seller ID")
		return nil
	}

	newSellerID = strings.TrimSpace(newSellerID)
	if len(newSellerID) == 0 {
		fmt.Printf("empty newSeller ID")
		return nil
	}

	var orders []models.Order
	curr, err := db.OrderCollection.Find(context.Background(), bson.M{"product.sellerID": sellerID})
	if err != nil {
		panic(err)
	}

	err = curr.All(context.Background(), &orders)
	if err != nil {
		panic(err)
	}

	fmt.Println("len: ", len(orders))
	for _, order := range orders {
		newOrder := order

		newOrder.ID = data.GetUUIDString("order")
		newOrder.UpdatedAt = time.Now()
		newOrder.Product.SellerID = newSellerID

		err := newOrder.CreateWithOutTimeUpdate()
		if err != nil {
			panic(err)
		}
	}

	return nil
}
