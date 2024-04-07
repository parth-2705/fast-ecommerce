package scripts

import (
	"context"
	"fmt"
	"hermes/db"
	"hermes/models"

	"go.mongodb.org/mongo-driver/bson"
)

var categoryMap = map[string]string{
	"Hat":            "category-2fc944f8-7555-4ba9-9c6c-d1666554ba10",
	"Cap":            "category-f8c410a6-033b-40b0-9474-b7e76875286e",
	"Stockings":      "category-af931c5e-61e7-43bf-bbea-b890ee794141",
	"Duffel Bag":     "category-dfbc1353-68ef-428a-8e4c-4e7baa2e947e",
	"Laptop Bag":     "category-3183fc1b-8d78-4902-b817-c2a124dd10d1",
	"Waist Pouch":    "category-42aed7d3-180f-467a-91e5-841a776abe65",
	"Sunglass":       "category-1171b557-4f33-4b4f-9c1d-113dfb460769",
	"Charms":         "category-c67b5a53-da57-4dc9-ba31-61cdffaf37f8",
	"Handbags":       "category-f138f4b0-f224-4251-8518-94b8c6c33f86",
	"Belt":           "category-602751fc-6cc4-4637-88be-1f57c58d6d93",
	"Watch Gift Set": "category-945302ad-9b07-4827-84d7-d8b5a2039b0d",
	"Stoles":         "category-ec7181d9-ac73-4e71-b786-bba88d73f539",
	"Socks":          "category-9ea60a6b-782e-4572-8a4a-64459bd115d4",
	"Wallets":        "category-bcf99b9e-1ea6-4cf7-85b9-05c5340e8f64",
	"Clutches":       "category-d8efd29c-d06c-4d7d-8ced-ab0e2b4123e3",
	"Watch":          "category-6febdfca-e48a-4849-b855-35286f13ea28",
	"Backpacks":      "category-417a30bf-9e63-4a31-ac22-d084963344f6",
}

func updateEmptyNewCategory() (err error) {
	var products []models.Product
	cur, err := db.ProductCollection.Find(context.Background(), bson.M{"newCategory": ""})
	if err != nil {
		return
	}
	err = cur.All(context.Background(), &products)
	if err != nil {
		return
	}
	for idx, val := range products {
		var tempCategory string
		if newCat, ok := categoryMap[val.ProductType]; ok {
			tempCategory = newCat
		}
		_, err = db.ProductCollection.UpdateOne(context.Background(), bson.M{"_id": val.ID}, bson.M{"$set": bson.M{"newCategory": tempCategory}})
		if err != nil {
			return
		}
		fmt.Println("Index:", idx, " done")
	}
	return nil
}

func updateEANCode(productID string, eanCode string) (err error) {
	_, err = db.ProductCollection.UpdateOne(context.Background(), bson.M{"_id": productID}, bson.M{"$set": bson.M{"EANCode": eanCode}})
	return
}
