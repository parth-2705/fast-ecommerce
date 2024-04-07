package models

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"hermes/configs/Redis"
	"hermes/db"
	"html/template"
	"io/ioutil"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AttributeVisualisationMethod int

const (
	Text AttributeVisualisationMethod = iota
	Color
	DisplayImage
)

type KeySpec struct {
	Key   string `json:"key" bson:"key"`
	Value string `json:"value" bson:"value"`
}

type Product struct {
	ID                  string               `json:"id" bson:"_id"`
	CreatedAt           time.Time            `json:"createdAt" bson:"createdAt"`
	UpdatedAt           time.Time            `json:"updatedAt" bson:"updatedAt"`
	KeySpec             KeySpec              `json:"keySpec" bson:"keySpec"`
	Name                string               `json:"name" bson:"name"`
	Description         template.HTML        `json:"description" bson:"description"`
	DescriptionEncoded  string               `json:"descriptionEncoded" bson:"descriptionEncoded"`
	RatingCount         int                  `json:"ratingCount" bson:"ratingCount"`
	AverageRating       float64              `json:"rating" bson:"rating"`
	RatingVisualization RatingVisualization  `json:"ratingVisualization" bson:"ratingVisualization"`
	Price               ProductPrice         `json:"price" bson:"price"`
	Quantity            int                  `json:"quantity" bson:"quantity"`
	Images              []Image              `json:"images" bson:"images"`
	Thumbnail           Image                `json:"thumbnail" bson:"thumbnail"`
	Media               []MediaObject        `json:"media" bson:"media"`
	Category            string               `json:"category" bson:"category"`
	Specifications      []KV                 `json:"specifications" bson:"specifications"`
	Overview            string               `json:"overview" bson:"overview"`
	AllDetails          []DetailObject       `json:"allDetails" bson:"allDetails"`
	SKU                 string               `json:"sku" bson:"sku"`
	Barcode             string               `json:"barcode" bson:"barcode"`
	Status              string               `json:"status" bson:"status"`
	BrandID             string               `json:"brandID" bson:"brandID"`
	Brand               Brand                `json:"brand" bson:"brand"`
	SellerID            string               `json:"sellerID" bson:"sellerID"`
	StoreID             string               `json:"storeID" bson:"storeID"`
	Location            string               `json:"location" bson:"location"`
	Attributes          []VariationAttribute `json:"attributes" bson:"attributes"`
	VariationIDs        []string             `json:"variationsIDs" bson:"variationIDs"`
	Variants            []Variation          `json:"variants" bson:"variants"`
	GST                 string               `json:"gst" bson:"gst"`
	Cess                string               `json:"cess" bson:"cess"`
	SubCategory         string               `json:"subCategory" bson:"subCategory"`
	ProductType         string               `json:"productType" bson:"productType"`
	Gender              string               `json:"gender" bson:"gender"`
	ShelfLifeDays       int                  `json:"shelfLifeDays" bson:"shelfLifeDays"`
	IsBestSeller        bool                 `json:"isBestSeller" bson:"isBestSeller"`
	IsHotDeal           bool                 `json:"isHotDeal" bson:"isHotDeal"`
	EANCode             string               `json:"EANCode" bson:"EANCode"`
	Similarity          map[string]int       `json:"similarity" bson:"similarity"`
	PageRanking         int                  `json:"pageRanking" bson:"pageRanking"`
	NewCategory         string               `json:"newCategory" bson:"newCategory"`
	PaymentMethods      PaymentMethodMap     `json:"paymentMethods" bson:"paymentMethods"`
	DeliveryTime        DeliveryTime         `json:"deliveryTime" bson:"deliveryTime"`
}

type DeliveryTime struct {
	AverageDeliveryTime int         `json:"avgDeliveryTime" bson:"avgDeliveryTime"`
	DeliveryCompleted   int         `json:"deliveryCompleted" bson:"deliveryCompleted"`
	DeliveryMap         map[int]int `json:"deliveryMap" bson:"deliveryMap"`
}

