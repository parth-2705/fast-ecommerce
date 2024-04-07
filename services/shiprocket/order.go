package shiprocket

import (
	"context"
	"encoding/json"
	"fmt"
	"hermes/admin/services/auth"
	"hermes/db"
	"hermes/models"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tryamigo/themis"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	MAXIMUM_SHIPROCKET_ALLOWED_LENGTH = 198
	YYYYMMDD                          = "2006-01-02"
)

type ShipRocketOrderCreatedResponse struct {
	OrderID     int    `json:"order_id,omitempty"`
	ShipmentID  int    `json:"shipment_id,omitempty"`
	Status      string `json:"status,omitempty"`
	StatusCode  int    `json:"status_code,omitempty"`
	AwbCode     string `json:"awb_code,omitempty"`
	CourierName string `json:"courier_name,omitempty"`
}

type ShipRocketOrder struct {
	OrderID               string                `json:"order_id,omitempty"`        // required
	OrderDate             string                `json:"order_date,omitempty"`      // required
	PickupLocation        string                `json:"pickup_location,omitempty"` // required
	ChannelID             string                `json:"channel_id,omitempty"`
	Comment               string                `json:"comment,omitempty"`
	ResellerName          string                `json:"reseller_name,omitempty"`
	CompanyName           string                `json:"company_name,omitempty"`
	BillingCustomerName   string                `json:"billing_customer_name,omitempty"` // required
	BillingLastName       string                `json:"billing_last_name,omitempty"`
	BillingAddress        string                `json:"billing_address,omitempty"` // required
	BillingAddress2       string                `json:"billing_address_2,omitempty"`
	BillingIsdCode        string                `json:"billing_isd_code,omitempty"`
	BillingCity           string                `json:"billing_city,omitempty"`    // required
	BillingPincode        string                `json:"billing_pincode,omitempty"` // required
	BillingState          string                `json:"billing_state,omitempty"`   // required
	BillingCountry        string                `json:"billing_country,omitempty"` // required
	BillingEmail          string                `json:"billing_email,omitempty"`   // required
	BillingPhone          string                `json:"billing_phone,omitempty"`   // required
	BillingAlternatePhone string                `json:"billing_alternate_phone,omitempty"`
	ShippingIsBilling     bool                  `json:"shipping_is_billing,omitempty"` // required
	ShippingCustomerName  string                `json:"shipping_customer_name,omitempty"`
	ShippingLastName      string                `json:"shipping_last_name,omitempty"`
	ShippingAddress       string                `json:"shipping_address,omitempty"`
	ShippingAddress2      string                `json:"shipping_address_2,omitempty"`
	ShippingCity          string                `json:"shipping_city,omitempty"`
	ShippingPincode       string                `json:"shipping_pincode,omitempty"`
	ShippingCountry       string                `json:"shipping_country,omitempty"`
	ShippingState         string                `json:"shipping_state,omitempty"`
	ShippingEmail         string                `json:"shipping_email,omitempty"`
	ShippingPhone         string                `json:"shipping_phone,omitempty"`
	OrderItems            []ShipRocketOrderItem `json:"order_items,omitempty"`    // required
	PaymentMethod         string                `json:"payment_method,omitempty"` // required
	ShippingCharges       string                `json:"shipping_charges,omitempty"`
	GiftwrapCharges       string                `json:"giftwrap_charges,omitempty"`
	TransactionCharges    string                `json:"transaction_charges,omitempty"`
	TotalDiscount         string                `json:"total_discount,omitempty"`
	SubTotal              string                `json:"sub_total,omitempty"` // required
	Length                string                `json:"length,omitempty"`    // required
	Breadth               string                `json:"breadth,omitempty"`   // required
	Height                string                `json:"height,omitempty"`    // required
	Weight                string                `json:"weight,omitempty"`    // required
	EwaybillNo            string                `json:"ewaybill_no,omitempty"`
	CustomerGstin         string                `json:"customer_gstin,omitempty"`
	InvoiceNumber         string                `json:"invoice_number,omitempty"`
	OrderType             string                `json:"order_type,omitempty"`
}

