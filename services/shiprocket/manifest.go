package shiprocket

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tryamigo/themis"
)

type ShipRocketManifestCreatedResponse struct {
	Status      int    `json:"status,omitempty"`
	ManifestURL string `json:"manifest_url,omitempty"`
}

type ShipRocketManifest struct {
	ShipmentID []string `json:"shipment_id,omitempty"` // required
}

func CreateShipRocketShipmentManifestRequestBody(shipmentID string) ShipRocketLabel {
	var shipRocketLabel ShipRocketLabel
	shipRocketLabel.ShipmentID = append(shipRocketLabel.ShipmentID, shipmentID)
	return shipRocketLabel
}

func CreateShipRocketShipmentManifest(shipmentID string, headers [][]string) (string, error) {
	shipRocketLabel := CreateShipRocketShipmentLabelRequestBody(shipmentID)
	requestBody, _ := json.Marshal(shipRocketLabel)

	url := "https://apiv2.shiprocket.in/v1/external/manifests/generate"

	statusCode, responseBody, err := themis.HitAPIEndpoint2(url, http.MethodPost, requestBody, headers, [][]string{})
	if err != nil {
		return "", err
	}
	if statusCode >= 400 {
		return "", fmt.Errorf("erorr: %s with status code: %d", string(responseBody), statusCode)
	}

	var ShipRocketManifestCreatedResponse ShipRocketManifestCreatedResponse
	err = json.Unmarshal(responseBody, &ShipRocketManifestCreatedResponse)
	if err != nil {
		return "", err
	}

	return ShipRocketManifestCreatedResponse.ManifestURL, nil
}