func (Product) GormDataType() string {
	return "json"
}

func (data Product) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *Product) Scan(value interface{}) error {

	if value == nil {
		return nil
	}

	var byteSlice []byte
	switch v := value.(type) {
	case []byte:
		if len(v) > 0 {
			byteSlice = make([]byte, len(v))
			copy(byteSlice, v)
		}
	case string:
		byteSlice = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	err := json.Unmarshal(byteSlice, &data)
	return err
}

func (Product Product) CreateIndexes() error {
	indexModels := []mongo.IndexModel{}

	createdAtModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "createdAt", Value: -1},
		},
	}
	indexModels = append(indexModels, createdAtModel)

	sellerIDModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "sellerID", Value: 1},
		},
	}
	indexModels = append(indexModels, sellerIDModel)

	brandIDModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "brandID", Value: 1},
		},
	}
	indexModels = append(indexModels, brandIDModel)

	nameIndex := mongo.IndexModel{
		Keys: bson.M{
			"name": "text",
		},
	}
	indexModels = append(indexModels, nameIndex)

	indexName, err := db.ProductCollection.Indexes().CreateMany(context.Background(), indexModels)
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

func (product Product) MarshalBinary() (data []byte, err error) {
	marshaled, err := json.Marshal(product)
	if err != nil {
		return
	}

	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err = gz.Write(marshaled); err != nil {
		return
	}
	if err = gz.Flush(); err != nil {
		return
	}
	if err = gz.Close(); err != nil {
		return
	}

	data = b.Bytes()

	return
}

type Variation struct {
	ProductID  string            `json:"productID" bson:"productID"`
	ID         string            `json:"id" bson:"_id"`
	CreatedAt  time.Time         `json:"createdAt" bson:"createdAt"`
	UpdatedAt  time.Time         `json:"updatedAt" bson:"updatedAt"`
	Price      ProductPrice      `json:"price" bson:"price"`
	Quantity   int               `json:"quantity" bson:"quantity"`
	Images     []Image           `json:"images" bson:"images"`
	Media      []MediaObject     `json:"media" bson:"media"`
	Thumbnail  Image             `json:"thumbnail" bson:"thumbnail"`
	SKU        string            `json:"sku" bson:"sku"`
	Barcode    string            `json:"barcode" bson:"barcode"`
	Status     string            `json:"status" bson:"status"`
	Weight     string            `json:"weight" bson:"weight"`
	Length     string            `json:"length" bson:"length"`
	Breadth    string            `json:"breadth" bson:"breadth"`
	Height     string            `json:"height" bson:"height"`
	Attributes map[string]string `json:"attributes" bson:"attributes"`
}

func (Variation) GormDataType() string {
	return "json"
}

func (data Variation) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *Variation) Scan(value interface{}) error {

	if value == nil {
		return nil
	}

	var byteSlice []byte
	switch v := value.(type) {
	case []byte:
		if len(v) > 0 {
			byteSlice = make([]byte, len(v))
			copy(byteSlice, v)
		}
	case string:
		byteSlice = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	err := json.Unmarshal(byteSlice, &data)
	return err
}

func (Variation Variation) CreateIndexes() error {
	indexModels := []mongo.IndexModel{}

	productIDModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "productID", Value: 1},
		},
	}
	indexModels = append(indexModels, productIDModel)

	op := options.Index()
	t := true
	op.Unique = &t
	skuModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "sku", Value: -1},
		},
		Options: op,
	}
	indexModels = append(indexModels, skuModel)

	indexName, err := db.VariantsCollection.Indexes().CreateMany(context.Background(), indexModels)
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

type VariationAttribute struct {
	Name    string                       `json:"name" bson:"name"`
	VisType AttributeVisualisationMethod `json:"visType" bson:"visType"`
	Options []string                     `json:"options" bson:"options"`
}

