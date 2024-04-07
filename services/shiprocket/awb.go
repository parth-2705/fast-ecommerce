package shiprocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/tryamigo/themis"
)

type ShiprocketAWBAssignStatus struct {
	AwbAssignStatus int `json:"awb_assign_status,omitempty"`
}

type ShiprocketAWBResponse struct {
	AwbAssignStatus int `json:"awb_assign_status,omitempty"`
	Response        struct {
		Data struct {
			AwbCode string `json:"awb_code,omitempty"`
		} `json:"data,omitempty"`
	} `json:"response,omitempty"`
}

type ShipRocketShipment struct {
	ShipmentID string `json:"shipment_id,omitempty"` // required
	CourierID  int    `json:"courier_id,omitempty"`
	Status     string `json:"status,omitempty"`
}

func CreateShipRocketOrderShipmentRequestBody(shipmentID string, courierID string) ShipRocketShipment {
	var shipRocketShipment ShipRocketShipment
	shipRocketShipment.ShipmentID = shipmentID
	if courierID != "" {
		courierIDInt, err := strconv.Atoi(courierID)
		if err != nil {
			return shipRocketShipment
		}
		shipRocketShipment.CourierID = courierIDInt
	}
	return shipRocketShipment
}

func CreateReAssignShipRocketOrderShipmentRequestBody(shipmentID string) ShipRocketShipment {
	var shipRocketShipment ShipRocketShipment
	shipRocketShipment.ShipmentID = shipmentID
	shipRocketShipment.Status = "reassign"
	return shipRocketShipment
}

func CreateShipRocketOrderShipmentUtil(shipRocketShipment ShipRocketShipment, headers [][]string) (string, error) {
	requestBody, _ := json.Marshal(shipRocketShipment)

	url := "https://apiv2.shiprocket.in/v1/external/courier/assign/awb"

	statusCode, responseBody, err := themis.HitAPIEndpoint2(url, http.MethodPost, requestBody, headers, [][]string{})
	if err != nil {
		return "", err
	}

	if statusCode >= 400 {
		return "", fmt.Errorf("error for id: %d with status code: %d in assigning awb", shipRocketShipment.CourierID, statusCode)
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

func CreateShipRocketOrderShipment(shipmentID string, headers [][]string, courierID string) (string, error) {
	shipRocketShipment := CreateShipRocketOrderShipmentRequestBody(shipmentID, courierID)
	awbCode, err := CreateShipRocketOrderShipmentUtil(shipRocketShipment, headers)
	if err != nil {
		return awbCode, err
	}
	return awbCode, nil
}

func CreateReAssignShipRocketOrderShipment(shipmentID string, headers [][]string) (string, error) {
	shipRocketShipment := CreateReAssignShipRocketOrderShipmentRequestBody(shipmentID)
	awbCode, err := CreateShipRocketOrderShipmentUtil(shipRocketShipment, headers)
	if err != nil {
		return awbCode, err
	}
	return awbCode, nil
}
