package models

import (
	"context"
	"fmt"
	"hermes/db"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type WishlistObject struct {
	WishlistID string    `json:"wishlistID" bson:"_wishlistID"`
	CreatedAt  time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt" bson:"updatedAt"`
	ProductId  string    `json:"productID" bson:"_productID"`
	Product    Product   `json:"product" bson:"product"`
}

func (WishlistObject WishlistObject) CreateIndexes() error {
	productIDModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "_productID", Value: 1},
		},
	}

	indexName, err := db.WishlistCollection.Indexes().CreateOne(context.Background(), productIDModel)
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

func GetAllProductsInWishlist(filters ...bson.D) ([]WishlistObject, error) {
	var wishlist []WishlistObject
	aggregrateSearchObject := bson.A{}
	for _, filter := range filters {
		aggregrateSearchObject = append(aggregrateSearchObject, filter)
	}

	aggregrateSearchObject = append(aggregrateSearchObject,
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "products"},
					{Key: "localField", Value: "_productID"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "product"},
				},
			},
		},
		bson.D{
			{Key: "$unwind",
				Value: bson.D{
					{Key: "path", Value: "$product"},
					{Key: "preserveNullAndEmptyArrays", Value: false},
				},
			},
		},
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "brands"},
					{Key: "localField", Value: "product.brandID"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "product.brand"},
				},
			},
		},
		bson.D{
			{Key: "$unwind",
				Value: bson.D{
					{Key: "path", Value: "$product.brand"},
					{Key: "preserveNullAndEmptyArrays", Value: true},
				},
			},
		},
	)
	productsCursor, err := db.WishlistCollection.Aggregate(context.Background(), aggregrateSearchObject)
	if err != nil {
		return wishlist, err
	}

	if err = productsCursor.All(context.TODO(), &wishlist); err != nil {
		return wishlist, err
	}
	defer productsCursor.Close(context.Background())
	return wishlist, nil
}

func GetIfItemIsInWishlistOfUser(user User, productID string) bool {
	var wishlistObject WishlistObject
	wishlistID, err := GetWishlistFromUser(user)
	if err != nil {
		fmt.Println("error", err)
		return false
	}
	err = db.WishlistCollection.FindOne(context.Background(), bson.M{"_wishlistID": wishlistID, "_productID": productID}).Decode(&wishlistObject)
	if err != nil {
		fmt.Println("error", err)
		return false
	}
	// fmt.Println("wishlistObject", wishlistObject)
	return true

}

func AddProductToWishlist(wishlistObject WishlistObject) error {
	_, err := db.WishlistCollection.InsertOne(context.Background(), wishlistObject)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func RemoveProductFromWishlist(wishlistObject WishlistObject) error {
	_, err := db.WishlistCollection.DeleteOne(context.Background(), wishlistObject)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func GetWishlistFromUser(user User) (string, error) {
	profile, err := user.GetProfile()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return profile.WishlistID, nil
}
