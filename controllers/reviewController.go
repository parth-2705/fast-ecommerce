package controllers

import (
	"context"
	"fmt"
	"hermes/configs/Redis"
	"hermes/db"
	"hermes/models"
	"hermes/utils/rw"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func ReviewEntryPage(c *gin.Context) {
	productID := c.Param("productID")
	product, err := models.GetCompleteProduct(productID)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	user, err := getUserObjectFromSession(c)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	order, canAddReview := models.CanAddReview(productID, user)
	if !canAddReview {
		c.Redirect(http.StatusTemporaryRedirect, "/product/"+productID)
		return
	}
	c.HTML(http.StatusOK, "enter-review", gin.H{
		"product": product,
		"user":    user,
		"order":   order,
	})
}

func GetReviewsForProduct(c *gin.Context) {
	productID := c.Query("productID")
	reviewArr, err := models.GetAllReviewsForProduct(productID)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"reviews": reviewArr})
}

func ReviewPostHandler2(c *gin.Context) {

	var review models.Review
	user, err := getUserObjectFromSession(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get user from session" + err.Error()})
		return
	}
	productID := c.Query("productID")
	err = c.BindJSON(&review)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to bind request body " + err.Error()})
		return
	}

	product, err := models.GetProduct(productID)
	if err != nil {
		rw.JSONErrorResponse(c, 404, fmt.Errorf("product not found: "+err.Error()))
		return
	}

	review, err = product.AddOrUpdateReview(user.ID, review.ReviewerName, float64(review.Rating), review.Review)
	if err != nil {
		rw.JSONErrorResponse(c, 500, err)
		return
	}

	err = Redis.DeleteProductCacheByID(product.ID)
	if err != nil {
		rw.JSONErrorResponse(c, 500, err)
		return
	}
}

func ReviewPostHandler(c *gin.Context) {
	var review models.Review
	user, err := getUserObjectFromSession(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get user from session" + err.Error()})
		return
	}
	productID := c.Query("productID")
	err = c.BindJSON(&review)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to bind request body " + err.Error()})
		return
	}
	review.UserID = user.ID
	review.ProductID = productID
	err = Redis.DeleteProductCacheByID(productID)
	if err != nil {
		fmt.Println(fmt.Errorf("Unable to remove review"))
	}
	err = user.UpdateUserName(review.ReviewerName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to add user name " + err.Error()})
		return
	}
	err = models.PushReviewToDB(review)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to push review to DB " + err.Error()})
		return
	}
	err = AddRatingForProduct(productID, review.Rating)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update product rating " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "success"})
}

func AddRatingForProduct(productID string, ratingToAdd float64) (err error) {
	var product models.Product
	err = db.ProductCollection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&product)
	if err != nil {
		return err
	}
	copy := product
	copy.AverageRating = ((product.AverageRating * float64(product.RatingCount)) + ratingToAdd) / float64(product.RatingCount+1)
	copy.RatingCount = product.RatingCount + 1
	copy.RatingVisualization, err = models.GetRatingVisualizationObject(copy.AverageRating)
	if err != nil {
		return err
	}
	_, err = db.ProductCollection.ReplaceOne(context.Background(), bson.M{"_id": productID}, copy)
	return
}
