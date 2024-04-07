package controllers

import (
	"context"
	"fmt"
	"hermes/admin/services/auth"
	"hermes/db"
	"hermes/models"
	"hermes/utils/data"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func AllBrandsPage(c *gin.Context) {

	seller, err := auth.GetSellerFromSession(c)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	var brands []models.Brand
	cur, err := db.BrandsCollection.Find(context.Background(), bson.M{"sellerID": seller.ID})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Brand not found" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}
	err = cur.All(context.Background(), &brands)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Brand not correct format" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}
	c.HTML(http.StatusOK, "seller-root", gin.H{"title": "Roovo", "brands": brands, "template": "brand"})
}

func NewBrandPage(c *gin.Context) {
	var brand models.Brand
	c.HTML(http.StatusOK, "seller-root", gin.H{"title": "Roovo", "brand": brand, "template": "brand-edit", "new": true})
}

func BrandEditPage(c *gin.Context) {
	brandId := c.Param("brandId")
	var brand models.Brand
	err := db.BrandsCollection.FindOne(context.Background(), bson.M{"_id": brandId}).Decode(&brand)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}
	c.HTML(http.StatusOK, "seller-root", gin.H{"title": "Roovo", "brand": brand, "template": "brand-edit", "new": false})
}

func CreateBrand(c *gin.Context) {
	seller, err := auth.GetSellerFromSession(c)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	newBrand := models.Brand{}
	err = c.BindJSON(&newBrand)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
		return
	}
	newBrand.ID = data.GetUUIDString("brand")
	newBrand.SellerID = seller.ID
	_, err = db.BrandsCollection.InsertOne(context.Background(), newBrand)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to update " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "Successfully created brand"})
}

func UpdateBrand(c *gin.Context) {
	brandId := c.Query("id")

	seller, err := auth.GetSellerFromSession(c)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var data models.Brand
	err = c.BindJSON(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
		return
	}

	data.SellerID = seller.ID
	result, err := db.BrandsCollection.ReplaceOne(context.Background(), bson.M{"_id": brandId}, data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to update " + err.Error()})
		return
	}
	if result.ModifiedCount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not find brand with given ID"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "Successfully updated brand with given ID"})
}

func DeleteBrand(c *gin.Context) {
	categoryId := c.Query("id")
	result, err := db.BrandsCollection.DeleteOne(context.Background(), bson.M{"_id": categoryId})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to delete " + err.Error()})
		return
	}
	if result.DeletedCount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not find brand with given ID"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "Successfully deleted brand with given ID"})
}
