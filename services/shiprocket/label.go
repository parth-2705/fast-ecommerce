package shiprocket

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tryamigo/themis"
)

type ShipRocketLabelCreatedResponse struct {
	LabelCreated int    `json:"label_created,omitempty"`
	LabelURL     string `json:"label_url,omitempty"`
	Response     string `json:"response,omitempty"`
}

type ShipRocketLabel struct {
	ShipmentID []string `json:"shipment_id,omitempty"` // required
}

func CreateShipRocketShipmentLabelRequestBody(shipmentID string) ShipRocketLabel {
	var shipRocketLabel ShipRocketLabel
	shipRocketLabel.ShipmentID = append(shipRocketLabel.ShipmentID, shipmentID)
	return shipRocketLabel
}

func CreateShipRocketShipmentLabel(shipmentID string, headers [][]string) (string, error) {
	shipRocketLabel := CreateShipRocketShipmentLabelRequestBody(shipmentID)
	requestBody, _ := json.Marshal(shipRocketLabel)

	url := "https://apiv2.shiprocket.in/v1/external/courier/generate/label"

	statusCode, responseBody, err := themis.HitAPIEndpoint2(url, http.MethodPost, requestBody, headers, [][]string{})
	if err != nil {
		return "", err
	}
	if statusCode >= 400 {
		return "", fmt.Errorf("erorr: %s with status code: %d", string(responseBody), statusCode)
	}

	var ShipRocketLabelCreatedResponse ShipRocketLabelCreatedResponse
	err = json.Unmarshal(responseBody, &ShipRocketLabelCreatedResponse)
	if err != nil {
		return "", err
	}

	return ShipRocketLabelCreatedResponse.LabelURL, nil
}
