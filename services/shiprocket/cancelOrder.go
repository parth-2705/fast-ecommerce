package shiprocket

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tryamigo/themis"
)

type ShipRocketCancelOrder struct {
	Ids []string `json:"ids,omitempty"` //required
}

func CreateShipRocketCancelOrderRequestBody(orderID string) ShipRocketCancelOrder {
	var ShipRocketCancelOrder ShipRocketCancelOrder
	ShipRocketCancelOrder.Ids = append(ShipRocketCancelOrder.Ids, orderID)
	return ShipRocketCancelOrder
}

func CancelShiprocketOrder(shiprocketOrderID string) error {
	token, err := GetShipRocketToken()
	if err != nil {
		return err
	}

	url := "https://apiv2.shiprocket.in/v1/external/orders/cancel"
	headers := [][]string{{"Authorization", "Bearer " + token}}

	shipRocketCancelOrder := CreateShipRocketShipmentInvoiceRequestBody(shiprocketOrderID)
	requestBody, _ := json.Marshal(shipRocketCancelOrder)

	statusCode, responseBody, err := themis.HitAPIEndpoint2(url, http.MethodPost, requestBody, headers, [][]string{})
	if err != nil {
		return err
	}
	if statusCode >= 400 {
		return fmt.Errorf("erorr: %s with status code: %d", string(responseBody), statusCode)
	}

	return nil
}
