package controllers

import (
	common "hermes/admin/controllers/common"
	"hermes/configs/Mysql"
	"hermes/controllers"
	"hermes/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type TransactionRow struct {
	OrderDate          string    `json:"order_date" bson:"order_date" gorm:"column:order_date"`
	UpdatedAt          time.Time `json:"updated_at" bson:"updated_at" gorm:"column:updated_at"`
	OrderID            string    `json:"order_id" bson:"order_id" gorm:"column:order_id"`
	ProductName        string    `json:"product_name" bson:"product_name" gorm:"column:product_name"`
	AWB                string    `json:"awb" bson:"awb" gorm:"column:awb"`
	DeliveredDate      time.Time `json:"delivered_date" bson:"delivered_date" gorm:"column:delivered_date"`
	RTODeliveredAt     time.Time `json:"rto_delivered_at" bson:"rto_delivered_at" gorm:"column:rto_delivered_at"`
	Quantity           int       `json:"Quantity" bson:"Quantity" gorm:"column:Quantity"`
	OrderAmount        float64   `json:"order_amount" bson:"order_amount" gorm:"column:order_amount"`
	PaymentMethod      string    `json:"payment_method" bson:"payment_method" gorm:"column:payment_method"`
	Status             string    `json:"status" bson:"status" gorm:"column:status"`
	CODCharges         float64   `json:"cod_charges" bson:"cod_charges" gorm:"column:cod_charges"`
	FWDAmount          float64   `json:"fwd_amount_charged" bson:"fwd_amount_charged" gorm:"column:fwd_amount_charged"`
	RTOAmount          float64   `json:"rto_amount_charged" bson:"rto_amount_charged" gorm:"column:rto_amount_charged"`
	TotalShippingCost  float64   `json:"total_shipping_cost" bson:"total_shipping_cost" gorm:"column:total_shipping_cost"`
	TransactionCharges float64   `json:"transaction_charges" bson:"transaction_charges" gorm:"column:transaction_charges"`
	TCS                float64   `json:"TCS_under_1_percent_GST" bson:"TCS_under_1_percent_GST" gorm:"column:TCS_under_1_percent_GST"`
	SellerID           string    `json:"seller_id" bson:"seller_id" gorm:"column:seller_id"`
	Commission         float64   `json:"commission" bson:"commission"`
}

func GetAllTransactions(c *gin.Context) {
	sellerID := c.Query("sellerID")
	var err error
	sellers, err := models.GetSellersList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
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

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": transactionArr,
		"sellers":      sellers,
	})
}

func GetTransactionByID(c *gin.Context) {
	transactionID := c.Query("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty transaction ID"})
		return
	}
	transaction, err := models.GetTransactionByID(transactionID)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty transaction ID"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"transaction": transaction})
}

func CreateTransaction(c *gin.Context) {
	newTransaction := models.Transaction{}
	err := c.BindJSON(&newTransaction)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
		return
	}

	if newTransaction.UTR != "" {
		err = models.MarkShipmentsAsRemitted(newTransaction.ShipmentsRemitted)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Could not mark shipments as remitted " + err.Error()})
			return
		}
		err = models.MarkOrdersAsRemitted(newTransaction.ShipmentsRemitted)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Could not mark shipments as remitted " + err.Error()})
			return
		}
		newTransaction.RemitReflected = true
	}
	err = newTransaction.Create()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"transaction": newTransaction})
}

func UpdateTransaction(c *gin.Context) {
	newBody := models.Transaction{}
	err := c.BindJSON(&newBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
		return
	}
	if newBody.UTR != "" {
		err = models.MarkShipmentsAsRemitted(newBody.ShipmentsRemitted)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Could not mark shipments as remitted " + err.Error()})
			return
		}
		err = models.MarkOrdersAsRemitted(newBody.ShipmentsRemitted)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Could not mark shipments as remitted " + err.Error()})
			return
		}
		newBody.RemitReflected = true
	}
	err = newBody.Update()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"transaction": newBody})
}

