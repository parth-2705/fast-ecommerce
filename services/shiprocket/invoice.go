package shiprocket

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tryamigo/themis"
)

type ShipRocketInvoiceCreatedResponse struct {
	IsInvoiceCreated bool   `json:"is_invoice_created,omitempty"`
	InvoiceURL       string `json:"invoice_url,omitempty"`
	IrnNo            string `json:"irn_no,omitempty"`
}

type ShipRocketInvoice struct {
	Ids []string `json:"ids,omitempty"` //required
}

func CreateShipRocketShipmentInvoiceRequestBody(orderID string) ShipRocketInvoice {
	var ShipRocketInvoice ShipRocketInvoice
	ShipRocketInvoice.Ids = append(ShipRocketInvoice.Ids, orderID)
	return ShipRocketInvoice
}

func CreateShipRocketShipmentInvoice(orderID string, headers [][]string) (string, error) {
	shipRocketInvoice := CreateShipRocketShipmentInvoiceRequestBody(orderID)
	requestBody, _ := json.Marshal(shipRocketInvoice)

	url := "https://apiv2.shiprocket.in/v1/external/orders/print/invoice"

	statusCode, responseBody, err := themis.HitAPIEndpoint2(url, http.MethodPost, requestBody, headers, [][]string{})
	if err != nil {
		return "", err
	}
	if statusCode >= 400 {
		return "", fmt.Errorf("erorr: %s with status code: %d", string(responseBody), statusCode)
	}

	var ShipRocketInvoiceCreatedResponse ShipRocketInvoiceCreatedResponse
	err = json.Unmarshal(responseBody, &ShipRocketInvoiceCreatedResponse)
	if err != nil {
		return "", err
	}
	return ShipRocketInvoiceCreatedResponse.InvoiceURL, nil
}
