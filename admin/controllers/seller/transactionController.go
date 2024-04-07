package controllers

import (
	common "hermes/admin/controllers/common"
	"hermes/admin/services/auth"
	"hermes/controllers"
	"hermes/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func GetAllTransactions(c *gin.Context) {
	sellerID, err := auth.GetSellerIdFromSession(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	limitInt := controllers.GetLimitFromQueryValue(c.Query("limit"))
	pageInt := controllers.GetPageFromQueryValue(c.Query("page"))

	var Paginater controllers.Pagination = controllers.Pagination{
		Limit: limitInt,
		Page:  pageInt,
	}

	var transactions []models.Transaction
	preSkip := bson.A{}
	if sellerID != "" {
		preSkip = append(preSkip, bson.D{{Key: "$match", Value: bson.D{{Key: "sellerID", Value: sellerID}}}})
	}

	dateFilter := bson.M{}
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	if startDate != "" {
		startDateTime, _ := time.Parse("2006-01-02", startDate)
		dateFilter["$gte"] = startDateTime
	}
	if endDate != "" {
		endDateTime, _ := time.Parse("2006-01-02", endDate)
		dateFilter["$lte"] = endDateTime.Add(24 * time.Hour)
	}

	if len(dateFilter) > 0 {
		preSkip = append(preSkip, bson.D{{Key: "$match", Value: bson.D{{Key: "createdAt", Value: dateFilter}}}})
	}

	postSkip := bson.A{}
	transactionArr, err := controllers.Paginate("transaction", &Paginater, transactions, preSkip, postSkip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": transactionArr,
	})
}

func DownloadTransactionReport(c *gin.Context) {
	transactionID := c.Param("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty transaction ID"})
		return
	}
	transactions, err := models.GetTransactionByID(transactionID)
	if err != nil || len(transactions) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No transactions found " + err.Error()})
		return
	}
	csvData, err := common.DownloadTransactionReportUtil(transactions[0])
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+transactionID+".csv")
	c.Header("Content-Type", "text/csv")
	c.Data(http.StatusOK, "text/csv", csvData)
}
