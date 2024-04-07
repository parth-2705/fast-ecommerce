package shiprocket

import (
	"context"
	"encoding/json"
	"fmt"
	"hermes/db"
	"hermes/models"
	"hermes/utils/data"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tryamigo/themis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type QuantityInfo struct {
	Quantity     int    `json:"quantity" bson:"quantity"`
	SKU          string `json:"sku" bson:"sku"`
	Name         string `json:"name" bson:"name"`
	SellingPrice int    `json:"sellingPrice" bson:"sellingPrice"`
	ProductID    string `json:"productID" bson:"productID"`
}

func CreateReturnShipRocketOrderShipmentUtil(shippingID string, headers [][]string) (awb string, err error) {
	reqBody := map[string]string{
		"shipment_id": shippingID,
		"is_return":   "1",
	}
	requestBody, _ := json.Marshal(reqBody)
	url := "https://apiv2.shiprocket.in/v1/external/courier/assign/awb"
	statusCode, responseBody, err := themis.HitAPIEndpoint2(url, http.MethodPost, requestBody, headers, [][]string{})
	if err != nil {
		return "", err
	}

	if statusCode >= 400 {
		return "", fmt.Errorf("erorr: %s with status code: %d", string(responseBody), statusCode)
	}

	var ShiprocketAWBAssignStatus ShiprocketAWBAssignStatus
	err = json.Unmarshal(responseBody, &ShiprocketAWBAssignStatus)
	if err != nil {
		return "", err
	}

	if ShiprocketAWBAssignStatus.AwbAssignStatus == 0 {
		return "", fmt.Errorf("erorr: %s with status code: %d", string(responseBody), http.StatusBadRequest)
	}

	var ShiprocketAWBResponse ShiprocketAWBResponse
	err = json.Unmarshal(responseBody, &ShiprocketAWBResponse)
	if err != nil {
		return "", err
	}

	return ShiprocketAWBResponse.Response.Data.AwbCode, nil
}

func CreateNewShipmentUtil(orderID string, quantityInfo map[string]QuantityInfo, shipmentType int) (err error) {
	order, err := models.GetOrder(orderID)
	if err != nil {
		return
	}
	seller, err := models.GetSellerByID(order.Product.SellerID)
	if err != nil {
		return
	}
	var productInfo []models.ProductInfoItem
	var prefix string = ""
	if shipmentType == models.Missing {
		prefix = "MI_"
	} else if shipmentType == models.Replacement {
		prefix = "RE_"
	} else {
		prefix = "EX_"
	}
	for key, val := range quantityInfo {
		productInfo = append(productInfo, models.ProductInfoItem{
			ProductID: val.ProductID,
			VariantID: key,
			SKU:       val.SKU,
			Quantity:  val.Quantity,
		})
	}

	var newCart models.Cart = order.Cart
	newCart.Items = []models.Item{}
	var newTotal = 0
	for _, val := range order.Cart.Items {
		if quantityItem, ok := quantityInfo[val.VariantID]; ok {
			val.Quantity = quantityItem.Quantity
			newTotal = quantityItem.Quantity * int(val.Variant.Price.SellingPrice)
			newCart.Items = append(newCart.Items, val)
		}
	}
	newCart.CartAmount.TotalAmount = float64(newTotal)
	order.ID = data.GetUUIDString(prefix + "Order")
	order.PaymentStatus = "Paid"
	order.Payment.Method = "Prepaid"
	order.Cart = newCart
	order.ParentID = orderID
	order.Source = models.BackwardShip
	err = order.Create()
	if err != nil {
		return
	}
	_, err = CreateShipRocketOrder(order, seller, orderID, shipmentType, "")
	if err != nil {
		return
	}
	err = AddToExistingOrCreateNewBackwardShipment(orderID, order.ID, shipmentType, productInfo)
	return
}

func CreateReturnUtil(orderID string, quantityInfo map[string]QuantityInfo) (err error) {

	shipment, err := models.GetShipmentByID(orderID)
	if err != nil {
		return fmt.Errorf("Unable to get shipment %s", err.Error())
	}

	order, err := models.GetOrder(orderID)
	if err != nil {
		return fmt.Errorf("Unable to get order %s", err.Error())
	}
	seller, err := models.GetSellerByID(order.Product.SellerID)
	if err != nil {
		return fmt.Errorf("Unable to get seller %s", err.Error())
	}

	token, err := GetShipRocketToken()
	if err != nil {
		return fmt.Errorf("Unable to get shiprocket token %s", err.Error())
	}

	var orderItems []map[string]interface{}
	var productInfo []models.ProductInfoItem

	var total int = 0

	for key, val := range quantityInfo {
		orderItems = append(orderItems, map[string]interface{}{
			"name":          val.Name,
			"sku":           val.SKU,
			"units":         val.Quantity,
			"selling_price": val.SellingPrice,
			"discount":      0,
		})
		productInfo = append(productInfo, models.ProductInfoItem{
			ProductID: val.ProductID,
			VariantID: key,
			SKU:       val.SKU,
			Quantity:  val.Quantity,
		})
		total = total + (val.SellingPrice * val.Quantity)
	}

	headers := [][]string{{"Authorization", "Bearer " + token}}
	reqBodyJSON := map[string]interface{}{
		"order_id":               "RR_" + shipment.Id,
		"channel_id":             3639343,
		"order_date":             time.Now().Format("2006-01-02"),
		"pickup_customer_name":   order.Address.Name,
		"pickup_address":         order.Address.HouseArea,
		"pickup_city":            order.Address.City,
		"pickup_state":           order.Address.State,
		"pickup_country":         "India",
		"pickup_pincode":         order.Address.PinCode,
		"pickup_phone":           strings.ReplaceAll(order.Address.Phone, "+91", ""),
		"shipping_customer_name": order.Product.Brand.Name,
		"shipping_address":       seller.HouseArea + ", " + seller.StreetName,
		"shipping_city":          seller.City,
		"shipping_country":       "India",
		"shipping_pincode":       seller.PinCode,
		"shipping_state":         seller.State,
		"shipping_email":         seller.Email,
		"pickup_email":           "",
		"shipping_phone":         strings.ReplaceAll(seller.Phone, "+91", ""),
		"order_items":            orderItems,
		"payment_method":         order.Payment.Method,
		"sub_total":              total,
		"length":                 order.Variant.Length,
		"breadth":                order.Variant.Breadth,
		"height":                 order.Variant.Height,
		"weight":                 order.Variant.Weight,
	}

	reqBody, _ := json.Marshal(reqBodyJSON)
	statusCode, response, err := themis.HitAPIEndpoint2("https://apiv2.shiprocket.in/v1/external/orders/create/return", http.MethodPost, reqBody, headers, [][]string{})
	if err != nil {
		return fmt.Errorf("Unable to make api call %s", err.Error())
	}
	if statusCode >= 400 {
		return fmt.Errorf("API Failed with error code %s: %s ", statusCode, string(response))
	}
	var respBody ReturnOrderResponse
	json.Unmarshal(response, &respBody)

	awb, err := CreateReturnShipRocketOrderShipmentUtil(fmt.Sprint(respBody.ShipmentID), headers)
	if err != nil {
		return err
	}

	labelUrl, err := CreateShipRocketShipmentLabel(fmt.Sprint(respBody.ShipmentID), headers)
	if err != nil {
		return err
	}

	var returnShipment models.Shipping = models.Shipping{
		Id:             "RR_" + shipment.Id,
		OrderId:        respBody.OrderID,
		ShippingId:     respBody.ShipmentID,
		SellerID:       shipment.SellerID,
		BrandID:        shipment.BrandID,
		ParentOrderID:  shipment.Id,
		ParentRelation: models.Return,
		LabelURL:       labelUrl,
		InvoiceURL:     "",
		AWB:            awb,
		Dispatched:     false,
		Processed:      false,
		Remitted:       false,
	}

	err = returnShipment.Insert()
	if err != nil {
		return err
	}

	err = AddToExistingOrCreateNewBackwardShipment(shipment.Id, "RR_"+shipment.Id, models.Return, productInfo)
	if err != nil {
		return err
	}

	return nil
}

func AddToExistingOrCreateNewBackwardShipment(parentID string, childID string, backwardType int, productInfo []models.ProductInfoItem) (err error) {
	for idx, val := range productInfo {
		var prod models.Product
		var vari models.Variation
		prod, err = models.GetCompleteProduct(val.ProductID)
		if err != nil {
			return
		}
		vari, err = models.GetVariant(val.VariantID)
		if err != nil {
			return
		}
		productInfo[idx].Product = prod
		productInfo[idx].Variant = vari
	}
	var backwardShipment models.BackwardShipment
	var exists bool = true
	err = db.BackwardShipmentCollection.FindOne(context.Background(), bson.M{"_id": parentID}).Decode(&backwardShipment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			exists = false
		} else {
			return
		}
	}

	if exists {
		backwardShipment.ChildShipments = append(backwardShipment.ChildShipments, models.ChildShipment{
			Type:        backwardType,
			ShipmentID:  childID,
			ProductInfo: productInfo,
			Status:      "Fulfilled",
		})
		err = backwardShipment.Update()
		if err != nil {
			return
		}
	} else {
		backwardShipment.ID = parentID
		backwardShipment.ChildShipments = []models.ChildShipment{{
			Type:        backwardType,
			ShipmentID:  childID,
			ProductInfo: productInfo,
			Status:      "Fulfilled",
		}}
		err = backwardShipment.Create()
		if err != nil {
			return
		}
	}
	return nil
}

