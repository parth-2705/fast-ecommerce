package shiprocket

import (
	"encoding/json"
	"fmt"
	"hermes/services/Sentry"
	"net/http"

	"github.com/tryamigo/themis"
)

type ShiprocketNDRResponse struct {
	Data []ShiprocketNDRData `json:"data"`
	Meta struct {
		Pagination struct {
			CurrentPage int `json:"current_page"`
			TotalPages  int `json:"total_pages"`
			Links       struct {
				Next string `json:"next"`
			} `json:"links"`
		} `json:"pagination"`
	} `json:"meta"`
}

type ShiprocketNDRData struct {
	AwbCode string `json:"awb_code"`
}

type ShiprocketNDRAction struct {
	Status string `json:"status"`
}

func GetAllShiprocketNDRs() ([]ShiprocketNDRData, error) {
	token, err := GetShipRocketToken()
	if err != nil {
		return []ShiprocketNDRData{}, err
	}

	url := "https://apiv2.shiprocket.in/v1/external/ndr/all"
	headers := [][]string{{"Authorization", "Bearer " + token}}
	params := [][]string{{"per_page", "100"}}

	statusCode, responseBody, err := themis.HitAPIEndpoint2(url, http.MethodGet, []byte{}, headers, params)
	if err != nil {
		return []ShiprocketNDRData{}, err
	}

	if statusCode >= 400 {
		return []ShiprocketNDRData{}, fmt.Errorf("erorr: %s with status code: %d", string(responseBody), statusCode)
	}

	var shiprocketNDRResponse ShiprocketNDRResponse
	err = json.Unmarshal(responseBody, &shiprocketNDRResponse)
	if err != nil {
		return []ShiprocketNDRData{}, err
	}

	var ndrs []ShiprocketNDRData

	ndrs = append(ndrs, shiprocketNDRResponse.Data...)

	for shiprocketNDRResponse.Meta.Pagination.CurrentPage < shiprocketNDRResponse.Meta.Pagination.TotalPages {

		statusCode, responseBody, err := themis.HitAPIEndpoint2(shiprocketNDRResponse.Meta.Pagination.Links.Next, http.MethodGet, []byte{}, headers, params)
		if err != nil {
			return []ShiprocketNDRData{}, err
		}

		if statusCode >= 400 {
			return []ShiprocketNDRData{}, fmt.Errorf("erorr: %s with status code: %d", string(responseBody), statusCode)
		}

		var shiprocketNDRResponse ShiprocketNDRResponse
		err = json.Unmarshal(responseBody, &shiprocketNDRResponse)
		if err != nil {
			return []ShiprocketNDRData{}, err
		}

		ndrs = append(ndrs, shiprocketNDRResponse.Data...)
	}

	return ndrs, nil
}

func parseShiprocketNDRResponse(shiprocketNDRAction ShiprocketNDRAction, responseBody string) {

	if len(shiprocketNDRAction.Status) == 0 || shiprocketNDRAction.Status != "Data Updated Sucessfully" {
		Sentry.SentryCaptureException(fmt.Errorf("unexpected shiprocket response got: %s and unmarshalled status: %s", responseBody, shiprocketNDRAction.Status))
	}

}

func ShiprocketReAttempt(awb string, deferred_date string) error {
	token, err := GetShipRocketToken()
	if err != nil {
		return err
	}

	url := "https://apiv2.shiprocket.in/v1/external/ndr/" + awb + "/action"
	headers := [][]string{{"Authorization", "Bearer " + token}}

	reAttempRequestBody := make(map[string]string, 0)
	reAttempRequestBody["action"] = "re-attempt"
	reAttempRequestBody["deferred_date"] = deferred_date

	requestBody, err := json.Marshal(&reAttempRequestBody)
	if err != nil {
		return err
	}

	statusCode, responseBody, err := themis.HitAPIEndpoint2(url, http.MethodPost, requestBody, headers, [][]string{})
	if err != nil {
		return err
	}

	if statusCode >= 400 {
		return fmt.Errorf("erorr: %s with status code: %d", string(responseBody), statusCode)
	}

	var shiprocketNDRAction ShiprocketNDRAction
	err = json.Unmarshal(responseBody, &shiprocketNDRAction)
	if err != nil {
		return err
	}

	parseShiprocketNDRResponse(shiprocketNDRAction, string(responseBody))

	return nil
}

