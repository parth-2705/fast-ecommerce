package scripts

import (
	"context"
	"fmt"
	"hermes/db"
	"hermes/models"

	"go.mongodb.org/mongo-driver/bson"
)

func addPaymentMethodsToProducts() (err error) {

	// Get all Products
	products, err := models.GetAllProducts()
	if err != nil {
		return
	}

	paymentOptions := models.CopyPaymentOptions()

	fmt.Printf("len(paymentOptions): %v\n", len(paymentOptions))
	// Add Payment Methods to Products

	errCnt := 0
	errList := make([]string, 0)

	for i := range products {
		// iterate over all payment methods
		for _, po := range paymentOptions {

			if products[i].PaymentMethods == nil {
				products[i].PaymentMethods = make(models.PaymentMethodMap)
			}

			_, ok := products[i].PaymentMethods[po.ID]
			if !ok {
				products[i].PaymentMethods[po.ID] = models.PaymentMethodConfiguration{Available: true}
			}
		}

		// If payment method is already added to product, do nothing

		// if not added add it and use it

		// Save Product in DB
		_, err = db.ProductCollection.UpdateOne(context.Background(), bson.M{"_id": products[i].ID}, bson.M{"$set": products[i]})
		if err != nil {
			errCnt++
			errList = append(errList, products[i].ID)
		}
	}

	fmt.Printf("errCnt: %v\n", errCnt)
	for _, id := range errList {
		fmt.Printf("id: %v\n", id)
	}

	return nil
}
