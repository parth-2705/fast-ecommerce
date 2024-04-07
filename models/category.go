package models

import (
	"context"
	"fmt"
	"hermes/db"
	"hermes/utils/data"
	utils "hermes/utils/queries"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Category struct {
	ID                 string     `json:"id" bson:"_id"`
	CreatedAt          time.Time  `json:"createdAt" bson:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt" bson:"updatedAt"`
	Image              Image      `json:"image" bson:"image"`
	Name               string     `json:"name" bson:"name"`
	Description        string     `json:"description" bson:"description"`
	StoreID            string     `json:"storeID" bson:"storeID"`
	Filters            Filter     `json:"filters" bson:"filters"`
	ParentID           string     `json:"parentID" bson:"parentID"`
	Parent             *Category  `json:"parent" bson:"parent"`
	ChildrenCategories []Category `json:"childrenCategories" bson:"childrenCategories"`
}

func (category Category) CreateIndexes() error {
	indexModels := []mongo.IndexModel{}

	createdAtModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "createdAt", Value: -1},
		},
	}
	indexModels = append(indexModels, createdAtModel)

	nameIndex := mongo.IndexModel{
		Keys: bson.M{
			"name": "text",
		},
	}
	indexModels = append(indexModels, nameIndex)

	updatedAtModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "updatedAt", Value: -1},
		},
	}
	indexModels = append(indexModels, updatedAtModel)

	indexName, err := db.CategoryCollection.Indexes().CreateMany(context.Background(), indexModels)
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

type FilterOption struct {
	Label string `json:"name" bson:"name"`
	Value string `json:"ID" bson:"ID"`
}

type Filter map[string][]FilterOption

func (newCategory Category) Create() (err error) {
	newCategory.ID = data.GetUUIDString("category")
	newCategory.Parent = nil
	_, err = db.CategoryCollection.InsertOne(context.Background(), newCategory)
	return
}

func (category Category) Update() (err error) {
	category.Parent = nil
	_, err = db.CategoryCollection.ReplaceOne(context.Background(), bson.M{"_id": category.ID}, category)
	return
}

func GetFullCategoryByID(categoryID string) (category Category, err error) {
	var categories []Category
	pipeline := bson.A{bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: categoryID}}}}}
	pipeline = append(pipeline, utils.FullCategoryQuery...)
	cur, err := db.CategoryCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return
	}
	err = cur.All(context.Background(), &categories)
	if err != nil {
		return
	}
	if len(categories) == 0 {
		return category, fmt.Errorf("No category with given ID")
	}
	return categories[0], nil
}

func GetCategoryByID(categoryID string) (category Category, err error) {
	err = db.CategoryCollection.FindOne(context.Background(), bson.M{"_id": categoryID}).Decode(&category)
	return
}

func GetCategories() (categories []Category, err error) {
	cur, err := db.CategoryCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return
	}
	err = cur.All(context.Background(), &categories)
	return
}
