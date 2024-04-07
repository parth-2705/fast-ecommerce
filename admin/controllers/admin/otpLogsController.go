package controllers

import (
	"hermes/controllers"
	"hermes/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func GetOTPLogsList(c *gin.Context) {

	limitInt := controllers.GetLimitFromQueryValue(c.Query("limit"))
	pageInt := controllers.GetPageFromQueryValue(c.Query("page"))

	var Paginater controllers.Pagination = controllers.Pagination{
		Limit: limitInt,
		Page:  pageInt,
	}

	var logsStruct []*models.AdminLogs
	preSkip := bson.A{bson.D{{Key: "$sort", Value: bson.D{{Key: "sendTime", Value: -1}}}}}
	postSkip := bson.A{bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "users"},
			{Key: "localField", Value: "mobileNumber"},
			{Key: "foreignField", Value: "phone"},
			{Key: "as", Value: "user"},
		}},
	},
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$user"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		bson.D{{Key: "$addFields", Value: bson.D{{Key: "internal", Value: "$user.internal"}}}}}

	logs, err := controllers.Paginate("logs", &Paginater, logsStruct, preSkip, postSkip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":       logs.Rows,
		"totalRows":  logs.TotalRows,
		"totalPages": logs.TotalPages,
	})
}
