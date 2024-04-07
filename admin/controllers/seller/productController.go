package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"hermes/admin/controllers/common"
	"hermes/admin/services/auth"
	"hermes/db"
	"hermes/models"
	"hermes/utils/data"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

var productStatusList = []string{"Active", "Draft"}

func AllProductsPage(c *gin.Context) {

	seller, err := auth.GetSellerFromSession(c)
	if err != nil {
		c.AbortWithError(http.StatusBadGateway, err)
		return
	}

	if !seller.ProfileCompleted {
		c.Redirect(http.StatusFound, "/info")
		return
	}

	var products []models.Product
	var data []map[string]interface{}
	cur, err := db.ProductCollection.Find(context.Background(), bson.M{"sellerID": seller.ID})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found " + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}
	err = cur.All(context.Background(), &products)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not correct format " + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}
	marshalledProducts, _ := json.Marshal(products)
	json.Unmarshal(marshalledProducts, &data)
	for idx, val := range products {
		var currCategory models.Category
		var currBrand models.Brand
		var variants []models.Variation
		temp := make(map[string]int)
		_ = db.CategoryCollection.FindOne(context.Background(), bson.M{"_id": val.Category}).Decode(&currCategory)
		data[idx]["category"] = currCategory
		_ = db.BrandsCollection.FindOne(context.Background(), bson.M{"_id": val.BrandID}).Decode(&currBrand)
		data[idx]["brand"] = currBrand

		cur, err := db.VariantsCollection.Find(context.Background(), bson.M{"productID": val.ID})
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Variant not found " + err.Error()})
			fmt.Printf("err: %v\n", err)
			return
		}
		err = cur.All(context.Background(), &variants)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Variant not correct format " + err.Error()})
			fmt.Printf("err: %v\n", err)
			return
		}
		if len(variants) > 0 {
			data[idx]["thumbnail"] = variants[0].Thumbnail
			data[idx]["showVariants"] = true
		} else {
			data[idx]["showVariants"] = false
		}
		temp["number"] = len(variants)
		temp["val"] = 0
		for _, variantVal := range variants {
			temp["val"] = temp["val"] + variantVal.Quantity
		}
		data[idx]["quantity"] = temp
	}

	var brands []models.Brand
	cur, err = db.BrandsCollection.Find(context.Background(), bson.M{"sellerID": seller.ID})
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

	c.HTML(http.StatusOK, "seller-root", gin.H{"title": "Roovo", "products": data, "brands": brands, "template": "product2"})
}

func NewProductsPage(c *gin.Context) {

	seller, err := auth.GetSellerFromSession(c)
	if err != nil {
		c.AbortWithError(http.StatusBadGateway, err)
		return
	}

	if !seller.ProfileCompleted {
		c.Redirect(http.StatusFound, "/info")
		return
	}

	var product models.Product
	product.Attributes = []models.VariationAttribute{}
	var category models.Category

	var categories []models.Category
	cur, err := db.CategoryCollection.Find(context.Background(), bson.D{})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}
	err = cur.All(context.Background(), &categories)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not correct format" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}

	var brands []models.Brand
	cur, err = db.BrandsCollection.Find(context.Background(), bson.M{"sellerID": seller.ID})
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

	var variations []models.Variation
	var variation models.Variation
	variation.Attributes = make(map[string]string)
	variation.Images = []models.Image{}
	variations = append(variations, variation)
	c.HTML(http.StatusOK, "seller-root", gin.H{"title": "Roovo", "product": product, "category": category, "categoryList": categories, "brandList": brands, "seller": seller, "variants": variations, "template": "product-edit2", "new": true, "productStatusList": productStatusList})
}

func ProductEditPage(c *gin.Context) {

	seller, err := auth.GetSellerFromSession(c)
	if err != nil {
		c.AbortWithError(http.StatusBadGateway, err)
		return
	}

	if !seller.ProfileCompleted {
		c.Redirect(http.StatusFound, "/info")
		return
	}

	productId := c.Param("productId")
	var product models.Product
	err = db.ProductCollection.FindOne(context.Background(), bson.M{"_id": productId}).Decode(&product)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}

	var categories []models.Category
	cur, err := db.CategoryCollection.Find(context.Background(), bson.D{})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}
	err = cur.All(context.Background(), &categories)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not correct format" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}

	var brands []models.Brand
	cur, err = db.BrandsCollection.Find(context.Background(), bson.M{"sellerID": seller.ID})
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

	var category models.Category
	err = db.CategoryCollection.FindOne(context.Background(), bson.M{"_id": product.Category}).Decode(&category)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product Category not found" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}

	var brand models.Brand
	err = db.BrandsCollection.FindOne(context.Background(), bson.M{"_id": product.BrandID}).Decode(&brand)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product Brand not found" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}

	var variants []models.Variation
	fmt.Println("product:", product.ID)
	cur, err = db.VariantsCollection.Find(context.Background(), bson.M{"productID": product.ID})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Variant not found " + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}
	err = cur.All(context.Background(), &variants)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Variant not correct format " + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}
	c.HTML(http.StatusOK, "seller-root", gin.H{"title": "Roovo", "product": product, "category": category, "brand": brand, "categoryList": categories, "brandList": brands, "seller": seller, "template": "product-edit2", "variants": variants, "new": false, "productStatusList": productStatusList})
}

