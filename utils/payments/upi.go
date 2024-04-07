package payments

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tryamigo/themis"
)

type decentroPaymentLinkRequest struct {
	ReferenceID    string  `json:"reference_id"`
	PayeeAccount   string  `json:"payee_account"`
	Amount         float64 `json:"amount"`
	PurposeMessage string  `json:"purpose_message"`
	GenerateQR     int     `json:"generate_qr"`
	ExpiryTime     int     `json:"expiry_time"`
	GenerateURI    int     `json:"generate_uri"`
}

type decentroPaymentLinkResponse struct {
	DecentroTransactionID string         `json:"decentroTxnId"`
	Status                string         `json:"status"`
	ResponseCode          string         `json:"responseCode"`
	Message               string         `json:"message"`
	Data                  UPIPaymentLink `json:"data"`
	ResponseKey           string         `json:"responseKey"`
}

type UPIPaymentLink struct {
	UPIURI string `json:"upiUri"`
	PSPURI struct {
		GpayURI    string `json:"gpayUri"`
		PhonepeURI string `json:"phonepeUri"`
		PaytmURI   string `json:"paytmUri"`
		CommonURI  string `json:"commonUri"`
	} `json:"pspUri"`
	GeneratedLink     string `json:"generatedLink"`
	TransactionID     string `json:"transactionId"`
	TransactionStatus string `json:"transactionStatus"`
}

type decentroErrorResponse struct {
	DecentroTransactionID string `json:"decentroTxnId"`
	Status                string `json:"status"`
	ResponseCode          string `json:"responseCode"`
	Message               string `json:"message"`
}

type UPIRequestResponse struct {
	ThirdPartyTransactionID string
	UPIDeepLinks            UPIPaymentLink
}

func (err decentroErrorResponse) Error() string {
	return fmt.Sprintf("DecentroTransactionID: %s, Status: %s, ResponseCode: %s, Message: %s", err.DecentroTransactionID, err.Status, err.ResponseCode, err.Message)
}

func makeDecentroAPIRequest(endpoint string, requestMethod string, requestBody []byte) ([]byte, error) {
	url := "https://in.decentro.tech/" + endpoint

	headers := [][]string{
		{"client_id", os.Getenv("DECENTRO_CLIENT_ID")},
		{"client_secret", os.Getenv("DECENTRO_CLIENT_SECRET")},
		{"module_secret", os.Getenv("DECENTRO_PAYMENTS_MODULE_SECRET")},
		{"provider_secret", os.Getenv("DECENTRO_COSMOS_PROVIDER_SECRET")},
	}

	statusCode, responseBody, err := themis.HitAPIEndpoint2(url, requestMethod, requestBody, headers, nil)

	if err != nil {
		return nil, err
	}

	if statusCode >= 400 {
		var err decentroErrorResponse
		unmarshalErr := json.Unmarshal(responseBody, &err)
		if unmarshalErr != nil {
			return nil, fmt.Errorf("could not read error from Decentro: %s", string(responseBody))
		}
		return nil, err
	}

	return responseBody, nil

}

func InitiateUPIPayment(paymentID string, amount float64, message string) (UPIRequestResponse, decentroPaymentLinkResponse, error) {

	var emptyResponse UPIRequestResponse
	var response decentroPaymentLinkResponse

	request := decentroPaymentLinkRequest{
		ReferenceID:    paymentID,
		PayeeAccount:   os.Getenv("Decentro_ACCOUNT_NO"),
		Amount:         amount,
		PurposeMessage: message,
		GenerateQR:     0,
		ExpiryTime:     15,
		GenerateURI:    1,
	}
	requestBody, _ := json.Marshal(request)

	responseBody, err := makeDecentroAPIRequest("v2/payments/upi/link", "POST", requestBody)
	if err != nil {
		return emptyResponse, response, err
	}

	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return emptyResponse, response, fmt.Errorf("could not read success response from Decentro: " + string(responseBody))
	}

	return UPIRequestResponse{
		ThirdPartyTransactionID: response.DecentroTransactionID,
		UPIDeepLinks:            response.Data,
	}, response, nil

}
