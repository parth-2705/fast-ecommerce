package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	helpers "hermes/admin/Helpers"
	common "hermes/admin/controllers/common"
	"hermes/db"
	"hermes/models"
	"hermes/services/Sentry"
	"hermes/services/shiprocket"
	"hermes/utils/data"
	"hermes/utils/payments"
	"hermes/utils/rw"
	"hermes/utils/whatsapp"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

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

func GetOrdersList(c *gin.Context) {

	currentSeller := c.Query("sellerID")
	currentBrand := c.Query("brandID")
	ordersStartDate := c.Query("startDate")
	ordersEndDate := c.Query("endDate")
	shipmentStatus := c.Query("status")
	awbs := c.Query("awb")

	orders, shipmentStatuses, err := common.OrderFilterUtility(currentSeller, currentBrand, ordersStartDate, ordersEndDate, shipmentStatus, awbs, false)
	if err != nil {
		return
	}

	var sellersList []models.Seller
	var brandsList []models.Brand
	cur, err := db.SellerCollection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sellers not found" + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}

	err = cur.All(context.Background(), &sellersList)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sellers not correct format: " + err.Error()})
		fmt.Printf("err: %v\n", err)
		return
	}

	cur, err = db.BrandsCollection.Find(context.Background(), bson.M{})
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

