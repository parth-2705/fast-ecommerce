package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"hermes/admin/controllers/common"
	"hermes/configs/Redis"
	"hermes/controllers"
	"hermes/db"
	"hermes/models"
	"hermes/scripts"
	"hermes/search"
	"hermes/utils/data"
	utils "hermes/utils/queries"
	"hermes/utils/rw"
	"hermes/utils/whatsapp"

	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

var productStatusList = []string{"Active", "Draft"}

func AllProductsPage(c *gin.Context) {
	var products []models.Product
	var data []map[string]interface{}
	cur, err := db.ProductCollection.Find(context.Background(), bson.D{})
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
		_ = db.CategoryCollection.FindOne(context.Background(), bson.M{"_id": val.Category}).Decode(&currCategory)
		data[idx]["category"] = currCategory
		_ = db.BrandsCollection.FindOne(context.Background(), bson.M{"_id": val.BrandID}).Decode(&currBrand)
		data[idx]["brand"] = currBrand
	}
	c.HTML(http.StatusOK, "root", gin.H{"title": "Admin | Roovo", "products": data, "template": "product"})
}

func AllProductsPage2(c *gin.Context) {
	var products []models.Product
	var data []map[string]interface{}
	cur, err := db.ProductCollection.Find(context.Background(), bson.D{})
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

	limit := c.Query("limit")
	limitInt := 20
	if limit != "" {
		limitInt, err = strconv.Atoi(limit)
		if err != nil {
			limitInt = 20
		}
	}

	page := c.Query("page")
	pageInt := 1
	if page != "" {
		pageInt, err = strconv.Atoi(page)
		if err != nil {
			pageInt = 1
		}
	}

	var Paginater controllers.Pagination = controllers.Pagination{
		Limit: limitInt,
		Page:  pageInt,
	}

	var productsStruct []*models.Product
	productsPaginator, err := controllers.Paginate("product", &Paginater, productsStruct, bson.A{}, utils.ProductQuery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error in pagination": err.Error()})
		return
	}

	marshalledProducts, err := json.Marshal(productsPaginator.Rows)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error in Marshal": err.Error()})
		return
	}
	err = json.Unmarshal(marshalledProducts, &data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error in Unmarshal": err.Error()})
		return
	}
	// fmt.Println(len(products))
	for idx, val := range products {
		fmt.Println(idx, val.Category)
		var currCategory models.Category
		var currBrand models.Brand
		var variants []models.Variation
		temp := make(map[string]int)
		err = db.CategoryCollection.FindOne(context.Background(), bson.M{"_id": val.Category}).Decode(&currCategory)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error in CategoryCollection ": err.Error()})
			return
		}
		data[idx]["category"] = currCategory

		err = db.BrandsCollection.FindOne(context.Background(), bson.M{"_id": val.BrandID}).Decode(&currBrand)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error in BrandsCollection ": err.Error()})
			return
		}

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
		fmt.Println("end")
	}
	c.HTML(http.StatusOK, "root", gin.H{"title": "Admin | Roovo", "products": data, "template": "product2", "paginater": Paginater})
}

func GetProductPage(c *gin.Context) {
	var products []models.Product
	var data []map[string]interface{}
	var err error
	limit := c.Query("limit")
	limitInt := 20
	if limit != "" {
		limitInt, err = strconv.Atoi(limit)
	}
	limitInt = int(math.Min(float64(limitInt), 20))

	page := c.Query("page")
	pageInt := 1
	if page != "" {
		pageInt, err = strconv.Atoi(page)
	}

	var Paginater controllers.Pagination = controllers.Pagination{
		Limit: limitInt,
		Page:  pageInt,
	}

	products, err = controllers.ProductPaginate(&Paginater)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get products " + err.Error()})
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

	c.JSON(http.StatusOK, gin.H{"products": data})
	return
}

func NewProductsPage(c *gin.Context) {
	var product models.Product
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
	cur, err = db.BrandsCollection.Find(context.Background(), bson.D{})
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

	c.HTML(http.StatusOK, "root", gin.H{"title": "Admin | Roovo", "product": product, "category": category, "categoryList": categories, "brandList": brands, "template": "product-edit", "new": true, "productStatusList": productStatusList})
}

