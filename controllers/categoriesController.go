package controllers

import (
	"bytes"
	"context"
	"fmt"
	"hermes/db"
	"hermes/models"
	"hermes/search"
	"hermes/utils/amplitude"
	"hermes/utils/network"
	utils "hermes/utils/queries"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func ListingsPage(c *gin.Context) {
	categoryID := c.Param("category")

	go amplitude.TrackEventWithPropertiesByAuth("Specific Category", map[string]interface{}{"id": categoryID}, c)
	CategoriesPageHelper(c, categoryID)
}

func HomePage(c *gin.Context) {
	go amplitude.TrackEventByAuth("Home Page", c)
	CategoriesPageHelper(c, "")
}

type ChildrenIDResponse struct {
	ID         string   `json:"_id" bson:"_id"`
	ChildrenID []string `json:"childrenID" bson:"childrenID"`
}

func RobotsTxt(c *gin.Context) {

	path := "static/assets/robots.txt"
	path = filepath.FromSlash(path)
	if len(path) > 0 && !os.IsPathSeparator(path[0]) {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		path = filepath.Join(wd, path)
	}

	file, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	reader := bytes.NewReader(file)

	c.Header("Content-Type", "text/plain")
	http.ServeContent(c.Writer, c.Request, "robots.txt", time.Now(), reader)
}

func SelectCategoriesPage(c *gin.Context) {

	go amplitude.TrackEventByAuth("Categories Page", c)

	categories, err := GetCategoriesInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if network.MobileRequest(c) {
		c.JSON(http.StatusOK, gin.H{
			"categories": categories,
		})

		return
	}

	trackingMap := amplitude.GetTrackingMap(c)

	c.HTML(http.StatusOK, "categoriesPage", gin.H{
		"categories":  categories,
		"trackingMap": trackingMap,
	})
}

func GetCategoriesInfo() (categories []models.Category, err error) {
	pipeline := bson.A{bson.D{{"$match", bson.M{"parentID": "", "ranking": bson.M{"$exists": true}}}}, bson.D{{Key: "$sort", Value: bson.M{"ranking": -1}}}}
	pipeline = append(pipeline, utils.CategoriesPageQuery...)
	categoriesCursor, err := db.CategoryCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return
	}
	if err = categoriesCursor.All(context.TODO(), &categories); err != nil {
		return
	}
	return
}

