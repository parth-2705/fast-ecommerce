package scripts

import (
	"context"
	"fmt"
	"hermes/configs/Redis"
	"hermes/db"
	"hermes/models"

	"go.mongodb.org/mongo-driver/bson"
)

func removeMultipleRatings() error {
	var reviews []models.Review
	cursor, _ := db.ReviewCollection.Find(context.Background(), bson.D{})
	err := cursor.All(context.Background(), &reviews)
	if err != nil {
		fmt.Printf("Review err: %v\n", err)
	}

	// Get all Reviews

	// Iterate through reviews
	productsList := map[string]struct{}{}
	for _, review := range reviews {
		productsList[review.ProductID] = struct{}{}
	}

	// Get list of all products reviewd

	// for each product

	for productID := range productsList {

		product, err := models.GetProduct(productID)
		if err != nil {
			continue
		}

		// change it to get reviews in order
		reviews, err := models.GetAllReviewsForProduct(productID)
		if err != nil {
			continue
		}

		validReviews := map[string]models.Review{}
		totalRating := 0.0
		for _, review := range reviews {
			if _, ok := validReviews[review.UserID]; !ok {
				totalRating += review.Rating
				validReviews[review.UserID] = review
			}
		}

		// Delete all reviews for this product
		_, err = db.ReviewCollection.DeleteMany(context.Background(), bson.M{"productID": productID})
		if err != nil {
			continue
		}

		for _, val := range validReviews {
			_, err := db.ReviewCollection.InsertOne(context.Background(), val)
			if err != nil {
				continue
			}
		}

		product.RatingCount = len(validReviews)
		if product.RatingCount == 0 {
			product.AverageRating = 0
		} else {
			product.AverageRating = totalRating / float64((len(validReviews)))
		}

		db.ProductCollection.ReplaceOne(context.Background(), bson.M{"_id": productID}, product)

		Redis.DeleteProductCacheByID(productID)

	}

	fmt.Println("complete")
	return nil
	// get all reviews

	// Iterate through all and keep only 1 review per customer

	// delete all reviews for this product

	// caluculate new rating and rating count

	// add the new selected reviews back to db

	// update product

	// delete product cache
}
