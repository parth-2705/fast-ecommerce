package controllers

import (
	"context"
	"fmt"
	"hermes/db"
	"hermes/models"
	"hermes/search"
	"hermes/utils/amplitude"
	utils "hermes/utils/queries"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type BrandCategory struct {
	Category models.Category `json:"category" bson:"category"`
}

func BrandPage(c *gin.Context) {
	brandID := c.Param("brandID")
	var products []models.Product
	processedAppliedFilterMap := map[string]search.FilterObject{}

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

	var Paginater Pagination = Pagination{
		Limit: limitInt,
		Page:  pageInt,
	}

	go amplitude.TrackEventWithPropertiesByAuth("Brand Category", map[string]interface{}{"id": brandID}, c)

	var brand models.Brand
	brand, err = models.GetBrand(brandID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	processedAppliedFilterMap["brandID"] = search.FilterObject{
		Values:   []string{brandID},
		Operator: "=",
		Path:     "brandID",
	}
	sortArr := []search.SortObject{}
	products, err = ProductPaginate2(&Paginater, "", processedAppliedFilterMap, sortArr)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	pipeline := utils.GetBrandCategoriesByID(brandID)
	var categories []BrandCategory
	cur, err := db.BrandsCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	err = cur.All(context.Background(), &categories)
	if err != nil {
		c.AbortWithError(http.StatusServiceUnavailable, err)
		return
	}

	c.HTML(http.StatusOK, "brandPage", gin.H{
		"brand":      brand,
		"products":   products,
		"categories": categories,
		"paginater":  Paginater,
	})
}

func GetProductByBrand(brandID string) (products []models.Product, err error) {
	filter := bson.D{{Key: "$match", Value: bson.M{"brandID": brandID}}}
	cur, err := getProductsWithBrand(filter)
	if err != nil {
		return nil, err
	}
	err = cur.All(context.Background(), &products)
	return
}

func GetCategoriesInProductArr(productArr []models.Product) (categories []models.Category, err error) {
	idArr := GetIdArrFromProductArr(productArr)
	fmt.Println("idArr:", idArr)
	categories, err = GetCategoryArrFromIdArr(idArr)
	if err != nil {
		return nil, err
	}
	return
}

func GetCategoryArrFromIdArr(idArr []string) (categories []models.Category, err error) {
	for _, val := range idArr {
		var category models.Category
		err = db.CategoryCollection.FindOne(context.Background(), bson.M{"_id": val}).Decode(&category)
		if err != nil {
			return
		}
		categories = append(categories, category)
	}
	return
}

func GetIdArrFromProductArr(productArr []models.Product) []string {
	var idMap map[string]struct{} = make(map[string]struct{})
	for _, val := range productArr {
		if _, ok := idMap[val.NewCategory]; !ok {
			idMap[val.NewCategory] = struct{}{}
		}
	}
	idArr := make([]string, 0, len(idMap))
	for k := range idMap {
		idArr = append(idArr, k)
	}
	return idArr
}