type DetailObject struct {
	Image   Image         `json:"image" bson:"image"`
	Heading string        `json:"heading" bson:"heading"`
	Text    template.HTML `json:"text" bson:"text"`
}

type KV struct {
	Key   string `json:"key" bson:"key"`
	Value string `json:"value" bson:"value"`
}

type ProductPrice struct {
	SellingPrice       float64 `json:"sellingPrice" bson:"sellingPrice"`
	MRP                float64 `json:"mrp" bson:"mrp"`
	Discount           float64 `json:"discount" bson:"discount"`
	DiscountPercentage float64 `json:"discountPercentage" bson:"discountPercentage"`
	SellerPrice        float64 `json:"sellerPrice" bson:"sellerPrice"`
}

func (price *ProductPrice) Add(priceToAdd ProductPrice, quantity int) {
	price.SellingPrice += (priceToAdd.SellingPrice) * float64(quantity)
	price.MRP += (priceToAdd.MRP) * float64(quantity)
	price.Discount += priceToAdd.Discount * float64(quantity)
}

type Image string

type MediaObject struct {
	Type        MediaType `bson:"type" json:"type"`
	ID          string    `bson:"id" json:"id"`
	ThumbnailID string    `bson:"thumbnailID" json:"thumbnailID"`
}

type MediaType string

const (
	TypeImage MediaType = "Image"
	TypeVideo MediaType = "Video"
)

func GetProduct(productID string) (Product, error) {
	// Find the product in the database
	var product Product
	err := db.ProductCollection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&product)
	if err != nil {
		fmt.Println(err)
		return product, err
	}

	return product, nil
}

func getCompleteProduct(productID string) (product Product, err error) {
	cursor, err := db.ProductCollection.Aggregate(context.Background(),
		bson.A{
			bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: productID}}}},
			bson.D{
				{Key: "$lookup",
					Value: bson.D{
						{Key: "from", Value: "brands"},
						{Key: "localField", Value: "brandID"},
						{Key: "foreignField", Value: "_id"},
						{Key: "as", Value: "brand"},
					},
				},
			},
			bson.D{
				{Key: "$unwind",
					Value: bson.D{
						{Key: "path", Value: "$brand"},
						{Key: "preserveNullAndEmptyArrays", Value: true},
					},
				},
			},
			bson.D{
				{Key: "$lookup",
					Value: bson.D{
						{Key: "from", Value: "variants"},
						{Key: "localField", Value: "_id"},
						{Key: "foreignField", Value: "productID"},
						{Key: "as", Value: "variants"},
					},
				},
			},
		},
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	cursor.Next(context.Background())
	err = cursor.Decode(&product)
	if err != nil {
		fmt.Println(err)
		return
	}

	cursor.Close(context.Background())
	return
}

func getProductFromRedis(productID string) (product Product, err error) {
	result, err := Redis.Client.Get(context.Background(), productID).Result()
	if err != nil {
		fmt.Println(err)
		return
	}

	rdata := bytes.NewReader([]byte(result))
	r, _ := gzip.NewReader(rdata)
	s, _ := ioutil.ReadAll(r)

	json.Unmarshal([]byte(s), &product)
	return
}

func setProductInRedis(product Product) (err error) {
	err = Redis.Client.Set(context.Background(), product.ID, product, 0).Err()
	if err != nil {
		fmt.Println(err)
	}
	return
}

