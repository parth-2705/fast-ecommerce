package controllers

import (
	"context"
	"fmt"
	"hermes/db"
	"hermes/models"
	"hermes/utils/data"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func AllBrandsPage(c *gin.Context) {
	var brands []models.Brand
	cur, err := db.BrandsCollection.Find(context.Background(), bson.D{})
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
	c.HTML(http.StatusOK, "root", gin.H{"title": "Admin | Roovo", "brands": brands, "template": "brand"})
}

func NewBrandPage(c *gin.Context) {
	var brand models.Brand
	c.HTML(http.StatusOK, "root", gin.H{"title": "Admin | Roovo", "brand": brand, "template": "brand-edit", "new": true})
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
	c.HTML(http.StatusOK, "root", gin.H{"title": "Admin | Roovo", "brand": brand, "template": "brand-edit", "new": false})
}

func CreateBrand(c *gin.Context) {

	var newBrand models.Brand
	err := c.BindJSON(&newBrand)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
		return
	}

	newBrand.ID = data.GetUUIDString("brand")
	err = newBrand.Create()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": "Successfully created brand"})
}

func UpdateBrand(c *gin.Context) {
	var updatedBrand models.Brand
	err := c.BindJSON(&updatedBrand)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
		return
	}
	result, err := updatedBrand.Update()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if result.MatchedCount == 0 {
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

func GetBrandsList(c *gin.Context) {
	brands, err := models.GetBrandsList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"brands": brands,
	})
}

func GetBrandByID(c *gin.Context) {

	brandID := c.Param("brandID")

	brand, err := models.GetBrand(brandID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"brand": brand,
	})
}

func GetListOfStates(c *gin.Context) {
	states, err := models.GetListOfAllStates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"states": states,
	})
}