func ShiprocketUpdateAddressOnReAttempt(awb string, address string) error {
	token, err := GetShipRocketToken()
	if err != nil {
		return err
	}

	url := "https://apiv2.shiprocket.in/v1/external/ndr/" + awb + "/action"
	headers := [][]string{{"Authorization", "Bearer " + token}}

	updateAddressAndReAttempt := make(map[string]string, 0)
	updateAddressAndReAttempt["action"] = "re-attempt"

	if len(address) > 28 {
		updateAddressAndReAttempt["address1"] = address[:27]
		updateAddressAndReAttempt["address2"] = address[28:]
	} else {
		updateAddressAndReAttempt["address1"] = address
	}

	Sentry.SentryCaptureMessage(fmt.Sprintf("Re-attemp with address %s %s for AWB:%s", updateAddressAndReAttempt["address1"], updateAddressAndReAttempt["address2"], awb))
	requestBody, err := json.Marshal(&updateAddressAndReAttempt)
	if err != nil {
		return err
	}

	statusCode, responseBody, err := themis.HitAPIEndpoint2(url, http.MethodPost, requestBody, headers, [][]string{})
	if err != nil {
		return err
	}

	if statusCode >= 400 {
		return fmt.Errorf("erorr: %s with status code: %d", string(responseBody), statusCode)
	}

	var shiprocketNDRAction ShiprocketNDRAction
	err = json.Unmarshal(responseBody, &shiprocketNDRAction)
	if err != nil {
		return err
	}

	parseShiprocketNDRResponse(shiprocketNDRAction, string(responseBody))

	return nil
}

func ShiprocketReturnShipment(awb string) error {
	token, err := GetShipRocketToken()
	if err != nil {
		return err
	}

	url := "https://apiv2.shiprocket.in/v1/external/ndr/" + awb + "/action"
	headers := [][]string{{"Authorization", "Bearer " + token}}

	returnShipment := make(map[string]string, 0)
	returnShipment["action"] = "return"

	requestBody, err := json.Marshal(&returnShipment)
	if err != nil {
		return err
	}

	statusCode, responseBody, err := themis.HitAPIEndpoint2(url, http.MethodPost, requestBody, headers, [][]string{})
	if err != nil {
		return err
	}

	if statusCode >= 400 {
		return fmt.Errorf("erorr: %s with status code: %d", string(responseBody), statusCode)
	}

	var shiprocketNDRAction ShiprocketNDRAction
	err = json.Unmarshal(responseBody, &shiprocketNDRAction)
	if err != nil {
		return err
	}

	parseShiprocketNDRResponse(shiprocketNDRAction, string(responseBody))

	return nil
}

func ShiprocketFakeAttempt(awb string, proof_image string) error {
	token, err := GetShipRocketToken()
	if err != nil {
		return err
	}

	url := "https://apiv2.shiprocket.in/v1/external/ndr/" + awb + "/action"
	headers := [][]string{{"Authorization", "Bearer " + token}}

	fakeAttemptRequestBody := make(map[string]string, 0)
	fakeAttemptRequestBody["action"] = "fake-attempt"
	fakeAttemptRequestBody["proof_image"] = proof_image

	requestBody, err := json.Marshal(&fakeAttemptRequestBody)
	if err != nil {
		return err
	}

	statusCode, responseBody, err := themis.HitAPIEndpoint2(url, http.MethodPost, requestBody, headers, [][]string{})
	if err != nil {
		return err
	}

	if statusCode >= 400 {
		return fmt.Errorf("erorr: %s with status code: %d", string(responseBody), statusCode)
	}

	var shiprocketNDRAction ShiprocketNDRAction
	err = json.Unmarshal(responseBody, &shiprocketNDRAction)
	if err != nil {
		return err
	}

	parseShiprocketNDRResponse(shiprocketNDRAction, string(responseBody))

	return nil
}
