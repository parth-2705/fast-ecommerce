package controllers

import (
	"hermes/controllers"
	"hermes/models"
	utils "hermes/utils/queries"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func GetPageViews(c *gin.Context) {

	limitInt := controllers.GetLimitFromQueryValue(c.Query("limit"))
	pageInt := controllers.GetPageFromQueryValue(c.Query("page"))

	var Paginater controllers.Pagination = controllers.Pagination{
		Limit: limitInt,
		Page:  pageInt,
	}

	var pageViewsStruct []*models.PageView
	filter := bson.A{bson.D{{Key: "$sort", Value: bson.D{{Key: "visitTime", Value: -1}}}}}
	pageViews, err := controllers.Paginate("pageView", &Paginater, pageViewsStruct, filter, utils.PageViewQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pageViews":  pageViews.Rows,
		"totalRows":  pageViews.TotalRows,
		"totalPages": pageViews.TotalPages,
	})
}
