package controllers

import (
	"context"
	"hermes/configs/Redis"
	"hermes/db"
	"hermes/models"
	"hermes/utils/data"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type CSV struct {
	Index int    `json:"index"`
	Value string `json:"value"`
}

func GetProductVariant(c *gin.Context) {

	productID := c.Query("productID")

	var variant models.Variation
	err := db.VariantsCollection.FindOne(context.Background(), bson.M{"productID": productID}).Decode(&variant)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to get variants" + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"variant": variant})
}

func GetAllProductVariants(c *gin.Context) {
	productID := c.Query("productID")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty Product ID"})
		return
	}
	var variants []models.Variation
	cur, err := db.VariantsCollection.Find(context.Background(), bson.M{"productID": productID})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to get variants" + err.Error()})
		return
	}
	err = cur.All(context.Background(), &variants)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to get variants" + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"variants": variants})
}

func GetVariants(c *gin.Context) {
	var variants []models.Variation
	cur, err := db.VariantsCollection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to get variants"})
		return
	}
	err = cur.All(context.Background(), &variants)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to get variants"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"testing": variants})
}

func BatchDeleteVariants(c *gin.Context) {
	productId := c.Query("productId")
	_, err := db.VariantsCollection.DeleteMany(context.Background(), bson.M{"productID": productId})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to delete variants"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Succefully Deleted variants for product ID"})
}

func BatchAddVariants(c *gin.Context) {
	var variantArr []models.Variation
	err := c.BindJSON(&variantArr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body " + err.Error()})
		return
	}
	for idx := range variantArr {
		variantArr[idx].ID = data.GetUUIDString("variant")
	}
	// var finalArr []interface{}
	// temp, _ := json.Marshal(variantArr)
	// _ = json.Unmarshal(temp, &finalArr)
	for _, variant := range variantArr {
		_, err = db.VariantsCollection.InsertOne(context.Background(), variant)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to add variants " + err.Error()})
			return
		}
	}
	err = Redis.DeleteProductCacheByID(variantArr[0].ProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to delete redis cache " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Succefully Deleted variants for product ID"})
}

func BatchReplaceVariants(c *gin.Context) {
	var variantArr []models.Variation
	err := c.BindJSON(&variantArr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body " + err.Error()})
		return
	}
	// for idx := range variantArr {
	// 	variantArr[idx].ID = data.GetUUIDString("variant")
	// }
	// var finalArr []interface{}
	// temp, _ := json.Marshal(variantArr)
	// _ = json.Unmarshal(temp, &finalArr)
	for _, variant := range variantArr {
		_, err = db.VariantsCollection.ReplaceOne(context.Background(), bson.M{"_id": variant.ID}, variant)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to add variants " + err.Error()})
			return
		}
	}
	err = Redis.DeleteProductCacheByID(variantArr[0].ProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to delete redis cache " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Succefully replaced variants for product ID"})
}

func EditVariantDetails(c *gin.Context) {

	variantID := c.Query("variantID")
	varaint, err := models.GetVariant(variantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to get variant " + err.Error()})
		return
	}

	if err := c.ShouldBind(&varaint); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body " + err.Error()})
		return
	}

	err = varaint.Update()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to update variant " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"variant": varaint})
}
