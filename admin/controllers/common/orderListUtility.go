package common

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"hermes/db"
	"hermes/models"
	"hermes/services/shiprocket"
	"hermes/utils/tmpl"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AdminOrder struct {
	models.Order                   `bson:",inline"`
	Seller                         models.Seller              `json:"seller" bson:"seller"`
	Shipping                       models.Shipping            `json:"shipment" bson:"shipment"`
	Tracking                       models.ShipRocketTracking  `json:"tracking" bson:"tracking"`
	HasPickUpError                 bool                       `json:"pickup-error" bson:"pickup-error"`
	Couriers                       []models.ShiprocketCourier `json:"couriers" bson:"couriers"`
	Recommendedcouriercompanyid    int                        `json:"recommendedcouriercompanyid" bson:"recommendedcouriercompanyid"`
	Shiprocketrecommendedcourierid int                        `json:"shiprocketrecommendedcourierid" bson:"shiprocketrecommendedcourierid"`
	ShiprocketOrder                models.FullShiprocketOrder `json:"shiprocketOrders" bson:"shiprocketOrders"`
}

type DistinctStatuses struct {
	Statuses []string `json:"statuses" bson:"statuses"`
}

func OrderFilterUtility(currentSeller string, currentBrand string, ordersStartDate string, ordersEndDate string, shipmentStatus string, awbs string, sellerAPI bool) ([]AdminOrder, []string, error) {

	// Find all distinct values of the field.

	var orders []AdminOrder
	var shipmentStatuses []string

	var distinctStatuses []DistinctStatuses
	cur, err := db.TrackingCollection.Aggregate(context.Background(), mongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.M{"_id": primitive.Null{}, "statuses": bson.M{"$addToSet": "$currentstatus"}}}},
		bson.D{{Key: "$match", Value: bson.M{"statuses": bson.M{"$nin": shiprocket.SHIPROCKET_STATUSES}}}},
	})
	if err != nil {
		return orders, shipmentStatuses, fmt.Errorf("Failed to find distinct values: " + err.Error())
	}

	err = cur.All(context.Background(), &distinctStatuses)
	if err != nil {
		return orders, shipmentStatuses, fmt.Errorf("Failed to read distinct values: " + err.Error())
	}
	shipmentStatuses = append(shipmentStatuses, shiprocket.SHIPROCKET_BUCKETS...)

	if len(distinctStatuses) > 0 && len(distinctStatuses[0].Statuses) > 0 {
		shipmentStatuses = append(shipmentStatuses, distinctStatuses[0].Statuses...)
	}

	paymentFilter := []string{"Paid", "Succeeded"}

	// Define the pipeline
	pipeline := mongo.Pipeline{
		// Filter orders by sellerID
		bson.D{{Key: "$sort", Value: bson.D{{Key: "createdAt", Value: -1}}}},
		bson.D{{Key: "$match", Value: bson.D{{Key: "paymentStatus", Value: bson.D{{Key: "$in", Value: paymentFilter}}}}}},
		bson.D{{Key: "$addFields", Value: bson.D{{Key: "internal", Value: "$user.internal"}}}},
		// Filter orders by internal = false
		bson.D{{Key: "$match", Value: bson.D{{Key: "$or", Value: bson.A{
			bson.D{{Key: "internal", Value: false}},
			bson.D{{Key: "internal", Value: bson.D{{Key: "$exists", Value: false}}}},
		}}}}},
		// Seller
		bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "seller"}, {Key: "localField", Value: "product.sellerID"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "seller"}}}},
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$seller"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		// Shipping
		bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "shipping"}, {Key: "localField", Value: "_id"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "shipment"}}}},
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$shipment"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		bson.D{{Key: "$addFields", Value: bson.D{{Key: "awb", Value: "$shipment.awb"}}}},
		// Tracking
		bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "tracking"}, {Key: "localField", Value: "awb"}, {Key: "foreignField", Value: "awb"}, {Key: "as", Value: "tracking"}}}},
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$tracking"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		// Shipping Charges
		bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "shippingCharges"}, {Key: "localField", Value: "_id"}, {Key: "foreignField", Value: "orderID"}, {Key: "as", Value: "charges"}}}},
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$charges"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		bson.D{{Key: "$addFields", Value: bson.D{{Key: "couriers", Value: "$charges.data.availablecouriercompanies"},
			{Key: "recommendedcouriercompanyid", Value: "$charges.data.recommendedcouriercompanyid"},
			{Key: "shiprocketrecommendedcourierid", Value: "$charges.data.shiprocketrecommendedcourierid"}}}},
		// Shiprocket Orders
		bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "shiprocketOrders"}, {Key: "localField", Value: "shipment.orderId"}, {Key: "foreignField", Value: "srid"}, {Key: "as", Value: "shiprocketOrders"}}}},
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$shiprocketOrders"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
	}

	if len(ordersStartDate) > 0 && len(ordersEndDate) > 0 {
		loc, _ := time.LoadLocation("Asia/Calcutta")
		filterStartDate, err := time.ParseInLocation("2006-01-02", ordersStartDate, loc)
		if err != nil {
			fmt.Println("err:", err.Error())
		}

		filterEndDate, err := time.ParseInLocation("2006-01-02", ordersEndDate, loc)

		if err != nil {
			fmt.Println("err:", err.Error())
		}

		pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.M{
			"$expr": bson.M{
				"$and": []bson.M{
					{
						"$gte": []interface{}{"$createdAt", filterStartDate},
					},
					{
						"$lt": []interface{}{"$createdAt", filterEndDate.Add(24 * time.Hour)},
					},
				},
			},
		}}})
	}

	if sellerAPI {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.D{{Key: "fulfillable", Value: true}}}})
	}

	if len(currentSeller) > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.D{{Key: "product.sellerID", Value: currentSeller}}}})
	}

	if len(currentBrand) > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.D{{Key: "product.brandID", Value: currentBrand}}}})
	}

	nonCancelledOrders := bson.D{{Key: "$match", Value: bson.D{{Key: "fulfillmentStatus", Value: bson.D{{Key: "$ne", Value: "Cancelled"}}}}}}

	if len(shipmentStatus) > 0 && shipmentStatus != "all" {
		if shipmentStatus == "pending-fulfillment" {
			pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.D{{Key: "awb", Value: bson.D{{Key: "$exists", Value: false}}}}}})
		} else if shipmentStatus == "ship-now" {
			pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.D{{Key: "shipment.shipmentCreated", Value: false}}}})
		} else {
			if shipmentStatus == "Pickups" {
				pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.D{{Key: "awb", Value: bson.M{"$exists": true}}}}})
				pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.D{{Key: "$or", Value: bson.A{
					bson.D{{Key: "tracking.currentstatus", Value: bson.M{"$in": shiprocket.SHIPROCKET_ORDER_TRACKING_BUCKET[shipmentStatus]}}},
					bson.D{{Key: "tracking", Value: bson.M{"$exists": false}}},
				}}}}})
			} else {
				pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.D{{Key: "tracking.currentstatus", Value: bson.M{"$in": shiprocket.SHIPROCKET_ORDER_TRACKING_BUCKET[shipmentStatus]}}}}})
			}
		}
		pipeline = append(pipeline, nonCancelledOrders)
	}

	if len(awbs) > 0 {
		allAWBS := strings.Split(awbs, ",")

		filterAdded := false

		// create a filter that matches any of the search strings with any of the fields
		searchFilters := make([]bson.M, 0)

		for _, awbWithWhiteSpace := range allAWBS {
			q := strings.TrimSpace(awbWithWhiteSpace)
			if len(q) > 0 {
				filterAdded = true
				filter := bson.M{"$regex": q, "$options": "i"}
				filters := []bson.M{
					{"_id": filter},
					{"shipment.awb": filter},
					{"address.phone": filter},
					{"address.name": filter},
					{"variant.sku": filter},
				}
				searchFilters = append(searchFilters, filters...)
			}
		}

		if filterAdded {
			pipeline = append(pipeline, bson.D{{
				Key: "$match", Value: bson.M{
					"$or": searchFilters,
				},
			}})
		}
	}

	cur, err = db.OrderCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return orders, shipmentStatuses, fmt.Errorf("Order not found" + err.Error())
	}

	err = cur.All(context.Background(), &orders)
	if err != nil {
		return orders, shipmentStatuses, fmt.Errorf("Order not correct format" + err.Error())
	}

	defer cur.Close(context.Background())

	return orders, shipmentStatuses, nil
}