func GetDetailsFromTransaction(transactionDetails []TransactionRow) (charges models.Charges, remittableIDs []string, err error) {
	for _, val := range transactionDetails {
		charges.TotalTax = charges.TotalTax + val.TCS
		charges.TotalTransactionCharges = charges.TotalTransactionCharges + val.TransactionCharges
		charges.TotalForwardAmount = charges.TotalForwardAmount + val.FWDAmount
		charges.TotalRTOAmount = charges.TotalRTOAmount + val.RTOAmount
		charges.TotalCODCharges = charges.TotalCODCharges + val.CODCharges
		charges.TotalPayableAmount = charges.TotalPayableAmount - val.TCS - val.TransactionCharges - val.FWDAmount - val.RTOAmount - val.CODCharges

		if val.Status == "DELIVERED" {
			charges.TotalOrderAmount = charges.TotalOrderAmount + val.OrderAmount
			charges.TotalPayableAmount = charges.TotalPayableAmount + val.OrderAmount
		}
		remittableIDs = append(remittableIDs, val.OrderID)
	}
	return
}

func GetRemittableDetailsForTimeRange(c *gin.Context) {
	var transactionDetails []TransactionRow
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty date"})
		return
	}
	sellerID := c.Query("sellerID")
	if sellerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty seller ID"})
		return
	}

	query := "select * from (with cte2 as (with cte as (select date(date_add(date_add(coalesce(o.createdAt, s.createdAt), interval 5 hour), interval 30 minute)) as created_at, date_add(date_add(coalesce(o.updatedAt, s.createdAt), interval 5 hour), interval 30 minute) as updated_at, cast(s.id as char) as order_id, json_unquote(json_extract(so.awb_data, '$.awb')) as awb, coalesce(json_unquote(json_extract(o.product, '$.name')), json_unquote(json_extract(child_shipments, '$[0].productInfo[0].product.name'))) as product_name, json_unquote(json_extract(o.cart, '$.items[0].quantity')) as Quantity, case when s.parent_relation in (2,3,4) then 0 when s.parent_relation = 1 then 0-so.total else json_extract(o.payment, '$.amount') end as payment_amount, json_unquote(json_extract(o.payment, '$.method')) as payment_method, so.status as status, STR_TO_DATE(so.delivered_date, '%d-%m-%Y %H:%i:%s') as delivered_date, STR_TO_DATE(json_unquote(json_extract(so.shipments, '$.rto_delivered_date')), '%Y-%m-%d %H:%i:%s') rto_delivered_at, json_unquote(json_extract(so.awb_data, '$.charges.applied_weight')) as applied_weight, json_unquote(json_extract(so.awb_data, '$.charges.charged_weight')) as charged_weight, round(cast(json_unquote(json_extract(so.awb_data, '$.charges.cod_charges')) as float),2) as cod_charges, round(cast(json_unquote(json_extract(so.awb_data, '$.charges.applied_weight_amount')) as float),2) as applied_weight_amount, round(cast(json_unquote(json_extract(so.awb_data, '$.charges.charged_weight_amount')) as float),2) as charged_weight_amount, round(cast(json_unquote(json_extract(so.awb_data, '$.charges.applied_weight_amount_rto')) as float),2) as applied_weight_amount_rto, round(cast(json_unquote(json_extract(so.awb_data, '$.charges.charged_weight_amount_rto')) as float),2) as charged_weight_amount_rto, count(o.id) over(partition by json_unquote(json_extract(so.awb_data, '$.awb'))) as r, case when json_unquote(json_extract(so.awb_data, '$.charges.charged_weight')) > 0 and json_unquote(json_extract(so.awb_data, '$.charges.charged_weight')) > json_unquote(json_extract(so.awb_data, '$.charges.applied_weight')) then 'disputed' else null end as is_disputed, s.seller_id from shippings s left join Orders o on cast(o.id as char) = cast(s.id as char) left join shiprocket_orders so on cast(s.order_id as char) = cast(so.sr_id as char) left join backward_shipments as bs on cast(s.id as char) = cast(bs.id as char) where s.seller_id = ? AND s.remitted is not true AND o.remitted is not true AND ((STR_TO_DATE(so.delivered_date, '%d-%m-%Y %H:%i:%s') >= ? AND STR_TO_DATE(so.delivered_date, '%d-%m-%Y %H:%i:%s') <= ?) or (STR_TO_DATE(json_unquote(json_extract(so.shipments, '$.rto_delivered_date')), '%Y-%m-%d %H:%i:%s') >= ? AND STR_TO_DATE(json_unquote(json_extract(so.shipments, '$.rto_delivered_date')), '%Y-%m-%d %H:%i:%s') <= ?)) ) select created_at, updated_at, order_id, awb, product_name, Quantity, payment_amount, payment_method, status, delivered_date, rto_delivered_at, applied_weight, charged_weight, case when lower(status) like '%rto%' then 0 else cod_charges/r end as cod_charges, applied_weight_amount, charged_weight_amount, case when is_disputed = 'disputed' then (charged_weight_amount - cod_charges)/r else (applied_weight_amount - cod_charges)/r end as fwd_amount_charged, applied_weight_amount_rto, charged_weight_amount_rto, case when is_disputed = 'disputed' AND lower(status) like '%rto%' then charged_weight_amount_rto/r when is_disputed is null AND lower(status) like '%rto%' then applied_weight_amount_rto/r else 0 end as rto_amount_charged, r, seller_id from cte) select created_at as order_date, updated_at as updated_at, order_id, product_name, awb, delivered_date, rto_delivered_at, Quantity, payment_amount as order_amount, payment_method, status, cod_charges, fwd_amount_charged, rto_amount_charged, (cod_charges + fwd_amount_charged + rto_amount_charged) as total_shipping_cost, case when lower(payment_method) not in ('cod', 'prepaid') then 0.02*payment_amount else null end as transaction_charges, case when status = 'DELIVERED' then 0.01*payment_amount else null end as TCS_under_1_percent_GST, seller_id from cte2 order by 1 desc) as a union (select date(date_add(date_add(o.createdAt, interval 5 hour), interval 30 minute)) as order_date, date_add(date_add(o.updatedAt, interval 5 hour), interval 30 minute) as updated_at, cast(o.id as char) as order_id, json_unquote(json_extract(o.product, '$.name')) as product_name, null as awb, null as delivered_date, null as rto_delivered_date, json_unquote(json_extract(o.cart, '$.items[0].quantity')) as Quantity, json_extract(o.payment,'$.amount') as order_amount, json_unquote(json_extract(o.payment, '$.method')) as payment_method, fulfillmentStatus as status, null as cod_charges, null as fwd_amount_charged, null as rto_amount_charged, null as total_shipping_cost, case when lower(json_unquote(json_extract(o.payment, '$.method'))) not in ('cod', 'prepaid') then 0.02*json_extract(o.payment,'$.amount') else null end as transaction_charges, null as TCS_under_1_percent_GST, json_unquote(json_extract(product, '$.brand.sellerID')) as seller_id from Orders o WHERE json_unquote(json_extract(o.payment, '$.method')) is not null AND json_unquote(json_extract(o.payment, '$.method')) not in ('COD', 'Prepaid', ' ') AND lower(fulfillmentStatus) = 'CANCELLED' AND o.remitted is not true AND json_unquote(json_extract(product, '$.brand.sellerID')) = ? AND date_add(date_add(o.updatedAt, interval 5 hour), interval 30 minute) >= ? AND date_add(date_add(o.updatedAt, interval 5 hour), interval 30 minute)<= ? ) order by 1"
	err := Mysql.DB.Raw(query, sellerID, startDate, endDate, startDate, endDate, sellerID, startDate, endDate).Scan(&transactionDetails).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get charges " + err.Error()})
		return
	}
	// var temp []map[string]interface{}
	// err := resp.
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get transaction details " + err.Error()})
	// 	return
	// }

	charges, remittableOrders, err := GetDetailsFromTransaction(transactionDetails)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get cumulatives " + err.Error()})

		return
	}
	c.JSON(http.StatusOK, gin.H{"charges": charges, "remittableIDs": remittableOrders})
	return
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
