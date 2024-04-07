package models

import (
	"context"
	"fmt"
	"hermes/configs"
	"hermes/configs/Redis"
	"hermes/db"
	"hermes/utils/data"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type pageType int

const (
	ProductPage pageType = iota
)

type PageView struct {
	ID          string    `json:"id" bson:"_id"`
	PageType    pageType  `json:"pageType" bson:"pageType"`
	VisitTime   time.Time `json:"visitTime" bson:"visitTime"`
	PhoneNumber string    `json:"phoneNumber" bson:"phoneNumber"`
	UserID      string    `json:"userID" bson:"userID"`
	UAID        string    `json:"userAgentID" bson:"userAgentID"`
	ResourceID  string    `json:"resourceID" bson:"resourceID"`
	Extra       any       `json:"extra" bson:"extra"`
	Resource    Product   `json:"resource" bson:"resource"`
}

func (PageView PageView) CreateIndexes() error {
	indexModels := []mongo.IndexModel{}

	phoneNumberModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "phoneNumber", Value: 1},
		},
	}
	indexModels = append(indexModels, phoneNumberModel)

	visitTimeModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "visitTime", Value: -1},
		},
	}
	indexModels = append(indexModels, visitTimeModel)

	userIDModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "userID", Value: 1},
		},
	}
	indexModels = append(indexModels, userIDModel)

	indexName, err := db.PageViewCollection.Indexes().CreateMany(context.Background(), indexModels)
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

func CreateProductPageViewEntry(ctx context.Context, phoneNumber string, userID string, productID string, userAgentID string) (err error) {

	defer sentry.Recover()

	pageView := PageView{
		ID:          data.GetUUIDString("PageView"),
		PageType:    ProductPage,
		VisitTime:   time.Now(),
		PhoneNumber: phoneNumber,
		UserID:      userID,
		UAID:        userAgentID,
		ResourceID:  productID,
		Extra:       nil,
		Resource:    Product{},
	}

	_, err = db.PageViewCollection.InsertOne(context.Background(), pageView)
	if err != nil {
		fmt.Println(err)
		return
	}

	return

}

func AddToRecentlyViewedList(ctx context.Context, c *gin.Context, productID string) error {

	defer sentry.Recover()

	userAgentID, _ := data.GetSessionValue(c, configs.UserAgentIdentifier).(string)

	// Add to Set for this user agent with unix timestamp (integer) as score
	cmd := Redis.Client.ZAdd(context.Background(), userAgentID, redis.Z{Score: float64(time.Now().Unix()), Member: productID})
	if cmd.Err() != nil {
		fmt.Println(cmd.Err())
		return cmd.Err()
	}

	// get set size
	cmd = Redis.Client.ZCard(context.Background(), userAgentID)
	if cmd.Err() != nil {
		fmt.Println(cmd.Err())
		return cmd.Err()
	}

	num, err := cmd.Result()
	if err != nil {
		fmt.Println(err)
		return err
	}

	if num > 10 { // if size of set increases 10
		Redis.Client.ZRemRangeByRank(context.Background(), userAgentID, 0, 0) //remove the lowest ranking product
	}

	return nil
}

func GetRecentlyViewedProducts(c *gin.Context, productToFilterOut string) (products []Product, err error) {

	userAgentID, _ := data.GetSessionValue(c, configs.UserAgentIdentifier).(string)
	userIDInterface := data.GetSessionValue(c, configs.Userkey)
	userID, _ := userIDInterface.(string)

	userAgentAndUserAggregation := bson.A{}

	if userAgentID != "" {
		userAgentAndUserAggregation = append(userAgentAndUserAggregation, bson.D{{"userAgentID", userAgentID}})
	}

	if userID != "" {
		userAgentAndUserAggregation = append(userAgentAndUserAggregation, bson.D{{"userID", userAgentID}})
	}

	//Create Filter
	filter := bson.D{
		{"$or", userAgentAndUserAggregation},
		{"userAgentID", userAgentID},
		{"pageType", ProductPage},
		{"resourceID", bson.D{{"$ne", productToFilterOut}}},
	}

	options := []*options.FindOptions{
		options.Find().SetSort(bson.D{{"visitTime", -1}}),
		options.Find().SetLimit(10),
	}

	// make mongodb query
	// should match userAgentID (or UserID ??)
	// Should be product Page
	// Filter Out Current Product
	// Limit 10
	cursor, err := db.PageViewCollection.Find(context.Background(), filter, options...)
	if err != nil {
		fmt.Println(err)
		return
	}

	// get Pageviews from mongo db
	var pages []PageView
	err = cursor.All(context.Background(), &pages)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create List of productIDs
	// Fetch those Products

	for _, page := range pages {
		product, err := GetCompleteProduct(page.ResourceID)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, product)
	}

	return
}

func GetRecentlyViewedProductsFromRedis(c *gin.Context, productToSkip string) ([]Product, error) {
	userAgentID, _ := data.GetSessionValue(c, configs.UserAgentIdentifier).(string)

	// Get the complete set
	cmd := Redis.Client.ZRange(context.Background(), userAgentID, 0, -1)
	if cmd.Err() != nil {
		fmt.Println(cmd.Err())
		return nil, cmd.Err()
	}

	// product IDs are received as a list of string
	productIDs, err := cmd.Result()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	products := make([]Product, 0)
	for i := range productIDs {

		// No need to show the current product in history
		if productIDs[i] == productToSkip {
			continue
		}

		product, err := GetCompleteProduct(productIDs[i])
		if err != nil {
			fmt.Println(err)
			continue
		}

		products = append(products, product)
	}

	// reverse the product list because redis returns it in increasing rank of scores
	for i, j := 0, len(products)-1; i < j; i, j = i+1, j-1 {
		products[i], products[j] = products[j], products[i]
	}

	return products, nil
}