func DownloadOrderReportUtil(orders []AdminOrder) ([]byte, error) {
	var report [][]string

	headers := []string{
		"Order ID",
		"Shiprocket Created At",
		"SKU",
		"Product Name",
		"Product Quantity",
		"order Created At",
		"Customer Name",
		"Customer Email",
		"Customer Mobile",
		"Address Line 1",
		"Address City",
		"Address State",
		"Address Pincode",
		"Payment Method",
		"Payment Status",
		"Product Price",
		"Order Total",
		"Tax",
		"Tax %",
		"Discount Value",
		"Product HSN",
		"Weight (KG)",
		"dimensions (CM)",
		"Charged Weight",
		"Courier Company",
		"AWB Code",
		"Shipping Bill URL",
		"AWB Assigned Date",
		"Pickup Location ID",
		"Pickup Address Name",
		"Pickup scheduled Date",
		"Order Picked Up Date",
		"Pickup First Attempt Date",
		"Pickedup Timestamp",
		"Order Shipped Date",
		"EDD Delayed Reason",
		"Order Delivered Date",
		"RTO Address",
		"RTO Initiated Date",
		"RTO Delivered Date",
		"COD Payble Amount",
		"UTR No",
		"COD Charges",
		"RTO Charges",
		"Freight Total Amount",
		"Total Shipping Charges",
		"Transaction charges",
		"Customer_invoice_id",
		"Pickup Exception Reason",
		"First Out For Delivery Date",
		"First_Pickup_Scheduled_Date",
		"Buyer's Lat/long",
		"Invoice Date",
		"Pickup Code",
		"Eway Bill Nos",
		"Last Updated AT",
		"Rto Risk",
		"RAD Datetimestamp",
		"Pickup Pincode",
		"RTO Reason",
	}
	report = append(report, headers)

	for _, order := range orders {
		row := CreateReportRowFromAdminOrder(order)
		report = append(report, row)
	}

	csvData, err := GenerateCSVData(report)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to generate CSV data")
	}

	return csvData, nil
}