func CategoriesPageHelper(c *gin.Context, categoryID string) {

	var err error

	//apply filters logic
	appliedfiltersQuery := c.Query("filter")
	appliedfiltersArr := []string{}
	if appliedfiltersQuery != "" {
		appliedfiltersArr = strings.Split(appliedfiltersQuery, ",")
	}
	appliedFilters := map[string]bool{}
	processedAppliedFilterMap := map[string]search.FilterObject{}
	for _, val := range appliedfiltersArr {
		appliedFilters[val] = true
		filterItem := strings.Split(val, `/\`)
		temp := append(processedAppliedFilterMap[filterItem[0]].Values, filterItem[1])
		processedAppliedFilterMap[filterItem[0]] = search.FilterObject{
			Values:   temp,
			Operator: "=",
		}
	}

	//Pagination Logic
	limit := c.Query("limit")
	limitInt := 20
	if limit != "" {
		limitInt, err = strconv.Atoi(limit)
	}
	pageInt := 1

	var Paginater Pagination = Pagination{
		Limit: limitInt,
		Page:  pageInt,
	}
	sortArr := []search.SortObject{}
	products := []models.Product{}
	homepage := false
	var category models.Category
	var categories []models.Category

	if categoryID != "" {
		//Apply Catalog ID Filter
		categoryIDArr, err := GetChildrenCategoryIDArr(categoryID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Error fetching children categories " + err.Error()})
			return
		}
		categoryFilter := []string{categoryID}
		categoryFilter = append(categoryFilter, categoryIDArr...)
		processedAppliedFilterMap["category"] = search.FilterObject{
			Values:   categoryFilter,
			Operator: "=",
			Path:     "newCategory",
		}
		sortArr = append(sortArr, search.SortObject{Order: "desc", Path: "pageRanking"})
		products, err = ProductPaginate2(&Paginater, "", processedAppliedFilterMap, sortArr)

		category, err = GetCompleteCategory(categoryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to fetch category"})
			return
		}
		fmt.Println("In this case")
		categories, err = GetSimilarCategories(categoryID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Error fetching categories"})
			fmt.Printf("err: %v\n", err)
			return
		}
	} else {
		//homepage
		homepage = true
		sortArr = append(sortArr, search.SortObject{Order: "desc", Path: "pageRanking"})
		products, err = ProductPaginate2(&Paginater, "", processedAppliedFilterMap, sortArr)
	}

	for i := range products {

		//calculate ratings 
		products[i].RatingVisualization, err = models.GetRatingVisualizationObject(products[i].AverageRating)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Error fetching products for the given category"})
			fmt.Printf("err: %v\n", err)
			return
		}

		// lowestPrice := products[i].Variants[0].Price.SellingPrice

		// for _, variant := range products[i].Variants {
		// 	if variant.Price.SellingPrice <= lowestPrice {
		// 		lowestPrice = variant.Price.SellingPrice
		// 		setProductViewPrice(&products[i], variant)
		// 	}
		// }

	}

	//filter options logic
	filtersMap := category.Filters
	filters := []string{}
	for key := range filtersMap {
		filters = append(filters, key)
	}

	trackingMap := amplitude.GetTrackingMap(c)

	if network.MobileRequest(c) {
		c.JSON(http.StatusOK, gin.H{
			"categories":     categories,
			"products":       products,
			"homePage":       homepage,
			"category":       category,
			"filters":        filters,
			"filtersMap":     filtersMap,
			"paginater":      Paginater,
			"appliedFilters": appliedFilters,
			"trackingMap":    trackingMap,
		})

		return
	}

	// Render the categories page
	c.HTML(http.StatusOK, "listingsPage", gin.H{
		"categories":     categories,
		"products":       products,
		"homePage":       homepage,
		"categoryID":     categoryID,
		"category":       category,
		"filters":        filters,
		"filtersMap":     filtersMap,
		"paginater":      Paginater,
		"appliedFilters": appliedFilters,
		"trackingMap":    trackingMap,
	})
}

func GetChildrenCategoryIDArr(categoryID string) (categoryIDArr []string, err error) {

	var resp []ChildrenIDResponse
	pipeline := bson.A{bson.D{{"$match", bson.D{{"_id", categoryID}}}}}
	pipeline = append(pipeline, utils.CategoriesChildrenIDQuery...)
	cur, err := db.CategoryCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return
	}
	err = cur.All(context.Background(), &resp)
	if err != nil {
		return
	}
	fmt.Println("THis is resp:", resp)
	categoryIDArr = resp[0].ChildrenID
	return
}

func GetSimilarCategories(categoryID string) (categories []models.Category, err error) {
	var temp models.Category
	err = db.CategoryCollection.FindOne(context.Background(), bson.M{"_id": categoryID}).Decode(&temp)
	if err != nil {
		return
	}
	cur, err := db.CategoryCollection.Find(context.Background(), bson.M{"parentID": temp.ParentID})
	if err != nil {
		return
	}
	err = cur.All(context.Background(), &categories)
	return
}

func GetCompleteCategory(categoryID string) (category models.Category, err error) {
	var resp []models.Category
	fmt.Println("This is resp:", categoryID)
	pipeline := bson.A{bson.D{{"$match", bson.D{{"_id", categoryID}}}}}
	pipeline = append(pipeline, utils.CategoriesPageQuery...)
	cur, err := db.CategoryCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return
	}
	err = cur.All(context.Background(), &resp)
	if err != nil {
		return
	}
	fmt.Printf("len(resp): %v\n", len(resp))
	fmt.Println("This is resp:", resp)

	for i := range resp {
		fmt.Printf("resp[i].Name: %v\n", resp[i].Name)
	}

	return resp[0], nil
}

func setProductViewPrice(product *models.Product, variant models.Variation) {
	product.Price = variant.Price
	product.Quantity = variant.Quantity
	product.Images = variant.Images
	product.Thumbnail = variant.Thumbnail
	product.SKU = variant.SKU
	product.Barcode = variant.Barcode
	product.Status = variant.Status
}