type ShipRocketOrderItem struct {
	Name         string `json:"name,omitempty"`          // required
	Sku          string `json:"sku,omitempty"`           // required
	Units        string `json:"units,omitempty"`         // required
	SellingPrice string `json:"selling_price,omitempty"` // required
	Discount     string `json:"discount,omitempty"`      // required
	Tax          string `json:"tax,omitempty"`           // required, total percentage discount
	Hsn          string `json:"hsn,omitempty"`
}

func addStringsAsFloats(num1Str string, num2Str string) (sumStr string, err error) {

	var num1, num2, sum float64

	num1, err = strconv.ParseFloat(num1Str, 64)
	if err != nil {
		return
	}

	num2, err = strconv.ParseFloat(num2Str, 64)
	if err != nil {
		return
	}

	sum = num1 + num2

	sumStr = strconv.FormatFloat(sum, 'f', 2, 64)

	return
}

func multiplyStringAsFloat(numStr string, multiply int) (finalStr string, err error) {

	var num float64

	num, err = strconv.ParseFloat(numStr, 64)
	if err != nil {
		return
	}

	final := num * float64(multiply)

	finalStr = strconv.FormatFloat(final, 'f', 2, 64)

	return

}

func maxStringAsFloats(num1Str, num2Str string) (maxStr string, err error) {
	var num1, num2, max float64

	num1, err = strconv.ParseFloat(num1Str, 64)
	if err != nil {
		return
	}

	num2, err = strconv.ParseFloat(num2Str, 64)
	if err != nil {
		return
	}

	max = math.Max(num1, num2)

	maxStr = strconv.FormatFloat(max, 'f', 2, 64)

	return
}

type SingleOrderDependantShipRocketOrderInfo struct {
	OrderItems []ShipRocketOrderItem
	Length     string
	Height     string
	Breadth    string
	Weight     string
	Cost       float64
}

type SingleSKUDependantShipRocketOrderInfo struct {
	OrderItem ShipRocketOrderItem
	Length    string
	Height    string
	Breadth   string
	Weight    string
	Cost      float64
}

func ProcessSingleSKUForShipment(item models.Item, cartAmount models.OrderAmount) (data SingleSKUDependantShipRocketOrderInfo, err error) {
	var shipRocketOrderItem ShipRocketOrderItem

	// Shiprocket has a maximum product name size constraint. So if required we cut the product name short.
	shipRocketOrderItem.Name = item.Product.Name[:int(math.Min(MAXIMUM_SHIPROCKET_ALLOWED_LENGTH, float64(len(item.Product.Name))))]

	shipRocketOrderItem.Sku = item.Variant.SKU
	shipRocketOrderItem.Units = strconv.FormatInt(int64(item.Quantity), 10)
	shipRocketOrderItem.SellingPrice = fmt.Sprint(item.Variant.Price.SellingPrice)
	// shipRocketOrderItem.Discount = fmt.Sprint(order.Variant.Price.Discount)
	shipRocketOrderItem.Discount = fmt.Sprint(cartAmount.Coupon.DiscountAmount)
	shipRocketOrderItem.Tax = getTaxFromBarcode(item.Variant.Barcode)

	data.OrderItem = shipRocketOrderItem

	// data.Cost = item.Variant.Price.SellingPrice * float64(item.Quantity)
	data.Cost = cartAmount.TotalAmount

	// Weight is multiplied by quantity
	data.Weight, err = multiplyStringAsFloat(item.Variant.Weight, item.Quantity)
	if err != nil {
		return
	}

	// The Volumetric Weight is calculated as such the objects are stacked on height basis
	data.Height, err = multiplyStringAsFloat(item.Variant.Height, item.Quantity)
	if err != nil {
		return
	}

	// Breadth will be same for all the products
	// data.Breadth, err = multiplyStringAsFloat(item.Variant.Breadth, item.Quantity)
	data.Breadth = item.Variant.Breadth

	// Length will be same for all the products
	// data.Length, err = multiplyStringAsFloat(item.Variant.Length, item.Quantity)
	data.Length = item.Variant.Length

	return
}

