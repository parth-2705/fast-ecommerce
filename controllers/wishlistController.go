package controllers

import (
	"fmt"
	"hermes/models"
	"hermes/utils/amplitude"
	"hermes/utils/data"
	"hermes/utils/network"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type WishlistPostBody struct {
	ProductID string `json:"productID" bson:"productID"`
}

func WishListGetHandler(c *gin.Context) {

	go amplitude.TrackEventByAuth("Wislist Page", c)

	user, _ := getUserObjectFromSession(c)

	wishlistID, err := models.GetWishlistFromUser(user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to get Wishlist"})
		return
	}

	products, err := models.GetAllProductsInWishlist(bson.D{{Key: "$match", Value: bson.D{{Key: "_wishlistID", Value: wishlistID}}}})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to get Wishlist products"})
		fmt.Println("error", err)
		return
	}
	back := c.Request.Referer()
	if back == "" {
		back = "/"
	}

	if network.MobileRequest(c) {
		c.JSON(http.StatusOK, gin.H{
			"wishlist": products,
		})

		return
	}

	c.HTML(200, "wishlist", gin.H{
		"wishlist": products,
		"back":     back,
	})
}

func WishListAddPostHandler(c *gin.Context) {

	go amplitude.TrackEventByAuth("Add Wishlist", c)

	var wishlistPostBody WishlistPostBody
	err := c.BindJSON(&wishlistPostBody)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
		return
	}
	user, err := getUserObjectFromSession(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get User"})
		data.SetSessionValue(c, "setProductToSession", wishlistPostBody.ProductID)
		return
	}
	err = addToWishlist(c, user, wishlistPostBody.ProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "Successfully added to Wishlist"})
}

func WishListRemovePostHandler(c *gin.Context) {

	go amplitude.TrackEventByAuth("Remove Wishlist", c)

	var wishlistPostBody WishlistPostBody
	err := c.BindJSON(&wishlistPostBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
		return
	}
	user, err := getUserObjectFromSession(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to get User"})
		data.SetSessionValue(c, "setProductToSession", wishlistPostBody.ProductID)
		return
	}
	err = removeFromWishlist(c, user, wishlistPostBody.ProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "Successfully added to Wishlist"})
}

func removeFromWishlist(c *gin.Context, user models.User, productID string) error {
	wishlistID, err := models.GetWishlistFromUser(user)
	if err != nil {
		return err
	}
	wishlistObject := models.WishlistObject{
		ProductId:  productID,
		WishlistID: wishlistID,
	}
	err = models.RemoveProductFromWishlist(wishlistObject)
	if err != nil {
		return err
	}

	return nil
}

func addToWishlist(c *gin.Context, user models.User, productID string) error {
	wishlistID, err := models.GetWishlistFromUser(user)
	if err != nil {
		return err
	}
	wishlistObject := models.WishlistObject{
		ProductId:  productID,
		WishlistID: wishlistID,
	}
	isWishlisted := models.GetIfItemIsInWishlistOfUser(user, productID)
	if isWishlisted {
		return nil
	}
	err = models.AddProductToWishlist(wishlistObject)
	if err != nil {
		return err
	}

	return nil
}