func CreateReturnForOrderOnShiprocket(c *gin.Context) {
	orderID := c.Query("id")
	if len(orderID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Empty order id",
		})
		return
	}
	var quantityInfo map[string]QuantityInfo
	err := c.BindJSON(&quantityInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = CreateReturnUtil(orderID, quantityInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "success"})
}

func CreateShipmentForMissingItems(c *gin.Context) {
	orderID := c.Query("id")
	if len(orderID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Empty order id",
		})
		return
	}
	var quantityInfo map[string]QuantityInfo
	c.BindJSON(&quantityInfo)
	err := CreateNewShipmentUtil(orderID, quantityInfo, models.Missing)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "success"})
}

func CreateShipmentForReplacementItems(c *gin.Context) {
	orderID := c.Query("id")
	if len(orderID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Empty order id",
		})
		return
	}
	var quantityInfo map[string]QuantityInfo
	c.BindJSON(&quantityInfo)
	err := CreateNewShipmentUtil(orderID, quantityInfo, models.Replacement)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "success"})
}

func CreateExchangeForOrderOnShiprocket(c *gin.Context) {
	orderID := c.Query("id")
	if len(orderID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Empty order id",
		})
		return
	}
	var quantityInfo map[string]QuantityInfo
	c.BindJSON(&quantityInfo)
	err := CreateReturnUtil(orderID, quantityInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to create return " + err.Error()})
		return
	}
	err = CreateNewShipmentUtil(orderID, quantityInfo, models.Exchange)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to create new shipment " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "success"})
}

func CancelReturnOrder(c *gin.Context) {
	id := c.Query("id")
	shiprocketOrderID := c.Query("shiprocketOrderID")

	if len(id) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Return Shipment ID empty"})
		return
	}

	if len(shiprocketOrderID) > 0 {
		err := CancelShiprocketOrder(shiprocketOrderID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	backwardShipment, err := models.GetReturnShipmentByChildID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	for idx, val := range backwardShipment.ChildShipments {
		if val.ShipmentID == id {
			temp := val
			temp.Status = "Cancelled"
			backwardShipment.ChildShipments[idx] = temp
		}
	}

	backwardShipment.Update()

	c.JSON(http.StatusOK, gin.H{"response": backwardShipment})
}