func ProcessSingleOrderForShipment(order models.Order) (data SingleOrderDependantShipRocketOrderInfo, err error) {
	data.Length = "0"
	data.Breadth = "0"
	data.Height = "0"
	data.Weight = "0"

	for _, item := range order.Cart.Items {
		var singleSKU SingleSKUDependantShipRocketOrderInfo
		singleSKU, err = ProcessSingleSKUForShipment(item, order.Cart.CartAmount)
		if err != nil {
			return
		}

		data.OrderItems = append(data.OrderItems, singleSKU.OrderItem)

		data.Cost += singleSKU.Cost

		// Packing for Multiple Items
		// Wieights are added
		data.Weight, err = addStringsAsFloats(data.Weight, singleSKU.Weight)
		if err != nil {
			return
		}

		// Heights are added as if stacked
		data.Height, err = addStringsAsFloats(data.Height, singleSKU.Height)
		if err != nil {
			return
		}

		// Maximum length of the two is taken
		// data.Length, err = addStringsAsFloats(data.Length, singleSKU.Length)
		data.Length, err = maxStringAsFloats(data.Length, singleSKU.Length)
		if err != nil {
			return
		}

		// Maximum Breadth of the two is taken
		// data.Breadth, err = addStringsAsFloats(data.Breadth, singleSKU.Breadth)
		data.Breadth, err = maxStringAsFloats(data.Breadth, singleSKU.Breadth)
		if err != nil {
			return
		}

	}

	return
}