func GenerateCSVData(data [][]string) ([]byte, error) {
	csvData := &bytes.Buffer{}
	writer := csv.NewWriter(csvData)

	for _, row := range data {
		err := writer.Write(row)
		if err != nil {
			return nil, err
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		return nil, err
	}

	return csvData.Bytes(), nil
}

func CreateReportRowFromAdminOrder(order AdminOrder) []string {
	var row []string

	var orderQuantity int
	for _, product := range order.ShiprocketOrder.Products {
		orderQuantity += product.Quantity
	}

	// "Order ID",
	row = append(row, order.ID)
	// 	"Shiprocket Created At",
	row = append(row, order.Shipping.CreatedAt.Local().Format("2006-01-02"))
	// 	"SKU",
	row = append(row, order.Cart.Items[0].Variant.SKU)
	// 	"Product Name",
	row = append(row, order.Cart.Items[0].Product.Name)
	// 	"Product Quantity",
	row = append(row, fmt.Sprint(order.Cart.Items[0].Quantity))
	// 	"order Created At",
	row = append(row, order.CreatedAt.Local().Format("2006-01-02"))
	// 	"Customer Name",
	row = append(row, order.Address.Name)
	// 	"Customer Email",
	row = append(row, order.User.Email)
	// 	"Customer Mobile",
	row = append(row, order.Address.Phone)
	// 	"Address Line 1",
	row = append(row, order.Address.HouseArea+","+order.Address.StreetName)
	// 	"Address City",
	row = append(row, order.Address.City)
	// 	"Address State",
	row = append(row, order.Address.State)
	// 	"Address Pincode",
	row = append(row, order.Address.PinCode)
	// 	"Payment Method",
	var paymentMethod string
	if order.Payment.Method == "COD" {
		paymentMethod = "Cash on Delivery"
	} else {
		paymentMethod = "Prepaid"
	}
	row = append(row, paymentMethod)
	// 	"Payment Status",
	row = append(row, order.PaymentStatus)
	// 	"Product Price",
	row = append(row, fmt.Sprint(order.Cart.Items[0].Variant.Price.SellingPrice))
	// 	"Order Total Amount",
	row = append(row, fmt.Sprint(order.Cart.CartAmount.TotalAmount))
	// 	"Tax",
	row = append(row, fmt.Sprint(tmpl.TaxValue(order.Cart.Items[0].Variant.Price.SellingPrice, order.Cart.Items[0].Product.GST)))
	// 	"Tax %",
	row = append(row, fmt.Sprint(order.Cart.Items[0].Product.GST)+"%")
	// 	"Discount Value",
	row = append(row, fmt.Sprint(order.Cart.CartAmount.ProductPrice.Discount))
	// 	"Product HSN",
	row = append(row, fmt.Sprint(order.Cart.Items[0].Product.Barcode))
	// 	"Weight (KG)",
	row = append(row, fmt.Sprint(order.Cart.Items[0].Variant.Weight))
	// 	"dimensions (CM)",
	row = append(row, fmt.Sprintf("%scm*%scm*%scm", order.Cart.Items[0].Variant.Length, order.Cart.Items[0].Variant.Breadth, order.Cart.Items[0].Variant.Height))
	// 	"Charged Weight",

	if order.ShiprocketOrder.AwbData.Charges.ChargedWeight == "0.00" {
		row = append(row, order.ShiprocketOrder.AwbData.Charges.AppliedWeight)
	} else {
		row = append(row, fmt.Sprint(order.ShiprocketOrder.AwbData.Charges.ChargedWeight))
	}
	// 	"Courier Company",
	row = append(row, order.Tracking.CourierName)
	// 	"AWB Code",
	row = append(row, order.Shipping.AWB)
	// 	"Shipping Bill URL",
	row = append(row, order.Shipping.LabelURL)
	// 	"AWB Assigned Date",
	row = append(row, order.Shipping.CreatedAt.Local().Format("2006-01-02"))
	// 	"Pickup Location ID",
	row = append(row, order.Seller.ID)
	// 	"Pickup Address Name",
	row = append(row, order.Seller.HouseArea+","+order.Seller.StreetName)
	// 	"Pickup scheduled Date",
	row = append(row, order.Shipping.CreatedAt.Local().Format("2006-01-02"))
	// 	"Order Picked Up Date",
	row = append(row, "")
	// 	"Pickup First Attempt Date",
	row = append(row, "")
	// 	"Pickedup Timestamp",
	row = append(row, "")
	// 	"Order Shipped Date",
	row = append(row, order.ShiprocketOrder.Shipments.ShippedDate)
	// 	"EDD Delayed Reason",
	row = append(row, "")
	// 	"Order Delivered Date",
	row = append(row, order.ShiprocketOrder.DeliveredDate)
	// 	"RTO Address",
	row = append(row, order.Seller.HouseArea+","+order.Seller.StreetName)
	// 	"RTO Initiated Date",
	row = append(row, order.ShiprocketOrder.Shipments.RtoInitiatedDate)
	// 	"RTO Delivered Date",
	row = append(row, order.ShiprocketOrder.Shipments.RtoDeliveredDate)
	// 	"COD Payble Amount",
	row = append(row, fmt.Sprint(order.Cart.CartAmount.TotalAmount))
	// 	"UTR No",
	row = append(row, order.ShiprocketOrder.RemittanceUtr)

	// 	"COD Charges",
	var codAmount float64

	if order.ShiprocketOrder.AwbData.Charges.CodCharges > 0 {
		codAmount = float64(order.ShiprocketOrder.AwbData.Charges.CodCharges) / float64(orderQuantity)
	} else {
		codAmount = 0
	}

	if order.ShiprocketOrder.Status == "RTO DELIVERED" {
		row = append(row, fmt.Sprint(0))
	} else {
		row = append(row, fmt.Sprint(codAmount))
	}

	// 	"RTO Charges",
	var rtoAmount float64

	if order.ShiprocketOrder.Status == "RTO DELIVERED" {
		// Shiprocket high spec coding standards
		if order.ShiprocketOrder.AwbData.Charges.ChargedWeightAmountRto == "0.00" {

			rtoInt, err := strconv.ParseFloat(order.ShiprocketOrder.AwbData.Charges.AppliedWeightAmountRto, 64)
			if err != nil {
				fmt.Println("rtoInt err:", err.Error())
				rtoAmount = 0
			} else {
				rtoAmount = rtoInt / float64(orderQuantity)
			}

		} else {

			rtoInt, err := strconv.ParseFloat(order.ShiprocketOrder.AwbData.Charges.ChargedWeightAmountRto, 64)
			if err != nil {
				fmt.Println("rtoInt err:", err.Error())
				rtoAmount = 0
			} else {
				rtoAmount = rtoInt / float64(orderQuantity)
			}
		}
		row = append(row, fmt.Sprint(rtoAmount))
	} else {
		row = append(row, "")
	}

	// 	"Freight Total Amount", Shiprocket high spec coding standards
	var freightAmount float64

	if order.ShiprocketOrder.AwbData.Charges.ChargedWeightAmount == "" {

		freightInt, err := strconv.ParseFloat(order.ShiprocketOrder.AwbData.Charges.AppliedWeightAmount, 64)
		if err != nil {
			fmt.Println("freightInt err:", err.Error())
			freightAmount = 0
		} else {
			freightAmount = freightInt / float64(orderQuantity)
		}

		row = append(row, fmt.Sprint(freightAmount-codAmount))
	} else {

		freightInt, err := strconv.ParseFloat(order.ShiprocketOrder.AwbData.Charges.ChargedWeightAmount, 64)
		if err != nil {
			fmt.Println("freightInt err:", err.Error())
			freightAmount = 0
		} else {
			freightAmount = freightInt / float64(orderQuantity)
		}

		row = append(row, fmt.Sprint(freightAmount-codAmount))
	}

	// 	"Total Shipping Charges",
	if order.ShiprocketOrder.Status == "RTO DELIVERED" {
		row = append(row, fmt.Sprint(rtoAmount+freightAmount-codAmount))
	} else {
		row = append(row, fmt.Sprint(freightAmount))
	}

	// 	"Transaction charges",
	transactionCharge := 0.0

	if order.Payment.Method != "COD" {
		transactionCharge = order.OrderAmount.TotalAmount * 0.02
	}

	row = append(row, fmt.Sprint(transactionCharge))
	// 	"Customer_invoice_id",
	row = append(row, fmt.Sprint(order.Shipping.OrderId))
	// 	"Pickup Exception Reason",
	row = append(row, "")
	// 	"First Out For Delivery Date",
	row = append(row, order.ShiprocketOrder.OutForDeliveryDate)
	// 	"First_Pickup_Scheduled_Date",
	row = append(row, "")
	// 	"Buyer's Lat/long",
	row = append(row, "")
	// 	"Invoice Date",
	row = append(row, order.Shipping.CreatedAt.Local().Format("2006-01-02"))
	// 	"Pickup Code",
	row = append(row, order.ShiprocketOrder.PickupCode)
	// 	"Eway Bill Nos",
	row = append(row, order.ShiprocketOrder.EwayBillNumber)
	// 	"Last Updated AT",
	row = append(row, order.UpdatedAt.Local().Format("2006-01-02"))
	// 	"Rto Risk",
	row = append(row, "")
	// 	"RAD Datetimestamp",
	row = append(row, "")
	// 	"Pickup Pincode",
	row = append(row, order.Seller.PinCode)
	// 	"RTO Reason",
	row = append(row, "")

	return row
}