func NewProductsPage2(c *gin.Context) {
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
	cur, err = db.BrandsCollection.Find(context.Background(), bson.D{})
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
	c.HTML(http.StatusOK, "root", gin.H{"title": "Admin | Roovo", "product": product, "category": category, "categoryList": categories, "brandList": brands, "variants": variations, "template": "product-edit2", "new": true, "productStatusList": productStatusList})
}

func ProductEditPage(c *gin.Context) {

	productId := c.Param("productId")
	var product models.Product
	err := db.ProductCollection.FindOne(context.Background(), bson.M{"_id": productId}).Decode(&product)
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
	cur, err = db.BrandsCollection.Find(context.Background(), bson.D{})
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

	c.HTML(http.StatusOK, "root", gin.H{"title": "Admin | Roovo", "product": product, "category": category, "brand": brand, "categoryList": categories, "brandList": brands, "template": "product-edit", "new": false, "productStatusList": productStatusList})
}

func ProductEditPage2(c *gin.Context) {
	productId := c.Param("productId")
	var product models.Product
	err := db.ProductCollection.FindOne(context.Background(), bson.M{"_id": productId}).Decode(&product)
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
	cur, err = db.BrandsCollection.Find(context.Background(), bson.D{})
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
	c.HTML(http.StatusOK, "root", gin.H{"title": "Admin | Roovo", "product": product, "category": category, "brand": brand, "categoryList": categories, "brandList": brands, "template": "product-edit2", "variants": variants, "new": false, "productStatusList": productStatusList})
}

func AddProduct(c *gin.Context) {
	newProduct := models.Product{}
	err := c.BindJSON(&newProduct)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
		return
	}
	newProduct.ID = data.GetUUIDString("product")

	newProduct.PaymentMethods = make(models.PaymentMethodMap)
	for _, po := range models.CopyPaymentOptions() {
		newProduct.PaymentMethods[po.ID] = models.PaymentMethodConfiguration{Available: true}
	}

	_, err = db.ProductCollection.InsertOne(context.Background(), newProduct)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to create product " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "Successfully created product", "id": newProduct.ID})
}

func UpdateProduct(c *gin.Context) {
	var data models.Product
	err := c.BindJSON(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body:" + err.Error()})
		return
	}
	result, err := db.ProductCollection.ReplaceOne(context.Background(), bson.M{"_id": data.ID}, data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to update " + err.Error()})
		return
	}
	if result.ModifiedCount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not find product with given ID: " + data.ID})
		return
	}
	fmt.Println("data:", data.ID, data)
	err = Redis.DeleteProductCacheByID(data.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to update in Redis " + err.Error()})
		return
	}
	err = scripts.AddOrReplaceProduct(data.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to update in Meilisearch " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "Successfully updated product with given ID", "id": data.ID})
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

func FilterProductByName(c *gin.Context) {

	searchTerm := c.Query("q")
	products, err := common.FilterProduct(searchTerm, bson.M{})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to find products " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"products": products})

}

func GetProductByID(c *gin.Context) {

	productID := c.Query("productID")
	var product models.Product

	if len(productID) == 0 {
		rw.JSONErrorResponse(c, 400, fmt.Errorf("empty Product ID"))
		return
	}

	product, err := models.GetProduct(productID)
	if err != nil {
		rw.JSONErrorResponse(c, 400, err)
		return
	}

	c.JSON(http.StatusOK, product)
}

func GetProductList(c *gin.Context) {
	searchTerm := c.Query("searchTerm")
	brandID := c.Query("brandID")
	categoryID := c.Query("categoryID")

	limitInt := controllers.GetLimitFromQueryValue(c.Query("limit"))
	pageInt := controllers.GetPageFromQueryValue(c.Query("page"))

	var Paginater controllers.Pagination = controllers.Pagination{
		Limit: limitInt,
		Page:  pageInt,
	}

	var productsStruct []*models.Product

	preSkip := bson.A{}

	if len(searchTerm) != 0 {
		preSkip = append(preSkip, bson.D{{Key: "$match", Value: bson.D{{Key: "$text", Value: bson.D{{Key: "$search", Value: searchTerm}, {Key: "$caseSensitive", Value: false}}}}}})
		preSkip = append(preSkip, bson.D{{Key: "$sort", Value: bson.D{{Key: "score", Value: bson.D{{Key: "$meta", Value: "textScore"}}}}}})
	} else {
		preSkip = append(preSkip, bson.D{{Key: "$sort", Value: bson.D{{Key: "updatedAt", Value: -1}}}})
	}
	if len(brandID) != 0 {
		preSkip = append(preSkip, bson.D{{Key: "$match", Value: bson.D{{Key: "brandID", Value: brandID}}}})
	}
	if len(categoryID) != 0 {
		categoryIDArr, err := controllers.GetChildrenCategoryIDArr(categoryID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Error fetching children categories " + err.Error()})
			return
		}
		categoryFilter := []string{categoryID}
		categoryFilter = append(categoryFilter, categoryIDArr...)
		preSkip = append(preSkip, bson.D{{Key: "$match", Value: bson.D{{Key: "newCategory", Value: bson.D{{Key: "$in", Value: categoryFilter}}}}}})
	}

	postSkip := bson.A{}
	postSkip = append(postSkip, bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "variants"}, {Key: "localField", Value: "_id"}, {Key: "foreignField", Value: "productID"}, {Key: "as", Value: "variants"}}}})

	products, err := controllers.Paginate("product", &Paginater, productsStruct, preSkip, postSkip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	var brands []models.Brand
	cur, err := db.BrandsCollection.Find(context.Background(), bson.M{})
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

	var categories []models.Category
	cur, err = db.CategoryCollection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Categories not found" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}
	err = cur.All(context.Background(), &categories)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Categories not correct format" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products":   products.Rows,
		"totalRows":  products.TotalRows,
		"totalPages": products.TotalPages,
		"brands":     brands,
		"categories": categories,
	})
}