func CreateShipRocketOrderRequestBody(order models.Order, seller models.Seller) (ShipRocketOrder, error) {

	var shipRocketOrder ShipRocketOrder

	now := time.Now().UTC()

	shipRocketOrder.OrderID = order.ID
	shipRocketOrder.OrderDate = fmt.Sprint(now.Format(YYYYMMDD))
	shipRocketOrder.PickupLocation = seller.ID

	shipRocketOrder.BillingCustomerName = order.Address.Name
	shipRocketOrder.BillingLastName = order.Address.Name
	shipRocketOrder.BillingAddress = order.Address.HouseArea
	shipRocketOrder.BillingCity = order.Address.City
	shipRocketOrder.BillingPincode = order.Address.PinCode
	shipRocketOrder.BillingState = order.Address.State
	shipRocketOrder.BillingCountry = "India"
	// shipRocketOrder.BillingEmail = ""
	shipRocketOrder.BillingPhone = strings.Replace(order.Address.Phone, "+91", "", -1)

	shipRocketOrder.ShippingIsBilling = true

	var paymentMethod string
	if order.Payment.Method == "COD" {
		paymentMethod = "COD"
	} else {
		paymentMethod = "Prepaid"
	}

	shipRocketOrder.PaymentMethod = paymentMethod

	singleOrder, err := ProcessSingleOrderForShipment(order)
	if err != nil {
		return shipRocketOrder, err
	}

	shipRocketOrder.OrderItems = append(shipRocketOrder.OrderItems, singleOrder.OrderItems...)
	shipRocketOrder.SubTotal = fmt.Sprint(singleOrder.Cost)
	shipRocketOrder.Length = singleOrder.Length
	shipRocketOrder.Height = singleOrder.Height
	shipRocketOrder.Breadth = singleOrder.Breadth
	shipRocketOrder.Weight = singleOrder.Weight

	// var subTotal float64 = 0
	// shipRocketOrder.Length = "0"
	// shipRocketOrder.Breadth = "0"
	// shipRocketOrder.Height = "0"
	// shipRocketOrder.Weight = "0"

	// for _, item := range order.Cart.Items {
	// 	var shipRocketOrderItem ShipRocketOrderItem
	// 	shipRocketOrderItem.Name = item.Product.Name
	// 	shipRocketOrderItem.Sku = item.Variant.SKU
	// 	shipRocketOrderItem.Units = strconv.FormatInt(int64(item.Quantity), 10)
	// 	shipRocketOrderItem.SellingPrice = fmt.Sprint(item.Variant.Price.SellingPrice)
	// 	// shipRocketOrderItem.Discount = fmt.Sprint(order.Variant.Price.Discount)
	// 	shipRocketOrderItem.Discount = "0"
	// 	shipRocketOrderItem.Tax = getTaxFromBarcode(item.Variant.Barcode)

	// 	shipRocketOrder.OrderItems = append(shipRocketOrder.OrderItems, shipRocketOrderItem)

	// 	subTotal += item.Variant.Price.SellingPrice * float64(item.Quantity)

	// 	shipRocketOrder.Length, _ = multiplyStringAsFloat(item.Variant.Length, item.Quantity)
	// 	if err != nil {
	// 		return shipRocketOrder, err
	// 	}

	// 	shipRocketOrder.Height, _ = multiplyStringAsFloat(item.Variant.Height, item.Quantity)
	// 	if err != nil {
	// 		return shipRocketOrder, err
	// 	}

	// 	shipRocketOrder.Breadth, _ = multiplyStringAsFloat(item.Variant.Breadth, item.Quantity)
	// 	if err != nil {
	// 		return shipRocketOrder, err
	// 	}

	// 	shipRocketOrder.Weight, _ = multiplyStringAsFloat(item.Variant.Weight, item.Quantity)
	// 	if err != nil {
	// 		return shipRocketOrder, err
	// 	}

	// }

	// shipRocketOrder.SubTotal = fmt.Sprint(subTotal)

	// var shipRocketOrderItem ShipRocketOrderItem
	// shipRocketOrderItem.Name = order.Product.Name
	// shipRocketOrderItem.Sku = order.Variant.SKU
	// shipRocketOrderItem.Units = "1"
	// shipRocketOrderItem.SellingPrice = fmt.Sprint(order.Variant.Price.SellingPrice)
	// // shipRocketOrderItem.Discount = fmt.Sprint(order.Variant.Price.Discount)
	// shipRocketOrderItem.Discount = "0"
	// shipRocketOrderItem.Tax = getTaxFromBarcode(order.Variant.Barcode)

	// shipRocketOrder.OrderItems = append(shipRocketOrder.OrderItems, shipRocketOrderItem)
	// shipRocketOrder.SubTotal = fmt.Sprint(order.Product.Price.SellingPrice)

	return shipRocketOrder, nil
}

func CreateShipRocketOrder(order models.Order, seller models.Seller, parentID string, relation int, courierID string) (models.Shipping, error) {

	if len(order.Payment.Method) == 0 {
		return models.Shipping{}, fmt.Errorf("order Payment Method is not defined")
	}

	shiprocketOrder, err := CreateShipRocketOrderRequestBody(order, seller)
	if err != nil {
		return models.Shipping{}, err
	}
	requestBody, _ := json.Marshal(shiprocketOrder)

	token, err := GetShipRocketToken()
	if err != nil {
		return models.Shipping{}, err
	}

	url := "https://apiv2.shiprocket.in/v1/external/orders/create/adhoc"
	headers := [][]string{{"Authorization", "Bearer " + token}}

	statusCode, responseBody, err := themis.HitAPIEndpoint2(url, http.MethodPost, requestBody, headers, [][]string{})
	if err != nil {
		return models.Shipping{}, err
	}
	if statusCode >= 400 {
		return models.Shipping{}, fmt.Errorf("erorr: %s with status code: %d", string(responseBody), statusCode)
	}

	var shipRocketOrderCreatedResponse ShipRocketOrderCreatedResponse
	err = json.Unmarshal(responseBody, &shipRocketOrderCreatedResponse)
	if err != nil {
		return models.Shipping{}, err
	}

	awbCode, err := CreateShipRocketOrderShipment(fmt.Sprint(shipRocketOrderCreatedResponse.ShipmentID), headers, courierID)
	if err != nil {
		return models.Shipping{}, err
	}

	labelURL, err := CreateShipRocketShipmentLabel(fmt.Sprint(shipRocketOrderCreatedResponse.ShipmentID), headers)
	if err != nil {
		return models.Shipping{}, err
	}

	invoiceURL, err := CreateShipRocketShipmentInvoice(fmt.Sprint(shipRocketOrderCreatedResponse.OrderID), headers)
	if err != nil {
		return models.Shipping{}, err
	}

	shipment, err := CreateShipmentInDatabase(shipRocketOrderCreatedResponse, labelURL, invoiceURL, awbCode, order, parentID, relation)
	if err != nil {
		return models.Shipping{}, err
	}

	go UpdateShiprocketOrderUtil(shipRocketOrderCreatedResponse.OrderID)

	return shipment, nil
}

