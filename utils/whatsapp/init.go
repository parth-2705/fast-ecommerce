package whatsapp

import (
	"fmt"
	"hermes/utils/data"
	"net/http"
	"os"

	"github.com/tryamigo/themis"
)

func Init() {
	WHATSAPP_BASE_URL = "https://graph.facebook.com/" + os.Getenv("WHATSAPP_API_VERSION")
	WHATSAPP_ACCESS_TOKEN = os.Getenv("WHATSAPP_ACCESS_TOKEN")
	WHATSAPP_PHONE_NUMBER_ID = os.Getenv("WHATSAPP_PHONE_NUMBER_ID")

	WHATSAPP_ORDER_CONFIRMATION_TEMPLATE_ID = os.Getenv("WHATSAPP_ORDER_CONFIRMATION_TEMPLATE_ID")
	WHATSAPP_ORDER_DELIVERED_TEMPLATE_ID = os.Getenv("WHATSAPP_ORDER_DELIVERED_TEMPLATE_ID")
	WHATSAPP_ORDER_OFD_TEMPLATE_ID = os.Getenv("WHATSAPP_ORDER_OFD_TEMPLATE_ID")
	WHATSAPP_ORDER_SHIPPED_TEMPLATE_ID = os.Getenv("WHATSAPP_ORDER_SHIPPED_TEMPLATE_ID")
	WHATSAPP_PAYMENT_LINK_TEMPLATE_ID = os.Getenv("WHATSAPP_PAYMENT_LINK_TEMPLATE_ID")
	WHATSAPP_COD_CONFIRM = os.Getenv("WHATSAPP_COD_CONFIRM")
	WHATSAPP_ACR_20MINS_TEMPLATE_ID = os.Getenv("WHATSAPP_ACR_20MINS_TEMPLATE_ID")
	WHATSAPP_ACR_4HRS_TEMPLATE_ID = os.Getenv("WHATSAPP_ACR_4HRS_TEMPLATE_ID")
	WHATSAPP_ACR_24HRS_TEMPLATE_ID = os.Getenv("WHATSAPP_ACR_24HRS_TEMPLATE_ID")
	WHATSAPP_ACR_3DAYS_TEMPLATE_ID = os.Getenv("WHATSAPP_ACR_3DAYS_TEMPLATE_ID")
	WHATSAPP_REVIEW_TEMPLATE_ID = os.Getenv("WHATSAPP_REVIEW_TEMPLATE_ID")
	WHATSAPP_NDR_TEMPLATE_ID = os.Getenv("WHATSAPP_NDR_TEMPLATE_ID")
	WHATSAPP_REORDER_TEMPLATE_ID = os.Getenv("WHATSAPP_REORDER_TEMPLATE_ID")
	WHATSAPP_ORDER_IN_TRANSIT = os.Getenv("WHATSAPP_ORDER_IN_TRANSIT")
	WHATSAPP_ORDER_OUT_FOR_PICKUP_TEMPLATE_ID = os.Getenv("WHATSAPP_ORDER_OUT_FOR_PICKUP_TEMPLATE_ID")
	WHATSAPP_ORDER_PICKEDUP_TEMPLATE_ID = os.Getenv("WHATSAPP_ORDER_PICKEDUP_TEMPLATE_ID")
	WHATSAPP_JOINAMBASSDOR_TEMPLATE_ID = os.Getenv("WHATSAPP_JOINAMBASSDOR_TEMPLATE_ID")
	WHATSAPP_AMBASSADOR_DEAL = os.Getenv("WHATSAPP_AMBASSADOR_DEAL")

	WHATSAPP_SERVICE_BASE_URL = os.Getenv("WHATSAPP_SERVICE_BASE_URL")
	WHATSAPP_SERVICE_API_KEY = os.Getenv("WHATSAPP_SERVICE_API_KEY")
	WHATSAPP_AMBASSOR_ONBOARDING_TEMPLATE = os.Getenv("WHATSAPP_AMBASSOR_ONBOARDING_TEMPLATE")
	WHATSAPP_INFLUENCER_ONBOARDING_TEMPLATE = os.Getenv("WHATSAPP_INFLUENCER_ONBOARDING_TEMPLATE")
	WHATSAPP_INFLUENCER_APPROVAL_TEMPLATE = os.Getenv("WHATSAPP_INFLUENCER_APPROVAL_TEMPLATE")
	WHATSAPP_INFLUENCER_DISAPPROVAL_TEMPLATE = os.Getenv("WHATSAPP_INFLUENCER_DISAPPROVAL_TEMPLATE")
}

func WhatsappAPIRequest(url string, method string, rawData interface{}) (*http.Response, error) {

	request, err := http.NewRequest(method, WHATSAPP_BASE_URL+url, data.InterfaceToIoReader(rawData))

	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+WHATSAPP_ACCESS_TOKEN)

	// send the request
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type RoovoSendMessageRequest struct {
	PhoneNumber string              `json:"phone_number"`
	TemplateID  string              `json:"template_id"`
	Components  []TemplateComponent `json:"components"`
	Text        string              `json:"text"`
	Type        string              `json:"type"`
	CommType    commType            `json:"commType"`
	Extra       map[string]string   `json:"extra"`
}

func WhatsAppServiceAPIRequest(requestBody []byte) (int, error) {

	endpoint := fmt.Sprintf("%s/%s", WHATSAPP_SERVICE_BASE_URL, "send-message")

	headers := [][]string{
		{"Authorization", fmt.Sprintf("Bearer %s", WHATSAPP_SERVICE_API_KEY)},
	}

	statusCode, responseBody, err := themis.HitAPIEndpoint2(endpoint, "POST", requestBody, headers, nil)

	if err != nil {
		return 0, err
	}

	if statusCode >= 400 {
		return statusCode, fmt.Errorf("error Received from Whatsapp Service: %s", string(responseBody))
	}

	return statusCode, nil

}