func AddProduct(c *gin.Context) {
	newProduct := models.Product{}
	err := c.BindJSON(&newProduct)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
		return
	}
	newProduct.ID = data.GetUUIDString("product")
	_, err = db.ProductCollection.InsertOne(context.Background(), newProduct)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to update " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "Successfully created product", "id": newProduct.ID})
}

func UpdateProduct(c *gin.Context) {
	productId := c.Query("id")
	var data models.Product
	err := c.BindJSON(&data)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body:" + err.Error()})
		return
	}

	result, err := db.ProductCollection.ReplaceOne(context.Background(), bson.M{"_id": productId}, data)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to update " + err.Error()})
		return
	}
	if result.ModifiedCount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not find product with given ID"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "Successfully updated product with given ID"})
}

func DeleteProduct(c *gin.Context) {
	productId := c.Query("id")

	result, err := db.ProductCollection.DeleteOne(context.Background(), bson.M{"_id": productId})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to delete " + err.Error()})
		return
	}
	if result.DeletedCount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not delete product with given ID"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "Successfully deleted product with given ID"})
}

func DuplicateProduct(c *gin.Context) {
	productId := c.Query("id")
	var product models.Product

	err := db.ProductCollection.FindOne(context.Background(), bson.M{"_id": productId}).Decode(&product)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to find the product " + err.Error()})
		return
	}

	product.ID = data.GetUUIDString("product")
	_, err = db.ProductCollection.InsertOne(context.Background(), product)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to insert new product " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": "Successfully duplicated product with given ID", "id": product.ID})
}

func GetProductListForSeller(c *gin.Context) {
	seller, err := auth.GetSellerFromSession(c)
	if err != nil {
		c.AbortWithError(http.StatusBadGateway, err)
		return
	}

	if !seller.ProfileCompleted {
		c.JSON(http.StatusTemporaryRedirect, gin.H{"redirect": "/info"})
		return
	}

	searchTerm := c.Query("searchTerm")
	brandID := c.Query("brandID")
	categoryID := c.Query("categoryID")

	filters := bson.M{"sellerID": seller.ID}

	if brandID != "" {
		filters["brandID"] = brandID
	}
	if categoryID != "" {
		filters["category"] = categoryID
	}

	var categoryMap map[string]struct{} = make(map[string]struct{})
	var categoryArr []models.Category
	var data []map[string]interface{}

	products, err := common.FilterProduct(searchTerm, filters)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Error in filtering " + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}

	marshalledProducts, _ := json.Marshal(products)
	json.Unmarshal(marshalledProducts, &data)

	for idx, val := range products {
		var currCategory models.Category
		var currBrand models.Brand
		var variants []models.Variation
		temp := make(map[string]int)
		_ = db.CategoryCollection.FindOne(context.Background(), bson.M{"_id": val.Category}).Decode(&currCategory)
		if _, ok := categoryMap[val.Category]; !ok {
			categoryMap[val.Category] = struct{}{}
			categoryArr = append(categoryArr, currCategory)
		}
		data[idx]["category"] = currCategory
		_ = db.BrandsCollection.FindOne(context.Background(), bson.M{"_id": val.BrandID}).Decode(&currBrand)
		data[idx]["brand"] = currBrand

		cur, err := db.VariantsCollection.Find(context.Background(), bson.M{"productID": val.ID})
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Variant not found " + err.Error()})
			fmt.Printf("err: %v\n", err)
			return
		}
		err = cur.All(context.Background(), &variants)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Variant not correct format " + err.Error()})
			fmt.Printf("err: %v\n", err)
			return
		}
		if len(variants) > 0 {
			data[idx]["thumbnail"] = variants[0].Thumbnail
			data[idx]["showVariants"] = true
		} else {
			data[idx]["showVariants"] = false
		}
		temp["number"] = len(variants)
		temp["val"] = 0
		for _, variantVal := range variants {
			temp["val"] = temp["val"] + variantVal.Quantity
		}
		data[idx]["quantity"] = temp
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

	c.JSON(http.StatusOK, gin.H{"products": data, "brands": brands, "categories": categoryArr})
}