func UpdateAverageDeliveryTimeForProduct(productID string, avgTime int) (err error) {
	product, err := GetProduct(productID)
	if err != nil {
		return
	}
	var newMap map[int]int
	newMap = product.DeliveryTime.DeliveryMap
	if product.DeliveryTime.DeliveryMap == nil {
		newMap = make(map[int]int)
	}
	newDeliveries := product.DeliveryTime.DeliveryCompleted + 1
	newAvg := (product.DeliveryTime.AverageDeliveryTime*product.DeliveryTime.DeliveryCompleted + avgTime) / newDeliveries
	if val, ok := newMap[avgTime]; !ok {
		newMap[avgTime] = 1
	} else {
		newMap[avgTime] = val + 1
	}
	newDeliveryTime := DeliveryTime{
		AverageDeliveryTime: newAvg,
		DeliveryCompleted:   newDeliveries,
		DeliveryMap:         newMap,
	}
	_, err = db.ProductCollection.UpdateOne(context.Background(), bson.M{"_id": productID}, bson.M{"$set": bson.M{"deliveryTime": newDeliveryTime}})
	if err != nil {
		return
	}
	err = Redis.DeleteProductCacheByID(productID)
	return
}

func GetCompleteProduct(productID string) (product Product, err error) {

	if len(productID) == 0 {
		return Product{}, fmt.Errorf("empty product ID")
	}

	// fmt.Println("Getting product from redis")

	// Check Redis for this product
	product, err = getProductFromRedis(productID)

	// If found in Cache, return that product
	if err == nil {
		// fmt.Println("Found product in redis")
		return
	}

	// fmt.Println("DId not find product in redis: ", productID)
	// If not found, Find in Mongo
	product, err = getCompleteProduct(productID)
	if err != nil {
		return
	}

	// fmt.Println("Setting product in redis")
	// Set product in Redis
	err = setProductInRedis(product)
	if err != nil {
		return
	}

	// fmt.Println("prodecut set in redi")

	// Return the product
	return
}

func GetVariant(variantID string) (Variation, error) {
	// Find the product in the database
	var variant Variation
	// _id is of type primitive.ObjectID
	// variantObjectID, err := primitive.ObjectIDFromHex(variantID)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return variant, err
	// }
	err := db.VariantsCollection.FindOne(context.Background(), bson.M{"_id": variantID}).Decode(&variant)
	if err != nil {
		fmt.Println(err)
		return variant, err
	}

	return variant, nil
}

func (variant Variation) Update() error {
	_, err := db.VariantsCollection.UpdateOne(context.Background(), bson.M{"_id": variant.ID}, bson.M{"$set": variant})
	if err != nil {
		return err
	}

	return nil
}

func GetVariantsByProductID(productID string) ([]Variation, error) {
	var variants []Variation
	cur, err := db.VariantsCollection.Find(context.Background(), bson.M{"productID": productID}, options.Find())
	if err != nil {
		fmt.Println(err)
		return variants, err
	}

	err = cur.All(context.Background(), &variants)
	if err != nil {
		return variants, err
	}
	return variants, nil
}

func GetAllVariants() ([]Variation, error) {
	var variants []Variation
	cur, err := db.VariantsCollection.Find(context.Background(), bson.M{}, options.Find())
	if err != nil {
		fmt.Println(err)
		return variants, err
	}

	err = cur.All(context.Background(), &variants)
	if err != nil {
		return variants, err
	}
	return variants, nil
}

func GetVariantByPrimitiveID(variantID string) (Variation, error) {
	// Find the product in the database
	var variant Variation
	// _id is of type primitive.ObjectID
	variantObjectID, err := primitive.ObjectIDFromHex(variantID)
	if err != nil {
		fmt.Println(err)
		return variant, err
	}

	err = db.VariantsCollection.FindOne(context.Background(), bson.M{"_id": variantObjectID}).Decode(&variant)
	if err != nil {
		fmt.Println(err)
		return variant, err
	}

	return variant, nil
}

// get all products for creating a deal in admin page / Or other Uses *wink* *wink*
func GetAllProducts() ([]Product, error) {
	var products []Product
	cur, err := db.ProductCollection.Find(context.Background(), bson.M{}, options.Find())
	if err != nil {
		return products, err
	}
	err = cur.All(context.Background(), &products)
	if err != nil {
		return products, err
	}
	return products, nil
}