func DownloadOrderReport(c *gin.Context) {
	currentSeller := c.Query("sellerID")
	currentBrand := c.Query("brandID")
	ordersStartDate := c.Query("startDate")
	ordersEndDate := c.Query("endDate")
	shipmentStatus := c.Query("status")
	awbs := c.Query("awb")

	orders, _, err := common.OrderFilterUtility(currentSeller, currentBrand, ordersStartDate, ordersEndDate, shipmentStatus, awbs, false)
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

func CreateNewOrder(c *gin.Context) {

	type CreateOrderRequestBody struct {
		models.Order `bson:",inline"`
		Quantity     string `json:"quantity" bson:"quantity"`
	}

	var requestBody CreateOrderRequestBody

	if err := c.ShouldBind(&requestBody); err != nil {
		fmt.Println("err:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	paymentMethod := c.Query("paymentMethod")
	if len(paymentMethod) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No Payment Method Specified"})
		return
	}

	var newOrder models.Order
	newOrder.ID = data.GetUUIDString("order")
	newOrder.CreatedAt = time.Now()
	newOrder.UpdatedAt = time.Now()
	newOrder.UserID = requestBody.UserID

	user, err := models.GetUserByID(newOrder.UserID)
	if err != nil {
		fmt.Println("err:", err.Error())
		Sentry.SendErrorToSentry(c, err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "chat": "user get error"})
		return
	}

	newOrder.User = user

	product, err := models.GetCompleteProduct(requestBody.Product.ID)
	if err != nil {
		fmt.Println("err:", err.Error())
		Sentry.SendErrorToSentry(c, err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "chat": "product get error"})
		return
	}

	newOrder.Product = product

	variants, err := models.GetVariantsByProductID(product.ID)
	if err != nil {
		fmt.Println("err:", err.Error())
		Sentry.SendErrorToSentry(c, err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "chat": "variants get error"})
		return
	}
	if len(variants) > 1 {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("more than one variant"))
		return
	}

	newOrder.VariantID = variants[0].ID
	newOrder.Variant = variants[0]
	address, err := models.GetAddress(requestBody.Address.ID)
	if err != nil {
		fmt.Println("err:", err.Error())
		Sentry.SendErrorToSentry(c, err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "chat": "address get error"})
		return
	}
	newOrder.Address = address
	newOrder.FulfillmentStatus = "Unfulfilled"

	brand, err := models.GetBrand(newOrder.Product.BrandID)
	if err != nil {
		fmt.Println("err:", err.Error())
		Sentry.SendErrorToSentry(c, err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "chat": "brand get error"})
		return
	}

	orderQuantity, err := strconv.Atoi(requestBody.Quantity)
	if err != nil {
		fmt.Println("err:", err.Error())
		Sentry.SendErrorToSentry(c, err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newOrder.Brand = brand

	var payment models.Payment
	payment.ID = data.GetUUIDString("payment")
	payment.OrderID = newOrder.ID
	payment.Method = paymentMethod
	payment.Amount = int64(newOrder.Product.Price.SellingPrice) * int64(orderQuantity)
	payment.Currency = "INR"
	if paymentMethod == "COD" || paymentMethod == "Prepaid" {
		payment.Status = "Paid"
	} else {
		payment.Status = "Unpaid"
	}

	newOrder.PaymentStatus = payment.Status
	items := []models.Item{{
		Product:   newOrder.Product,
		Variant:   newOrder.Variant,
		ProductID: newOrder.Product.ID,
		VariantID: newOrder.Variant.ID,
		Quantity:  orderQuantity,
	}}

	cart, err := models.CreateNewCart2(newOrder.UserID, "", items, models.AdminPortal)
	// cart, err := models.CreateNewCart("", &newOrder.Product, &newOrder.Variant, &newOrder.Brand, "", requestBody.UserID, orderQuantity, models.AdminPortal)
	if err != nil {
		fmt.Println("err:", err.Error())
		Sentry.SendErrorToSentry(c, err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newOrder.CartID = cart.ID
	newOrder.Cart = cart
	newOrder.OrderAmount = cart.CartAmount
	newOrder.Source = models.AdminPortal

	err = newOrder.Create()
	if err != nil {
		fmt.Println("err:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	newOrder.Payment = payment
	err = newOrder.Update()
	if err != nil {
		fmt.Println("err:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if payment.Status == "Paid" {
		err = newOrder.MarkOrderAsCompleted(payment)
		if err != nil {
			fmt.Println("err:", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		payment, err = payment.Create()
		if err != nil {
			fmt.Println("err:", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err = newOrder.MarkOrderAsFulfillable()
		if err != nil {
			fmt.Println("err:", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		upiResponse, decentroResponse, err := payments.InitiateUPIPayment(payment.ID, float64(payment.Amount), "Roovo Payment")
		if err != nil {
			fmt.Println("err:", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		payment.ThirdPartyPaymentObject = decentroResponse
		payment, err = payment.Create()
		link := upiResponse.UPIDeepLinks.GeneratedLink
		_, err = models.SaveUPIPayObject(link, payment.ID)
		if err != nil {
			return
		}
		upiPay, err := models.GetUPIPayObjectByPayment(payment.ID)
		whatsapp.SendPaymentLinkMessage(address.Phone, upiPay.ID, payment.Amount)

	}
	c.JSON(http.StatusOK, gin.H{
		"response": newOrder,
	})
}

func CreateNewOrderFromCart(c *gin.Context) {
	var newOrder models.Order

	// raw, err := c.GetRawData()
	// fmt.Printf("raw: %v\n", string(raw))
	if err := c.ShouldBindJSON(&newOrder); err != nil {
		fmt.Println("unmarshalling err:", err.Error())
		Sentry.SendErrorToSentry(c, err, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tags := map[string]string{
		"user":   newOrder.UserID,
		"source": "whatsapp",
		"action": "order",
		"phone":  newOrder.Address.Phone,
	}

	newOrder.ID = data.GetUUIDString("order")
	newOrder.CreatedAt = time.Now()
	newOrder.UpdatedAt = time.Now()

	user, err := models.GetUserByID(newOrder.UserID)
	if err != nil {
		Sentry.SendErrorToSentry(c, err, tags)
		fmt.Println("get user err:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "chat": "user get error"})
		return
	}

	newOrder.User = user

	address, err := models.GetAddress(newOrder.Address.ID)
	if err != nil {
		Sentry.SendErrorToSentry(c, err, tags)
		fmt.Println("get address err:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "chat": "address get error"})
		return
	}
	newOrder.Address = address
	newOrder.FulfillmentStatus = "Unfulfilled"

	cart, err := models.GetCart(newOrder.CartID)
	if err != nil {
		Sentry.SendErrorToSentry(c, err, tags)
		fmt.Println("cart get err:", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	newOrder.Cart = cart
	var payment models.Payment
	payment.ID = data.GetUUIDString("payment")
	payment.OrderID = newOrder.ID
	payment.Method = newOrder.Payment.Method

	payment.Currency = "INR"

	newOrder.OrderAmount.AddPaymentMethodDiscountAndCalculateDiscount(payment.Method, cart.CartAmount.TotalAmount)
	payment.Amount = int64(newOrder.OrderAmount.TotalAmount)

	newOrder.PaymentStatus = payment.Status

	newOrder.CartID = cart.ID
	newOrder.Cart = cart
	newOrder.Source = (newOrder.Cart.Source)
	newOrder.Payment = payment

	newOrder.Product = cart.Items[0].Product
	newOrder.Variant = cart.Items[0].Variant
	newOrder.VariantID = cart.Items[0].VariantID
	newOrder.Brand = cart.Items[0].Brand
	newOrder.PaymentStatus = "Unpaid"

	err = newOrder.Create()
	if err != nil {
		Sentry.SendErrorToSentry(c, err, tags)
		fmt.Println("order creation err:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if payment.Method == "COD" {
		err = newOrder.MarkOrderAsCompleted(payment)
		if err != nil {
			Sentry.SendErrorToSentry(c, err, tags)
			fmt.Println("payment creation err:", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		err = newOrder.MarkOrderAsCODUnconfirmed()
		if err != nil {
			fmt.Printf("err order: %v\n", err)
			c.AbortWithError(500, err)
			return
		}
		payment, err = payment.Create()
		if err != nil {
			Sentry.SendErrorToSentry(c, err, tags)
			fmt.Println("payment creation err:", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

	} else {
		upiResponse, decentroResponse, err := payments.InitiateUPIPayment(payment.ID, float64(payment.Amount), "Roovo Payment")
		if err != nil {
			Sentry.SendErrorToSentry(c, err, tags)
			fmt.Println("UPI err:", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		payment.ThirdPartyPaymentObject = decentroResponse
		payment, err = payment.Create()
		if err != nil {
			Sentry.SendErrorToSentry(c, err, tags)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		link := upiResponse.UPIDeepLinks.GeneratedLink
		_, err = models.SaveUPIPayObject(link, payment.ID)
		if err != nil {
			Sentry.SendErrorToSentry(c, err, tags)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		upiPay, err := models.GetUPIPayObjectByPayment(payment.ID)
		if err != nil {
			Sentry.SendErrorToSentry(c, err, tags)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		phone := strings.ReplaceAll(address.Phone, "+", "")
		err = whatsapp.SendPaymentLinkMessage(phone, upiPay.ID, payment.Amount)
		if err != nil {
			Sentry.SendErrorToSentry(c, err, tags)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"response": newOrder,
	})
}

func ConfirmCODPayment(ctx *gin.Context) {
	orderID := ctx.Query("orderID")
	order, err := models.GetOrder(orderID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Order ID Invalid"})
		return
	}
	err = order.MarkOrderAsFulfillable()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Error"})
		return
	}
	ctx.JSON(http.StatusOK, order)
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
	orderID := c.Query("id")

	type requestStruct struct {
		Reason string `json:"reason"`
	}

	var requestBody requestStruct
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		fmt.Printf("err: %v\n", err)
		rw.JSONErrorResponse(c, 400, err)
		return
	}

	reason := requestBody.Reason
	if len(reason) < 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cancellation reason needed of min length of 10 characters"})
		return
	}

	// Check Order ID was received in the Request
	if len(orderID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID empty"})
		return
	}

	// Find Order ID in the DB
	order, err := models.GetOrder(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Get SHipment for thi sDB
	shipping, err := models.GetShipmentByID(orderID)
	if err == nil { // if Shipment was found
		// Cancel Shipment on Shiprocket
		shipRocketOrderIDStr := strconv.FormatInt(int64(shipping.OrderId), 10)
		err = shiprocket.CancelShiprocketOrder(shipRocketOrderIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	// Cancel Order in DB
	order.FulfillmentStatus = "Cancelled"
	order.CancellationReason = reason
	curTime := time.Now()
	order.CancellationTime = &curTime
	err = order.Update()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": order})
}

func GetOrdersSummaryForUserByPhone(c *gin.Context) {
	phone := "+" + c.Param("phone")
	fmt.Println("API")
	user, err := models.GetUser(phone)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unable to find user"})
		return
	}
	fmt.Println("API")
	count, err := models.GetOrderCount(user.ID)
	fmt.Printf("err: %v\n", err)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unable to find user"})
		return
	}
	orderValue := int64(0)
	if count > 0 {
		orderValue, err = models.GetTotalOrderValue(user.ID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"ordersCount": count, "ordersValue": orderValue})
}
