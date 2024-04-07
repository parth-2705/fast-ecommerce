package scripts

import (
	"context"
	"encoding/json"
	"fmt"
	"hermes/db"
	"hermes/models"
	"strings"
	"sync"

	"github.com/getsentry/sentry-go"
	"github.com/tryamigo/themis"
	"go.mongodb.org/mongo-driver/bson"
)

type imagoTransferResponse struct {
	Path string `json:"path"`
}

var imagoURL = "https://imago.amigotest.org/transfer-to-bucket/"

var bucketBaseURL = "https://storage.googleapis.com/roovo-images/rawImages/"

func getBucketURL(newURL string) (bucketURL string) {
	bucketURL = newURL
	status, resp, err := themis.HitAPIEndpoint2(imagoURL+newURL, "POST", nil, nil, nil)
	if status >= 400 || err != nil {
		return
	}
	var urlResponseBody imagoTransferResponse
	json.Unmarshal(resp, &urlResponseBody)

	bucketURL = strings.Replace(urlResponseBody.Path, bucketBaseURL, "", -1)
	return
}

func removeURLFromImages() (err error) {
	var products []models.Product
	var newProducts []interface{}
	cur, err := db.ProductCollection.Find(context.Background(), bson.M{})
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	err = cur.All(context.Background(), &products)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	var wg sync.WaitGroup
	for productIdx, product := range products {
		images := product.Images
		for idx, image := range images {
			wg.Add(1)
			go func(idx int, image models.Image) {
				defer sentry.Recover()
				defer wg.Done()
				newURL := strings.Replace(string(image), "rawImages/", "", -1)
				// newURL = getBucketURL(newURL)
				product.Images[idx] = models.Image(newURL)
			}(idx, image)
		}
		wg.Wait()
		product.Thumbnail = product.Images[0]
		newProducts = append(newProducts, product)
		fmt.Println(productIdx+1, "/", len(products), " PRODUCTS DONE")
	}
	resultsDeleted, _ := db.ProductCollection.DeleteMany(context.Background(), bson.M{})
	fmt.Println()
	fmt.Println("DELETED", resultsDeleted.DeletedCount)
	fmt.Println()
	resultsInserted, _ := db.ProductCollection.InsertMany(context.Background(), newProducts)
	fmt.Println()
	fmt.Println("INSERTED", len(resultsInserted.InsertedIDs))
	fmt.Println()

	var variants []models.Variation
	var newVariants []interface{}
	cur, err = db.VariantsCollection.Find(context.Background(), bson.M{})
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	err = cur.All(context.Background(), &variants)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	for variantIdx, variant := range variants {
		images := variant.Images
		for idx, image := range images {
			wg.Add(1)
			go func(idx int, image models.Image) {
				defer sentry.Recover()
				defer wg.Done()
				newURL := strings.Replace(string(image), "rawImages/", "", -1)
				// newURL = getBucketURL(newURL)
				variant.Images[idx] = models.Image(newURL)
			}(idx, image)
		}
		wg.Wait()
		variant.Thumbnail = variant.Images[0]
		newVariants = append(newVariants, variant)
		fmt.Println(variantIdx+1, "/", len(variants), " VARIANTS DONE")
	}
	resultsDeleted, _ = db.VariantsCollection.DeleteMany(context.Background(), bson.M{})
	fmt.Println()
	fmt.Println("DELETED", resultsDeleted.DeletedCount)
	fmt.Println()
	resultsInserted, _ = db.VariantsCollection.InsertMany(context.Background(), newVariants)
	fmt.Println()
	fmt.Println("INSERTED", len(resultsInserted.InsertedIDs))
	fmt.Println()

	return
}

func fixImageURLsInCategories() (err error) {
	categories, err := models.GetCategories()
	if err != nil {
		return err
	}
	for idx, category := range categories {
		url := string(category.Image)
		url = strings.Replace(url, "https://storage.googleapis.com/roovo-images/rawImages/", "", 1)
		url = strings.Replace(url, "https://storage.googleapis.com/roovo/", "", 1)
		category.Image = models.Image(url)
		_, err := db.CategoryCollection.ReplaceOne(context.Background(), bson.M{"_id": category.ID}, &category)
		if err != nil {
			fmt.Printf("err: %d %v\n", idx, err)
		}
	}
	return
}
