package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	helpers "hermes/admin/Helpers"
	"hermes/admin/controllers/common"
	"hermes/admin/services/auth"
	"hermes/db"
	"hermes/models"
	"hermes/services/shiprocket"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SellerOrder struct {
	models.Order `bson:",inline"`
	Shipping     models.Shipping           `json:"shipment" bson:"shipment"`
	Tracking     models.ShipRocketTracking `json:"tracking" bson:"tracking"`
}

var paymentStatusList = []string{"Paid", "Refunded", "Pending", "Completed"}
var fulfillmentStatusList = []string{"Fulfilled", "Pending"}

func AllOrdersPage(c *gin.Context) {
	var sellersList []models.Seller
	var sellerMap []map[string]interface{}

	currentSeller := c.Query("sellerID")
	orderDate := c.Query("date")

	cur, err := db.SellerCollection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sellers not found" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}

	err = cur.All(context.Background(), &sellersList)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sellers not correct format" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}

	marshalledSellers, _ := json.Marshal(sellersList)
	err = json.Unmarshal(marshalledSellers, &sellerMap)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sellers unmarshaaling error" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}

	var orders []models.Order
	var data []map[string]interface{}

	if len(currentSeller) == 0 && len(orderDate) == 0 {
		if len(sellersList) > 0 {
			currentSeller = sellersList[0].ID
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error ": "No Sellers found" + err.Error() + " currentSeller:" + currentSeller})
			return
		}
	}

	paymentFilter := []string{"Paid", "Succeeded"}

	// Define the pipeline
	pipeline := mongo.Pipeline{
		// Filter orders by sellerID
		bson.D{{Key: "$sort", Value: bson.D{{Key: "createdAt", Value: -1}}}},
		bson.D{{Key: "$match", Value: bson.D{{Key: "paymentStatus", Value: bson.D{{Key: "$in", Value: paymentFilter}}}}}},
		// Lookup user by userID and add internal to orders documents
		bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "users"}, {Key: "localField", Value: "userId"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "user"}}}},
		bson.D{{Key: "$unwind", Value: "$user"}},
		bson.D{{Key: "$addFields", Value: bson.D{{Key: "internal", Value: "$user.internal"}}}},
		// Filter orders by internal = false
		bson.D{{Key: "$match", Value: bson.D{{Key: "$or", Value: bson.A{
			bson.D{{Key: "internal", Value: false}},
			bson.D{{Key: "internal", Value: bson.D{{Key: "$exists", Value: false}}}},
		}}}}},
	}

	if len(orderDate) > 0 {
		filterDate, err := time.Parse("2006-01-02", orderDate)

		if err != nil {
			fmt.Println("err:", err.Error())
		} else {
			pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.D{{Key: "createdAt", Value: bson.D{{Key: "$gte", Value: primitive.NewDateTimeFromTime(filterDate)}}}}}})
		}
	} else {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.D{{Key: "product.sellerID", Value: currentSeller}}}})
	}

	cur, err = db.OrderCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}

	err = cur.All(context.Background(), &orders)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not correct format" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}

	defer cur.Close(context.Background())

	marshalledOrders, err := json.Marshal(orders)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order marshalling not correct format" + err.Error()})
		return
	}
	err = json.Unmarshal(marshalledOrders, &data)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order unmarshalling not correct format" + err.Error()})
		return
	}

	for idx, val := range orders {
		var shipment map[string]interface{}
		db.ShippingCollection.FindOne(context.Background(), bson.M{"_id": val.ID}).Decode(&shipment)
		data[idx]["createdAt"] = helpers.MakeTimeStringfromTimestamp(val.UpdatedAt)
		data[idx]["shipment"] = shipment
	}

	c.HTML(http.StatusOK, "root", gin.H{"title": "Roovo", "orders": data, "sellers": sellerMap, "currentSeller": currentSeller, "orderDate": orderDate, "template": "admin-order"})
}

