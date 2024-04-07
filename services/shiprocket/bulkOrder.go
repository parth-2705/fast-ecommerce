package shiprocket

import (
	"encoding/json"
	"fmt"
	"hermes/models"
	"net/http"
	"strings"
	"time"

	"github.com/tryamigo/themis"
)

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func CreateBulkShipRocketOrderRequestBody(orders []models.Order, seller models.Seller, discount int64) (ShipRocketOrder, error) {

	var shipRocketOrder ShipRocketOrder

	now := time.Now().UTC()

	shipRocketOrder.OrderID = orders[0].ID
	shipRocketOrder.OrderDate = fmt.Sprint(now.Format(YYYYMMDD))
	shipRocketOrder.PickupLocation = seller.ID

	shipRocketOrder.BillingCustomerName = orders[0].Address.Name
	shipRocketOrder.BillingLastName = orders[0].Address.Name
	shipRocketOrder.BillingAddress = orders[0].Address.HouseArea
	shipRocketOrder.BillingCity = orders[0].Address.City
	shipRocketOrder.BillingPincode = orders[0].Address.PinCode
	shipRocketOrder.BillingState = orders[0].Address.State
	shipRocketOrder.BillingCountry = "India"
	shipRocketOrder.BillingPhone = strings.Replace(orders[0].Address.Phone, "+91", "", -1)

	shipRocketOrder.ShippingIsBilling = true

	var SKUsAdded []string
	SKUQuantity := make(map[string]int, 0)

	// perOrderDiscount := (discount / int64(len(orders))) ///////// Handling for this removed

	var paymentMethod string
	if orders[0].Payment.Method == "COD" {
		paymentMethod = "COD"
	} else {
		paymentMethod = "Prepaid"
	}

	shipRocketOrder.PaymentMethod = paymentMethod

	for _, order := range orders {
		if len(order.Variant.SKU) == 0 {
			return shipRocketOrder, fmt.Errorf("SKU empty")
		}

		SKUQuantity[order.Variant.SKU] += 1
	}
	subTotal := 0

	shipRocketOrder.Length = "0"
	shipRocketOrder.Height = "0"
	shipRocketOrder.Breadth = "0"
	shipRocketOrder.Weight = "0"

	for _, order := range orders {
		if Contains(SKUsAdded, order.Variant.SKU) {
			continue
		}

		singleOrder, err := ProcessSingleOrderForShipment(order)
		if err != nil {
			return shipRocketOrder, err
		}

		shipRocketOrder.OrderItems = append(shipRocketOrder.OrderItems, singleOrder.OrderItems...)
		subTotal += int(singleOrder.Cost)

		// Adding Weights
		shipRocketOrder.Weight, err = addStringsAsFloats(shipRocketOrder.Weight, singleOrder.Weight)
		if err != nil {
			return shipRocketOrder, err
		}

		//Adding Heights
		shipRocketOrder.Height, err = addStringsAsFloats(shipRocketOrder.Height, singleOrder.Height)
		if err != nil {
			return shipRocketOrder, err
		}

		// Taking maximum Length
		// shipRocketOrder.Length, err = addStringsAsFloats(shipRocketOrder.Length, singleOrder.Length)
		shipRocketOrder.Length, err = maxStringAsFloats(shipRocketOrder.Length, singleOrder.Length)
		if err != nil {
			return shipRocketOrder, err
		}

		// Taking maximum Breadth
		// shipRocketOrder.Breadth, err = addStringsAsFloats(shipRocketOrder.Breadth, singleOrder.Breadth)
		shipRocketOrder.Breadth, err = maxStringAsFloats(shipRocketOrder.Breadth, singleOrder.Breadth)
		if err != nil {
			return shipRocketOrder, err
		}

	}
	shipRocketOrder.SubTotal = fmt.Sprint(subTotal - int(discount))

	// for _, order := range orders {
	// 	if Contains(SKUsAdded, order.Variant.SKU) {
	// 		continue
	// 	}

	// 	var shipRocketOrderItem ShipRocketOrderItem
	// 	shipRocketOrderItem.Name = order.Product.Name

	// 	shipRocketOrderItem.Sku = order.Variant.SKU

	// 	shipRocketOrderItem.Units = fmt.Sprint(SKUQuantity[order.Variant.SKU])
	// 	SKUsAdded = append(SKUsAdded, order.Variant.SKU)

	// 	shipRocketOrderItem.SellingPrice = fmt.Sprint(order.Variant.Price.SellingPrice)
	// 	shipRocketOrderItem.Discount = fmt.Sprint(perOrderDiscount)
	// 	shipRocketOrderItem.Tax = getTaxFromBarcode(order.Variant.Barcode)

	// 	shipRocketOrder.OrderItems = append(shipRocketOrder.OrderItems, shipRocketOrderItem)
	// }

	// subTotal := 0
	// for _, order := range orders {
	// 	subTotal += int(order.Product.Price.SellingPrice)
	// }
	// shipRocketOrder.SubTotal = fmt.Sprint(subTotal - int(discount))

	// netLength := 0.0
	// for _, order := range orders {
	// 	if length, err := strconv.ParseFloat(order.Variant.Length, 64); err == nil {
	// 		netLength += length
	// 	} else {
	// 		return shipRocketOrder, err
	// 	}
	// }
	// shipRocketOrder.Length = fmt.Sprint(netLength)

	// netBreadth := 0.0
	// for _, order := range orders {
	// 	if breadth, err := strconv.ParseFloat(order.Variant.Breadth, 64); err == nil {
	// 		netBreadth += breadth
	// 	} else {
	// 		return shipRocketOrder, err
	// 	}
	// }
	// shipRocketOrder.Breadth = fmt.Sprint(netBreadth)

	// netHeights := 0.0
	// for _, order := range orders {
	// 	if height, err := strconv.ParseFloat(order.Variant.Height, 64); err == nil {
	// 		netHeights += height
	// 	} else {
	// 		return shipRocketOrder, err
	// 	}
	// }
	// shipRocketOrder.Height = fmt.Sprint(netHeights)

	// netWeight := 0.0
	// for _, order := range orders {
	// 	if weight, err := strconv.ParseFloat(order.Variant.Weight, 64); err == nil {
	// 		netWeight += weight
	// 	} else {
	// 		return shipRocketOrder, err
	// 	}
	// }
	// shipRocketOrder.Weight = fmt.Sprint(netWeight)

	return shipRocketOrder, nil
}