func GetProductList2(c *gin.Context) {
	searchTerm := c.Query("searchTerm")
	brandID := c.Query("brandID")

	limitInt := controllers.GetLimitFromQueryValue(c.Query("limit"))
	pageInt := controllers.GetPageFromQueryValue(c.Query("page"))

	var Paginater controllers.Pagination = controllers.Pagination{
		Limit: limitInt,
		Page:  pageInt,
	}

	processedAppliedFilterMap := map[string]search.FilterObject{}

	if brandID != "" {
		processedAppliedFilterMap["brandID"] = search.FilterObject{
			Values:   []string{brandID},
			Operator: "=",
			Path:     "brandID",
		}
	}

	sortArr := []search.SortObject{}
	productsStruct, err := controllers.ProductPaginate2(&Paginater, searchTerm, processedAppliedFilterMap, sortArr)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var brands []models.Brand
	cur, err := db.BrandsCollection.Find(context.Background(), bson.M{})
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

	var categories []models.Category
	cur, err = db.CategoryCollection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Categories not found" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}
	err = cur.All(context.Background(), &categories)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Categories not correct format" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products":   productsStruct,
		"totalRows":  Paginater.TotalRows,
		"totalPages": Paginater.TotalPages,
	})
}

func AddReviewForProduct(c *gin.Context) {

	var reqBody whatsapp.ProductReviewPayload
	err := c.BindJSON(&reqBody)
	if err != nil {
		JSONErrorResponse(c, 400, (fmt.Errorf("Request Invalid")))
		return
	}

	product, err := models.GetProduct(reqBody.ProductID)
	if err != nil {
		JSONErrorResponse(c, 404, fmt.Errorf("product not found: "+err.Error()))
		return
	}

	user, err := models.GetUserByID(reqBody.UserID)
	if err != nil {
		JSONErrorResponse(c, 404, fmt.Errorf("User not found: "+err.Error()))
		return
	}

	review, err := product.AddOrUpdateReview(user.ID, user.Name, float64(reqBody.Rating), reqBody.RatingStr)
	if err != nil {
		JSONErrorResponse(c, 500, err)
		return
	}

	err = Redis.DeleteProductCacheByID(product.ID)
	if err != nil {
		JSONErrorResponse(c, 500, err)
		return
	}
	c.JSON(200, review)
}