func OrderViewPage(c *gin.Context) {
	orderId := c.Param("orderId")
	var order models.Order
	var data map[string]interface{}
	err := db.OrderCollection.FindOne(context.Background(), bson.M{"_id": orderId}).Decode(&order)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}
	marshalledOrders, _ := json.Marshal(order)
	json.Unmarshal(marshalledOrders, &data)
	var temp map[string]interface{}
	db.UserCollection.FindOne(context.Background(), bson.M{"_id": data["userId"]}).Decode(&temp)
	data["userId"] = temp
	data["createdAt"] = helpers.MakeTimeStringfromTimestamp(order.CreatedAt)
	c.HTML(http.StatusOK, "root", gin.H{"title": "Admin | Roovo", "order": data, "template": "order-view", "paymentStatusList": paymentStatusList, "fulfillmentStatusList": fulfillmentStatusList})
}

func NewOrderPage(c *gin.Context) {
	var order models.Order

	var products []models.Product
	cur, err := db.ProductCollection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}
	err = cur.All(context.Background(), &products)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Product not right format" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}

	var users []models.User
	cur, err = db.UserCollection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Users not found" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}
	err = cur.All(context.Background(), &users)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Users not right format" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}
	var primaryArr []models.Address
	cur, _ = db.AddressCollection.Find(context.Background(), bson.M{"userID": users[0].ID})
	err = cur.All(context.Background(), &primaryArr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Users not right format" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}

	c.HTML(http.StatusOK, "root", gin.H{"title": "Admin | Roovo", "order": order, "template": "order-create", "paymentStatusList": paymentStatusList, "fulfillmentStatusList": fulfillmentStatusList, "products": products, "users": users, "firstAddress": primaryArr})
}

func GetAddressesForUser(c *gin.Context) {
	userId := c.Query("userId")
	var addresses []models.Address
	cur, err := db.AddressCollection.Find(context.Background(), bson.M{"userID": userId})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Address not found" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}
	err = cur.All(context.Background(), &addresses)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Address not right format" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"addresses": addresses})
}

// func AddOrder(c *gin.Context) {
// 	newOrder := make(map[string]interface{})
// 	err := c.BindJSON(&newOrder)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
// 		return
// 	}

// 	var tempProduct models.Product
// 	db.ProductCollection.FindOne(context.Background(), bson.M{"_id": newOrder["product"]}).Decode(&tempProduct)
// 	newOrder["product"] = tempProduct

// 	var tempAddress models.Address
// 	db.AddressCollection.FindOne(context.Background(), bson.M{"_id": newOrder["address"]}).Decode(&tempAddress)
// 	newOrder["address"] = tempAddress

// 	newOrder["_id"] = data.GetUUIDString("order")
// 	newOrder["createdAt"] = time.Now()
// 	newOrder["updatedAt"] = time.Now()
// 	_, err = db.OrderCollection.InsertOne(context.Background(), newOrder)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to update " + err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, gin.H{"response": "Successfully created order"})
// }

func UpdateOrder(c *gin.Context) {
	orderId := c.Query("id")
	var data map[string]interface{}
	err := c.BindJSON(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body:" + err.Error()})
		return
	}

	var tempProduct models.Product
	db.ProductCollection.FindOne(context.Background(), bson.M{"_id": data["product"]}).Decode(&tempProduct)
	data["product"] = tempProduct

	var tempAddress models.Address
	db.AddressCollection.FindOne(context.Background(), bson.M{"_id": data["address"]}).Decode(&tempAddress)
	data["address"] = tempAddress

	result, err := db.OrderCollection.ReplaceOne(context.Background(), bson.M{"_id": orderId}, data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to update " + err.Error()})
		return
	}
	if result.ModifiedCount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not find order with given ID"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "Successfully updated product with given ID"})
}

