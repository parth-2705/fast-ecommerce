package whatsapp

import (
	"encoding/json"
	"fmt"
	"hermes/models/Logs"
)

type InteractiveTemplate struct {
	Type   string                    `json:"type"`
	Header *InteractiveHeader        `json:"header,omitempty"`
	Body   *InteractiveBodyAndFooter `json:"body,omitempty"`
	Footer *InteractiveBodyAndFooter `json:"footer,omitempty"`
	Action *InteractiveAction        `json:"action,omitempty"`
}

type InteractiveHeader struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}

type InteractiveBodyAndFooter struct {
	Text string `json:"text,omitempty"`
}

type InteractiveAction struct {
	Button []InteractiveButton `json:"button,omitempty"`
}

type InteractiveButton struct {
	Type  string       `json:"type,omitempty"`
	Reply *ReplyObject `json:"reply,omitempty"`
}

type ReplyObject struct {
	ID    string `json:"id,omitempty"`
	Title string `json:"title,omitempty"`
}

type RoovoSendInteractiveMessageRequest struct {
	PhoneNumber string              `json:"phone_number"`
	TemplateID  string              `json:"template_id"`
	Components  InteractiveTemplate `json:"components"`
}

func SendInteractiveMessage(phoneNumber string, interactiveMessageID string, components InteractiveTemplate) error {

	formattedPhoneNumber := FormatPhoneNumberForWhatsapp(phoneNumber)
	if phoneNumber == "" {
		return fmt.Errorf("invalid phone number")
	}

	requestBody := RoovoSendInteractiveMessageRequest{
		PhoneNumber: formattedPhoneNumber,
		TemplateID:  interactiveMessageID,
		Components:  components,
	}

	requestBodyJson, _ := json.Marshal(requestBody)

	// fmt.Printf("requestBodyJson: %v\n", string(requestBodyJson))

	fmt.Println("Sending Request to Whatsapp")

	log, _ := Logs.CreateWhatsappMessageLog(phoneNumber, interactiveMessageID)

	responseStatus, err := WhatsAppServiceAPIRequest(requestBodyJson)
	errString := ""
	if err != nil {
		fmt.Printf("err: %v\n", err)
		errString = err.Error()
	}

	log.UpdateResponseStatus(responseStatus, errString)

	return nil
}

// func SendNDRMessage(customerName string, productName string, paymentMethod string, phoneNumber string) (err error) {

// 	body := fmt.Sprintf("Hi %s, we couldn't deliver your Roovo order. \r\nItems included in the shipment:\r\n%s\r\n\r\nPayment Mode:  %s", customerName, productName, paymentMethod)

// 	var components = InteractiveTemplate{
// 		Type: "button",
// 		Body: &InteractiveBodyAndFooter{
// 			Text: body,
// 		},
// 		Footer: &InteractiveBodyAndFooter{
// 			Text: "Please help us by selecting your delivery preference:",
// 		},
// 		Action: &InteractiveAction{
// 			Button: []InteractiveButton{
// 				{
// 					Type: "reply", Reply: &ReplyObject{
// 						ID:    "Reschedule Delivery",
// 						Title: "Reschedule Delivery",
// 					},
// 				},
// 				{
// 					Type: "reply", Reply: &ReplyObject{
// 						ID:    "No Attempt Made",
// 						Title: "No Attempt Made",
// 					},
// 				},
// 				{
// 					Type: "reply", Reply: &ReplyObject{
// 						ID:    "More Options",
// 						Title: "More Options",
// 					},
// 				},
// 			},
// 		},
// 	}

// 	err = SendInteractiveMessage(phoneNumber, WHATSAPP_REVIEW_TEMPLATE_ID, components)
// 	return
// }
