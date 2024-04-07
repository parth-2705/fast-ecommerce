package models

import (
	"context"
	"fmt"
	"hermes/db"
	"hermes/utils/data"
	"math"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RatingVisualization struct {
	FullStars  []struct{}
	HalfStars  []struct{}
	EmptyStars []struct{}
}

type Review struct {
	ID                  string              `json:"id" bson:"_id"`
	ProductID           string              `json:"productId" bson:"productID"`
	UserID              string              `json:"userID" bson:"userID"`
	ReviewerName        string              `json:"reviewerName" bson:"reviewerName"`
	Rating              float64             `json:"rating" bson:"rating"`
	RatingVisualization RatingVisualization `json:"ratingVisualization" bson:"ratingVisualization"`
	Review              string              `json:"review" bson:"review"`
	CreatedAt           time.Time           `json:"createdAt" bson:"createdAt"`
	UpdatedAt           time.Time           `json:"updatedAt" bson:"updatedAt"`
	ReviewDate          string              `json:"reviewDate" bson:"reviewDate"`
}

func (review Review) CreateIndexes() error {

	userIDIndex := mongo.IndexModel{
		Keys: bson.M{
			"userID": 1, // 1 for ascending, -1 for descending
		},
	}

	productIDIndex := mongo.IndexModel{
		Keys: bson.M{
			"productID": 1, // 1 for ascending, -1 for descending
		},
	}

	op := options.Index()
	t := true
	op.Unique = &t
	productIDUserIdIndex := mongo.IndexModel{
		Keys: bson.D{
			{"productID", 1},
			{"userID",    1},
		},
		Options: op,
	}

	indexModels := []mongo.IndexModel{userIDIndex, productIDIndex, productIDUserIdIndex}

	_, err := db.ReviewCollection.Indexes().CreateMany(context.Background(), indexModels)
	if err != nil {
		// handle error
		if strings.Contains(err.Error(), "Index with name") && strings.Contains(err.Error(), "already exists") {
			fmt.Println("Indexes already exist")
		} else {
			// Handle other errors
			fmt.Println("Error creating indexes:", err)
			return err
		}
	}
	return nil
}

func CreateNewReview(productID string, userID string, reviewerName string, rating float64, reviewStr string) (review Review, err error) {

	if rating > 5 || rating < 0 {
		return review, fmt.Errorf("invalid Rating Value")
	}

	review = Review{
		ID:           data.GetUUIDString("review"),
		ProductID:    productID,
		UserID:       userID,
		ReviewerName: reviewerName,
		Rating:       rating,
		Review:       reviewStr,

		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		ReviewDate: time.Now().Format("2006-01-02"),
	}

	review.GetRatingVisualizationObject()
	_, err = db.ReviewCollection.InsertOne(context.Background(), review)
	if err != nil {
		fmt.Printf("save review err: %v\n", err)
		return
	}

	return
}

func (review *Review) GetRatingVisualizationObject() {
	var fullStarCount, halfStarCount, emptyStarCount int

	fullStarCount = int(math.Trunc(review.Rating))
	if math.Trunc(review.Rating) != review.Rating {
		halfStarCount = 1
	}
	emptyStarCount = 5 - halfStarCount - fullStarCount

	visualizer := RatingVisualization{
		FullStars:  make([]struct{}, fullStarCount),
		HalfStars:  make([]struct{}, halfStarCount),
		EmptyStars: make([]struct{}, emptyStarCount),
	}

	review.RatingVisualization = visualizer

}

func GetRatingVisualizationObject(rating float64) (visualizer RatingVisualization, err error) {
	if rating > 5 || rating < 0 {
		return visualizer, fmt.Errorf("invalid Rating Value")
	}

	var fullStarCount, halfStarCount, emptyStarCount int

	fullStarCount = int(math.Trunc(float64(rating)))
	if math.Trunc(float64(rating)) != float64(rating) {
		halfStarCount = 1
	}
	emptyStarCount = 5 - halfStarCount - fullStarCount

	visualizer = RatingVisualization{
		FullStars:  make([]struct{}, fullStarCount),
		HalfStars:  make([]struct{}, halfStarCount),
		EmptyStars: make([]struct{}, emptyStarCount),
	}

	return
}

func GetAllReviewsForProduct(productID string) (reviewArr []Review, err error) {
	cur, err := db.ReviewCollection.Find(context.Background(), bson.M{"productID": productID})
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())
	err = cur.All(context.Background(), &reviewArr)
	if err != nil {
		return nil, err
	}
	return reviewArr, nil
}

func PushReviewToDB(review Review) (err error) {
	review.ID = data.GetUUIDString("review")
	review.CreatedAt = time.Now()
	review.UpdatedAt = time.Now()
	review.ReviewDate = time.Now().Format("2006-01-02")
	review.RatingVisualization, err = GetRatingVisualizationObject(review.Rating)
	_, err = db.ReviewCollection.InsertOne(context.Background(), review)
	return err
}

func GetReviewForProductNotFromUser(productID string, userID string) (reviewArr []Review, err error) {
	filter := bson.D{
		{"$and",
			bson.A{
				bson.D{{"productID", bson.D{{"$eq", productID}}}},
				bson.D{{"userID", bson.D{{"$not", bson.D{{"$eq", userID}}}}}},
			},
		},
	}
	cur, err := db.ReviewCollection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	err = cur.All(context.Background(), &reviewArr)
	if err != nil {
		return nil, err
	}
	return reviewArr, nil
}

func GetReviewForProductFromUser(productID string, userID string) (reviewArr []Review, err error) {
	filter := bson.D{
		{"$and",
			bson.A{
				bson.D{{"productID", bson.D{{"$eq", productID}}}},
				bson.D{{"userID", bson.D{{"$eq", userID}}}},
			},
		},
	}
	cur, err := db.ReviewCollection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	err = cur.All(context.Background(), &reviewArr)
	if err != nil {
		return nil, err
	}
	return reviewArr, nil
}

func CanAddReview(productID string, user User) (order Order, canReview bool) {
	if user.ID == "" || productID == "" {
		return order, false
	}
	filter := bson.D{
		{"$and",
			bson.A{
				bson.D{{"userId", bson.D{{"$eq", user.ID}}}},
				bson.D{{"product._id", bson.D{{"$eq", productID}}}},
			},
		},
	}
	err := db.OrderCollection.FindOne(context.Background(), filter).Decode(&order)
	if err != nil {
		return order, false
	}
	return order, true
}

func GetReviewForProductByUser(productID string, userID string) (review Review, err error) {
	res := db.ReviewCollection.FindOne(context.Background(), bson.M{"productID": productID, "userID": userID})
	err = res.Decode(&review)
	if err != nil {
		fmt.Printf("Error Getting Review:  %v\n", err)
	}

	return
}

func (review *Review) UpdateRating(rating float64, reviewStr string) (err error) {

	review.Rating = float64(rating)
	review.Review = reviewStr

	review.GetRatingVisualizationObject()
	err = review.UpdateReviewObj()
	return
}

func (review *Review) UpdateReviewObj() (err error) {
	review.UpdatedAt = time.Now()
	review.ReviewDate = time.Now().Format("2006-01-02")
	_, err = db.ReviewCollection.UpdateOne(context.Background(), bson.M{"_id": review.ID}, bson.M{"$set": review})
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}