func CreateShipRocketBulkOrder(orders []models.Order, seller models.Seller, discount int64) error {

	if len(orders) < 2 {
		return fmt.Errorf("bulk Shipment cannot be created")
	}

	paymentMethod := orders[0].Payment.Method
	for _, order := range orders {
		if len(order.Payment.Method) == 0 {
			return fmt.Errorf("order Payment Method is not defined")
		}

		if order.Payment.Method != paymentMethod {
			return fmt.Errorf("order Payment Method is not consistent")
		}
	}

	shiprocketOrder, err := CreateBulkShipRocketOrderRequestBody(orders, seller, discount)
	if err != nil {
		return err
	}

	requestBody, err := json.Marshal(shiprocketOrder)
	if err != nil {
		return err
	}

	token, err := GetShipRocketToken()
	if err != nil {
		return err
	}

	url := "https://apiv2.shiprocket.in/v1/external/orders/create/adhoc"
	headers := [][]string{{"Authorization", "Bearer " + token}}

	statusCode, responseBody, err := themis.HitAPIEndpoint2(url, http.MethodPost, requestBody, headers, [][]string{})
	if err != nil {
		return err
	}
	if statusCode >= 400 {
		return fmt.Errorf("erorr: %s with status code: %d", string(responseBody), statusCode)
	}

	var shipRocketOrderCreatedResponse ShipRocketOrderCreatedResponse
	err = json.Unmarshal(responseBody, &shipRocketOrderCreatedResponse)
	if err != nil {
		return err
	}

	awbCode, err := CreateShipRocketOrderShipment(fmt.Sprint(shipRocketOrderCreatedResponse.ShipmentID), headers, "")
	if err != nil {
		return err
	}

	labelURL, err := CreateShipRocketShipmentLabel(fmt.Sprint(shipRocketOrderCreatedResponse.ShipmentID), headers)
	if err != nil {
		return err
	}

	invoiceURL, err := CreateShipRocketShipmentInvoice(fmt.Sprint(shipRocketOrderCreatedResponse.OrderID), headers)
	if err != nil {
		return err
	}

	for _, order := range orders {
		_, err := CreateShipmentInDatabase(shipRocketOrderCreatedResponse, labelURL, invoiceURL, awbCode, order, shiprocketOrder.OrderID, models.Parent)
		if err != nil {
			return err
		}
	}

	err = CreateShipRocketPickUp(shipRocketOrderCreatedResponse.ShipmentID, "")
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return err
	}

	return nil
}