func (product Product) Update() error {
	_, err := db.ProductCollection.UpdateOne(context.Background(), bson.M{"_id": product.ID}, bson.M{"$set": product})
	if err != nil {
		return err
	}

	return nil
}

// func that takes range x, y and decimal places n and returns a random float64
func RandomFloat(x, y, n float64) float64 {
	return math.Round((x+rand.Float64()*(y-x))*math.Pow(10, n)) / math.Pow(10, n)
}

func RandomFloat32WithNDP(x, y float32, n int) float32 {
	// Set the seed value for the random number generator
	rand.Seed(time.Now().UnixNano())

	// Calculate the range of values
	rangeValue := y - x

	// Generate a random float32 value between 0 and 1
	randomValue := rand.Float32()

	// Scale the random value to the range of values and add the minimum value
	scaledValue := (randomValue * rangeValue) + x

	// Round the scaled value to n decimal places
	roundedValue := float32(math.Round(float64(scaledValue)*math.Pow(10, float64(n))) / math.Pow(10, float64(n)))

	return roundedValue
}

func RoundDecimals(numToRound float64, placesToRound int) float64 {
	return math.Round(numToRound*math.Pow(10, float64(placesToRound))) / math.Pow(10, float64(placesToRound))
}

func CreateDummyData(n int) error {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	categories := make([]Category, 3)
	for i := 0; i < 3; i++ {
		category := Category{
			ID:    gofakeit.UUID(),
			Name:  gofakeit.Name(),
			Image: Image(gofakeit.ImageURL(600, 400)),
		}
		categories[i] = category
	}

	brands, err := createDummyBrands(5)
	if err != nil {
		return err
	}

	for _, category := range categories {
		_, err := db.CategoryCollection.InsertOne(context.Background(), category)
		if err != nil {
			return err
		}
	}

	// Generate n dummy products using gofakeit
	products := make([]Product, n)
	for i := 0; i < n; i++ {

		// mrp is a float with only two decimal places
		mrp := RandomFloat(100, 10000, 2)
		discountPercent := RandomFloat(5, 95, 1)

		// round off selling price to two decimal places
		sellingPrice := math.Round((mrp-(mrp*discountPercent/100))*100) / 100
		discount := RoundDecimals(mrp-sellingPrice, 2)

		rating := RandomFloat(0.0, 5.0, 1)
		status := ""

		if time.Now().UTC().UnixNano()%2 == 0 {
			status = "Draft"
		} else {
			status = "Active"
		}

		product := Product{
			ID:          gofakeit.UUID(),
			Name:        gofakeit.Name(),
			Description: template.HTML("<p>Can you see the p tags?</p>"),
			Price: ProductPrice{
				SellingPrice:       sellingPrice,
				MRP:                mrp,
				Discount:           discount,
				DiscountPercentage: discountPercent,
			},
			Quantity:      gofakeit.Number(1, 1000),
			AverageRating: rating,
			RatingCount:   gofakeit.IntRange(10, 200),
			Images:        make([]Image, 0),
			Thumbnail:     Image(gofakeit.ImageURL(600, 400)),
			Category:      categories[gofakeit.Number(0, 2)].ID,
			Overview:      gofakeit.Person().Job.Descriptor,
			Status:        status,
			Barcode:       gofakeit.Name(),
			SKU:           gofakeit.Name(),
			BrandID:       brands[gofakeit.Number(0, len(brands)-1)].ID,
		}
		for j := 0; j < gofakeit.Number(3, 8); j++ {
			image := Image(gofakeit.ImageURL(600, 400))
			spec := KV{
				Key:   gofakeit.AdjectiveDescriptive(),
				Value: gofakeit.AdjectiveQuantitative(),
			}
			deets := DetailObject{
				Image:   image,
				Heading: gofakeit.CarModel(),
				Text:    template.HTML("<p>Can you see the p tags?</p>"),
			}
			product.Specifications = append(product.Specifications, spec)
			product.Images = append(product.Images, image)
			product.AllDetails = append(product.AllDetails, deets)

		}
		products[i] = product
	}

	// Insert the products into the database
	for _, product := range products {
		_, err := db.ProductCollection.InsertOne(context.Background(), product)
		if err != nil {
			return err
		}
	}

	users, err := createDummyUsers(10)
	if err != nil {
		return err
	}

	err = createDummyReviews(products, users, 20)
	if err != nil {
		return err
	}

	return nil
}