func CreateShipmentInDatabase(shipRocketOrderCreatedResponse ShipRocketOrderCreatedResponse, labelURL string, invoiceURL string, awbCode string, order models.Order, parentID string, relation int) (models.Shipping, error) {
	shipping := models.Shipping{
		Id:              order.ID,
		OrderId:         shipRocketOrderCreatedResponse.OrderID,
		ShippingId:      shipRocketOrderCreatedResponse.ShipmentID,
		AWB:             awbCode,
		LabelURL:        labelURL,
		InvoiceURL:      invoiceURL,
		BrandID:         order.Product.BrandID,
		SellerID:        order.Product.SellerID,
		ShipmentCreated: false,
		ParentOrderID:   parentID,
		ParentRelation:  relation,
	}

	err := shipping.Insert()
	if err != nil {
		return shipping, err
	}

	return shipping, nil
}

func getTaxFromBarcode(barcode string) string {
	return "18"
}

func CreateOrderOnShipRocketUtil(c *gin.Context, seller models.Seller, orderID string, courierID string) {

	var order models.Order
	err := db.OrderCollection.FindOne(context.Background(), bson.M{"_id": orderID}).Decode(&order)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("order not found %s", err.Error())})
		fmt.Printf("err: %v\n", err)
		return
	}

	shipment, err := CreateShipRocketOrder(order, seller, order.ID, models.Parent, courierID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("couldn't process order %s", err.Error())})
		fmt.Printf("err: %v\n", err)
		return
	}

	shippingCharges, err := GetOrderShippingCharges(order.ID, fmt.Sprint(shipment.OrderId))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("couldn't fetch shippingCharges %s", err.Error())})
		fmt.Printf("err: %v\n", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": shippingCharges})

}

func CreateOrderOnShipRocket(c *gin.Context) {

	seller, err := auth.GetSellerFromSession(c)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	orderID := c.Query("id")
	courierID := c.Query("courierID")

	if len(orderID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order id not found"})
		return
	}

	CreateOrderOnShipRocketUtil(c, seller, orderID, courierID)
}

func CreateAdminOrderOnShipRocket(c *gin.Context) {
	orderID := c.Query("id")
	sellerID := c.Query("sellerID")
	courierID := c.Query("courierID")
	if len(orderID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order id not found"})
		return
	}

	order, err := models.GetOrder(orderID)
	if err != nil || !order.Fulfillable {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order not fulfillable"})
		return
	}

	var seller models.Seller
	err = db.SellerCollection.FindOne(context.Background(), bson.M{"_id": sellerID}).Decode(&seller)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Seller id not found"})
		return
	}

	CreateOrderOnShipRocketUtil(c, seller, orderID, courierID)
}

func CreateSellerOrderOnShipRocket(c *gin.Context) {
	orderID := c.Query("id")
	courierID := c.Query("courierID")
	seller, err := auth.GetSellerFromSession(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Seller not logged in"})
		return
	}
	if len(orderID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order id not found"})
		return
	}

	CreateOrderOnShipRocketUtil(c, seller, orderID, courierID)
}
