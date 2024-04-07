package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"hermes/controllers"
	"hermes/db"
	"hermes/models"
	utils "hermes/utils/queries"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func AllCategoriesPage(c *gin.Context) {
	var categories []models.Category
	var data []map[string]interface{}
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

	marshalledProducts, _ := json.Marshal(categories)
	json.Unmarshal(marshalledProducts, &data)
	for idx, val := range categories {
		currCount, _ := db.ProductCollection.CountDocuments(context.Background(), bson.M{"category": val.ID})
		data[idx]["count"] = currCount
	}

	c.HTML(http.StatusOK, "root", gin.H{"title": "Admin | Roovo", "categories": data, "template": "category"})
}

func CategoryEditPage(c *gin.Context) {
	categoryId := c.Param("categoryId")
	var category models.Category
	err := db.CategoryCollection.FindOne(context.Background(), bson.M{"_id": categoryId}).Decode(&category)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}
	c.HTML(http.StatusOK, "root", gin.H{"title": "Admin | Roovo", "category": category, "template": "category-edit", "new": false})
}

func NewCategoryPage(c *gin.Context) {
	var category models.Category
	c.HTML(http.StatusOK, "root", gin.H{"title": "Admin | Roovo", "category": category, "template": "category-edit", "new": true})
}

func CreateCategory(c *gin.Context) {
	newCategory := models.Category{}
	err := c.BindJSON(&newCategory)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
		return
	}
	err = newCategory.Create()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create category"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "Successfully created category"})
}

func UpdateCategory(c *gin.Context) {
	var data models.Category
	err := c.BindJSON(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
		return
	}
	err = data.Update()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update category"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "Successfully updated category with given ID"})
}

func DeleteCategory(c *gin.Context) {
	categoryId := c.Query("id")
	result, err := db.CategoryCollection.DeleteOne(context.Background(), bson.M{"_id": categoryId})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to delete " + err.Error()})
		return
	}
	if result.DeletedCount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not find category with given ID"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "Successfully deleted category with given ID"})
}

func GetCategoryListPage(c *gin.Context) {
	var categories []models.Category
	name := c.Query("name")
	if name == "" {
		GetCategoryPage(c)
		return
	} else {
		pipeline := bson.A{bson.D{{Key: "$match", Value: bson.D{{Key: "$text", Value: bson.D{{Key: "$search", Value: name}}}}}}}
		pipeline = append(pipeline, utils.FullCategoryQuery...)
		cur, err := db.CategoryCollection.Aggregate(context.Background(), pipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get categories " + err.Error()})
			return
		}
		err = cur.All(context.Background(), &categories)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to unmarshal categories " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"categories": categories})
		return
	}
}

func GetCompleteCategoryList(c *gin.Context) {
	categoryList, err := models.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get category list " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"categories": categoryList})
}

func GetCategoryByID(c *gin.Context) {
	categoryId := c.Param("id")
	var category models.Category
	var err error
	if categoryId != "new" {
		category, err = models.GetCategoryByID(categoryId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get full category " + err.Error()})
			return
		}
	}
	categoryList, err := models.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get category list " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"category": category, "categoryList": categoryList})
}

func GetCategoryPage(c *gin.Context) {
	var err error
	var categories []models.Category
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

	var categoryPagination *controllers.Pagination
	name := c.Query("name")
	if name != "" {
		pipeline := bson.A{bson.D{{Key: "$match", Value: bson.D{{Key: "$text", Value: bson.D{{Key: "$search", Value: name}}}}}}}
		pipeline = append(pipeline, utils.FullCategoryQuery...)
		categoryPagination, err = controllers.Paginate("category", &Paginater, categories, pipeline, bson.A{})
	} else {
		categoryPagination, err = controllers.Paginate("category", &Paginater, categories, utils.FullCategoryQuery, bson.A{})
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error in pagination": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"categories": categoryPagination})
}