func createDummyBrands(count int) (brands []Brand, err error) {
	for i := 0; i < count; i++ {
		brand := Brand{
			ID:          gofakeit.UUID(),
			Name:        gofakeit.BeerName(),
			Description: template.HTML(gofakeit.Sentence(6)),
			Logo:        gofakeit.ImageURL(20, 20),
		}
		_, err = db.BrandsCollection.InsertOne(context.Background(), brand)
		if err != nil {
			fmt.Println(err)
			return
		}

		brands = append(brands, brand)
	}

	return
}

func createDummyUsers(count int) (users []User, err error) {

	for i := 0; i < count; i++ {
		user := User{
			ID:        gofakeit.UUID(),
			CreatedAt: gofakeit.Date(),
			Name:      gofakeit.Name(),
			Phone:     gofakeit.Phone(),
		}

		_, err = db.UserCollection.InsertOne(context.Background(), user)
		if err != nil {
			fmt.Println(err)
			return
		}

		users = append(users, user)
	}
	return
}

func createDummyReviews(products []Product, users []User, count int) (err error) {

	for i := 0; i <= count; i++ {

		reviewDate := gofakeit.Date()

		review := Review{
			ID:           gofakeit.UUID(),
			ProductID:    products[gofakeit.Number(0, len(products)-1)].ID,
			UserID:       users[gofakeit.Number(0, len(users)-1)].ID,
			ReviewerName: gofakeit.Name(),
			Rating:       math.Round(gofakeit.Float64Range(0, 5)*100) / 100,
			Review:       gofakeit.Paragraph(1, 5, 10, " "),
			CreatedAt:    reviewDate,
			ReviewDate:   reviewDate.Format("01/02/2006"),
		}

		_, err = db.ReviewCollection.InsertOne(context.Background(), review)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	return
}

func GetFullProducts(filters ...bson.D) (products []Product, err error) {
	aggregrateSearchObject := bson.A{}
	aggregrateSearchObject = append(aggregrateSearchObject, bson.D{
		{Key: "$lookup",
			Value: bson.D{
				{Key: "from", Value: "brands"},
				{Key: "localField", Value: "brandID"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "brand"},
			},
		},
	},
		bson.D{
			{Key: "$unwind",
				Value: bson.D{
					{Key: "path", Value: "$brand"},
					{Key: "preserveNullAndEmptyArrays", Value: true},
				},
			},
		},
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "variants"},
					{Key: "localField", Value: "_id"},
					{Key: "foreignField", Value: "productID"},
					{Key: "as", Value: "variants"},
				},
			},
		})
	productsCursor, err := db.ProductCollection.Aggregate(context.Background(), aggregrateSearchObject)
	if err != nil {
		return
	}
	if err = productsCursor.All(context.TODO(), &products); err != nil {
		return
	}
	defer productsCursor.Close(context.Background())
	return
}

func (product Product) GetReviews(n int64) (reviews []Review, err error) {

	op := options.Find()
	op.Sort = bson.M{"rating": -1}
	op.Limit = &n

	reviewsCursor, err := db.ReviewCollection.Find(context.Background(), bson.M{"productID": product.ID, "review": bson.M{"$ne": ""}}, op)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer reviewsCursor.Close(context.Background())

	if err = reviewsCursor.All(context.TODO(), &reviews); err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	// for i := range reviews {
	// 	reviews[i].RatingVisualization, _ = GetRatingVisualizationObject(reviews[i].Rating)
	// }

	return

}