func DeleteOrder(c *gin.Context) {
	orderId := c.Query("id")
	result, err := db.OrderCollection.DeleteOne(context.Background(), bson.M{"_id": orderId})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to delete " + err.Error()})
		return
	}
	if result.DeletedCount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not find order with given ID"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "Successfully deleted order with given ID"})
}

func GetOrdersListForSeller(c *gin.Context) {
	seller, err := auth.GetSellerFromSession(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Seller not found" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}

	currentSeller := seller.ID
	currentBrand := c.Query("brandID")
	ordersStartDate := c.Query("startDate")
	ordersEndDate := c.Query("endDate")
	shipmentStatus := c.Query("status")
	awbs := c.Query("awb")

	orders, shipmentStatuses, err := common.OrderFilterUtility(currentSeller, currentBrand, ordersStartDate, ordersEndDate, shipmentStatus, awbs, true)
	if err != nil {
		return
	}

	var sellersList []models.Seller
	var brandsList []models.Brand
	sellersList = append(sellersList, seller)

	cur, err := db.BrandsCollection.Find(context.Background(), bson.M{"sellerID": currentSeller})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Brands not found" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}

	err = cur.All(context.Background(), &brandsList)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Brands not correct format: " + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders, "sellers": sellersList, "brands": brandsList, "statuses": shipmentStatuses, "currentSeller": currentSeller, "currentBrand": currentBrand, "ordersStartDate": ordersStartDate, "ordersEndDate": ordersEndDate, "shipmentStatus": shipmentStatus})
}

func CreateBulkShipment(c *gin.Context) {

	type BulkShipmentRequestbody struct {
		OrderIds []string `json:"orderIDs"`
		Discount int64    `json:"discount"`
	}

	var bulkShipmentRequestbody BulkShipmentRequestbody
	if err := c.ShouldBind(&bulkShipmentRequestbody); err != nil {
		fmt.Printf("err: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var orders []models.Order
	for _, orderID := range bulkShipmentRequestbody.OrderIds {
		order, err := models.GetOrder(orderID)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		orders = append(orders, order)
	}

	sellerID := orders[0].Product.SellerID
	seller, err := models.GetSellerByID(sellerID)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = shiprocket.CreateShipRocketBulkOrder(orders, seller, bulkShipmentRequestbody.Discount)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": "Success"})
}

func CancelOrder(c *gin.Context) {

	seller, err := auth.GetSellerFromSession(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Seller not found" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}

	orderID := c.Query("id")
	shiprocketOrderID := c.Query("shiprocketOrderID")

	var requestBody map[string]string
	if err := c.ShouldBind(&requestBody); err != nil {
		fmt.Printf("err: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reason := requestBody["reason"]
	if len(reason) < 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cancellation reason needed of min length of 10 characters"})
		return
	}

	if len(orderID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID empty"})
		return
	}

	if len(shiprocketOrderID) > 0 {
		err := shiprocket.CancelShiprocketOrder(shiprocketOrderID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	order, err := models.GetOrder(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if order.Product.SellerID != seller.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Seller mismatch required: %s found: %s", order.Product.SellerID, seller.ID)})
		return
	}

	order.FulfillmentStatus = "Cancelled"
	order.CancellationReason = reason + " ,cancelled on seller"
	err = order.Update()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": order})
}

func DownloadOrderReport(c *gin.Context) {

	seller, err := auth.GetSellerFromSession(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Seller not found" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}

	currentSeller := seller.ID
	currentBrand := c.Query("brandID")
	ordersStartDate := c.Query("startDate")
	ordersEndDate := c.Query("endDate")
	shipmentStatus := c.Query("status")
	awbs := c.Query("awb")

	orders, _, err := common.OrderFilterUtility(currentSeller, currentBrand, ordersStartDate, ordersEndDate, shipmentStatus, awbs, true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	csvData, err := common.DownloadOrderReportUtil(orders)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=roovo-orders.csv")
	c.Header("Content-Type", "text/csv")
	c.Data(http.StatusOK, "text/csv", csvData)
}