func (product *Product) AddReview(userID string, reviewerName string, rating float64, reviewStr string) (review Review, err error) {

	review, err = CreateNewReview(product.ID, userID, reviewerName, rating, reviewStr)
	if err != nil {
		return
	}

	err = product.AddRating(review)
	if err != nil {
		return
	}

	return
}

func (product *Product) CreateNewReview(userID string, reviewerName string, rating float64, reviewStr string) (review Review, err error) {
	review, err = CreateNewReview(product.ID, userID, reviewerName, rating, reviewStr)
	if err != nil {
		return
	}

	err = product.AddRating(review)
	return
}

func (product *Product) AddOrUpdateReview(userID string, reviewerName string, rating float64, reviewStr string) (review Review, err error) {

	// Get if User has already reiewed this product
	review, err = GetReviewForProductByUser(product.ID, userID)
	if err != nil {
		// If review is not found
		// Create Review
		if err == mongo.ErrNoDocuments {
			review, err = product.CreateNewReview(userID, reviewerName, rating, reviewStr)
			if err != nil {
				return
			}
		} else {
			return
		}
	}

	// If review was found
	// remove old rating from product
	err = product.RemoveRating(review)
	if err != nil {
		return
	}

	// Update Review
	// err = CreateNewReview(product.ID, userID, reviewerName, rating, reviewStr)
	err = review.UpdateRating(rating, reviewStr)
	if err != nil {
		return
	}

	// add new rating from product
	err = product.AddRating(review)
	if err != nil {
		return
	}

	return
}

func (product *Product) AddRating(review Review) (err error) {

	if review.ProductID != product.ID {
		return fmt.Errorf("review is not for this product")
	}

	product.AverageRating = ((product.AverageRating * float64(product.RatingCount)) + review.Rating) / float64(product.RatingCount+1)
	product.RatingCount = product.RatingCount + 1
	product.RatingVisualization, err = GetRatingVisualizationObject(product.AverageRating)
	if err != nil {
		return err
	}
	_, err = db.ProductCollection.ReplaceOne(context.Background(), bson.M{"_id": product.ID}, product)
	return
}

func (product *Product) RemoveRating(review Review) (err error) {

	if review.ProductID != product.ID {
		return fmt.Errorf("review is not for this product")
	}

	if product.RatingCount == 1 {
		product.AverageRating = 0
		product.RatingCount = 0
	} else {
		product.AverageRating = ((product.AverageRating * float64(product.RatingCount)) - review.Rating) / float64(product.RatingCount-1)
		fmt.Printf("product.AverageRating: %v\n", product.AverageRating)
		product.RatingCount = product.RatingCount - 1
	}

	product.RatingVisualization, err = GetRatingVisualizationObject(product.AverageRating)
	if err != nil {
		return err
	}
	_, err = db.ProductCollection.ReplaceOne(context.Background(), bson.M{"_id": product.ID}, product)

	return
}

// {
// 	Title:        "Number 1 Item",
// 	Review:       "Theres no better product for this in the market",
// 	ReviewerName: "Ishant Lassi",
// 	Rating:       3,
// 	RatingVisualization: models.RatingVisualization{
// 		FullStars:  make([]struct{}, 3),
// 		HalfStars:  make([]struct{}, 0),
// 		EmptyStars: make([]struct{}, 2),
// 	},
// },
// {
// 	Title:        "ek Number Item",
// 	ReviewerName: "Peter Ambre",
// 	Review:       "Best Cooker Ever. 10/10 would by again",
// 	Rating:       5,
// 	RatingVisualization: models.RatingVisualization{
// 		FullStars:  make([]struct{}, 5),
// 		HalfStars:  make([]struct{}, 0),
// 		EmptyStars: make([]struct{}, 0),
// 	},
// },
