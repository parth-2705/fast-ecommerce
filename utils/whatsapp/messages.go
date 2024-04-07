package whatsapp

import (
	"encoding/json"
	"fmt"
	"hermes/models/Logs"

	"github.com/getsentry/sentry-go"
)

type TemplateParameter struct {
	Type    string        `json:"type"`
	Text    string        `json:"text,omitempty"`
	Value   string        `json:"value,omitempty"`
	Image   MediaObject   `json:"image,omitempty"`
	Payload string        `json:"payload,omitempty"`
	Action  *ActionObject `json:"action,omitempty"`
}

type ActionObject struct {
	ThumbnailProductRetailerId string    `json:"thumbnail_product_retailer_id,omitempty"`
	Sections                   []Section `json:"sections,omitempty"`
}

type Section struct {
	Title        string        `json:"title,omitempty"`
	ProductItems []ProductItem `json:"product_items,omitempty"`
}

type ProductItem struct {
	ProductRetailerID string `json:"product_retailer_id,omitempty"`
}

type MediaObject struct {
	ID   string `json:"id,omitempty"`
	Link string `json:"link,omitempty"`
}

type TemplateComponent struct {
	Type       string              `json:"type"`
	SubType    string              `json:"sub_type,omitempty"`
	Index      int                 `json:"index"`
	Parameters []TemplateParameter `json:"parameters"`
}

func FormatPhoneNumberForWhatsapp(phone_number string) string {
	// if length is 10, add 91
	if len(phone_number) == 10 {
		phone_number = "91" + phone_number
	}

	// if length is 13, remove +
	if len(phone_number) == 13 {
		phone_number = phone_number[1:]
	}

	if len(phone_number) == 12 {
		return phone_number
	}

	return ""
}

func SendMessageWithTemplate(phone_number string, template_id string, tms []TemplateParameter) error {
	// make the request body
	phone_number = FormatPhoneNumberForWhatsapp(phone_number)
	if phone_number == "" {
		return fmt.Errorf("invalid phone number")
	}
	var requestBody = map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                phone_number,
		"type":              "template",
		"template": map[string]interface{}{
			"name":     template_id,
			"language": map[string]string{"code": "en_US"},
			"components": []TemplateComponent{
				{
					Type:       "body",
					Parameters: tms,
				},
			},
		},
	}

	// make the request
	resp, err := WhatsappAPIRequest("/"+WHATSAPP_PHONE_NUMBER_ID+"/messages", "POST", requestBody)
	if err != nil {
		return err
	}

	// check if the status code is 200
	if resp.StatusCode != 200 {
		return fmt.Errorf("whatsapp api returned status code %d", resp.StatusCode)
	}

	return nil

}

func SendMessageWithTemplate2(phoneNumber string, templateId string, components []TemplateComponent, commType commType, extra map[string]string) error {

	formattedPhoneNumber := FormatPhoneNumberForWhatsapp(phoneNumber)
	if phoneNumber == "" {
		return fmt.Errorf("invalid phone number")
	}

	requestBody := RoovoSendMessageRequest{
		PhoneNumber: formattedPhoneNumber,
		TemplateID:  templateId,
		Type:        "template",
		Components:  components,
		CommType:    commType,
		Extra:       extra,
	}

	requestBodyJson, _ := json.Marshal(requestBody)

	// fmt.Printf("requestBodyJson: %v\n", string(requestBodyJson))

	fmt.Println("Sending Request to Whatsapp")

	log, _ := Logs.CreateWhatsappMessageLog(phoneNumber, templateId)

	responseStatus, err := WhatsAppServiceAPIRequest(requestBodyJson)
	errString := ""
	if err != nil {
		fmt.Printf("err: %v\n", err)
		errString = err.Error()
	}

	log.UpdateResponseStatus(responseStatus, errString)

	return err
}

func SendOrderConfirmationMessage(phone_number string, customer_name string, product_name string, order_id string, tracking_url string) error {
	if customer_name == "" {
		customer_name = "Customer"
	}
	var tms = []TemplateParameter{
		{
			Type: "text",
			Text: customer_name,
		},
		{
			Type: "text",
			Text: product_name,
		},
	}

	var components = []TemplateComponent{
		{
			Type:       "body",
			Parameters: tms,
		},
	}

	err := SendMessageWithTemplate2(phone_number, WHATSAPP_ORDER_CONFIRMATION_TEMPLATE_ID, components, POSTTRANSACTIONCOMM, nil)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return err
	}

	return nil
}

func SendOrderConfirmationMessage2(phone_number string, customer_name string, product_name string, order_id string) error {

	defer sentry.Recover()

	magicID, err := GetMagicURLFromWhatsapp(phone_number, "/order/"+order_id+"?trackingOpen=true")
	if err != nil {
		return err
	}

	if customer_name == "" {
		customer_name = "Customer"
	}
	var tms = []TemplateParameter{
		{
			Type: "text",
			Text: customer_name,
		},
		{
			Type: "text",
			Text: product_name,
		},
	}

	var components = []TemplateComponent{
		{
			Type:       "body",
			Parameters: tms,
		},
		{
			Type:    "button",
			SubType: "url",
			Index:   0,
			Parameters: []TemplateParameter{
				{
					Type: "text",
					Text: magicID,
				},
			},
		},
	}

	extra := map[string]string{
		"OrderID": order_id,
	}

	err = SendMessageWithTemplate2(phone_number, WHATSAPP_ORDER_CONFIRMATION_TEMPLATE_ID, components, POSTTRANSACTIONCOMM, extra)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return err
	}

	return nil
}

func SendOrderOutForPickupMessage(phone_number string, customer_name string, product_name string, order_id string) error {

	magicID, err := GetMagicURLFromWhatsapp(phone_number, "/order/"+order_id+"?trackingOpen=true")
	if err != nil {
		return err
	}

	if customer_name == "" {
		customer_name = "Customer"
	}
	var tms = []TemplateParameter{
		{
			Type: "text",
			Text: customer_name,
		},
		{
			Type: "text",
			Text: product_name,
		},
	}

	var components = []TemplateComponent{
		{
			Type:       "body",
			Parameters: tms,
		},
		{
			Type:    "button",
			SubType: "url",
			Index:   0,
			Parameters: []TemplateParameter{
				{
					Type: "text",
					Text: magicID,
				},
			},
		},
	}

	err = SendMessageWithTemplate2(phone_number, WHATSAPP_ORDER_OUT_FOR_PICKUP_TEMPLATE_ID, components, POSTTRANSACTIONCOMM, nil)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return err
	}

	return nil
}

func SendOrderDeliveredMessage(phoneNumber string, customerName string, productName string, productID string) (err error) {

	defer sentry.Recover()

	var components = []TemplateComponent{
		{
			Type: "body",
			Parameters: []TemplateParameter{
				{
					Type: "text",
					Text: customerName,
				},
				{
					Type: "text",
					Text: productName,
				},
			},
		},
		{
			Type:    "button",
			SubType: "url",
			Index:   0,
			Parameters: []TemplateParameter{
				{
					Type: "text",
					Text: productID,
				},
			},
		},
	}

	err = SendMessageWithTemplate2(phoneNumber, WHATSAPP_ORDER_DELIVERED_TEMPLATE_ID, components, POSTTRANSACTIONCOMM, nil)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
	return
}

func SendOrderDeliveredMessage2(phoneNumber string, customerName string, productName string, productID string) (err error) {

	magicID, err := GetMagicURLFromWhatsapp(phoneNumber, "/review/"+productID)
	if err != nil {
		return
	}

	var components = []TemplateComponent{
		{
			Type: "body",
			Parameters: []TemplateParameter{
				{
					Type: "text",
					Text: customerName,
				},
				{
					Type: "text",
					Text: productName,
				},
			},
		},
		{
			Type:    "button",
			SubType: "url",
			Index:   0,
			Parameters: []TemplateParameter{
				{
					Type: "text",
					Text: magicID,
				},
			},
		},
	}

	err = SendMessageWithTemplate2(phoneNumber, WHATSAPP_ORDER_DELIVERED_TEMPLATE_ID, components, POSTTRANSACTIONCOMM, nil)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
	return
}

func SendPaymentLinkMessage(phoneNumber string, payID string, amount int64) (err error) {
	var components = []TemplateComponent{
		{
			Type: "body",
			Parameters: []TemplateParameter{
				{
					Type: "text",
					Text: fmt.Sprint(amount),
				},
			},
		},
		{
			Type:    "button",
			SubType: "url",
			Index:   0,
			Parameters: []TemplateParameter{
				{
					Type: "text",
					Text: payID,
				},
			},
		},
	}
	err = SendMessageWithTemplate2(phoneNumber, WHATSAPP_PAYMENT_LINK_TEMPLATE_ID, components, ORDERCOMM, nil)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
	return
}

func SendCODConfirmationTemplate(ImageLink string, phoneNumber string, customerName string, productName string, quantity int, price int, orderID string) (err error) {
	var components = []TemplateComponent{
		{
			Type: "header",
			Parameters: []TemplateParameter{
				{
					Type: "image",
					Image: MediaObject{
						Link: ImageLink,
					},
				},
			},
		},
		{
			Type: "body",
			Parameters: []TemplateParameter{
				{
					Type: "text",
					Text: customerName,
				},
				{
					Type: "text",
					Text: productName,
				},
				{
					Type: "text",
					Text: fmt.Sprint(quantity),
				},
				{
					Type: "text",
					Text: fmt.Sprint(price),
				},
			},
		},
		{
			Type:    "button",
			SubType: "quick_reply",
			Index:   0,
			Parameters: []TemplateParameter{
				{Type: "payload", Payload: fmt.Sprintf("%d//%s//%s", CODCONFIRM, "confirm", orderID)},
			},
		},
		{
			Type:    "button",
			SubType: "quick_reply",
			Index:   1,
			Parameters: []TemplateParameter{
				{Type: "payload", Payload: fmt.Sprintf("%d//%s//%s", CODCONFIRM, "convert_to_prepaid", orderID)},
			},
		},
		{
			Type:    "button",
			SubType: "quick_reply",
			Index:   2,
			Parameters: []TemplateParameter{
				{Type: "payload", Payload: fmt.Sprintf("%d//%s//%s", CODCONFIRM, "cancel", orderID)},
			},
		},
	}
	err = SendMessageWithTemplate2(phoneNumber, WHATSAPP_COD_CONFIRM, components, ORDERCOMM, nil)
	return
}

func SendOrderOutForDeliveryMessage(phoneNumber string, customerName string, productName string) (err error) {

	defer sentry.Recover()

	var components = []TemplateComponent{
		{
			Type: "body",
			Parameters: []TemplateParameter{
				{
					Type: "text",
					Text: customerName,
				},
				{
					Type: "text",
					Text: productName,
				},
			},
		},
	}

	err = SendMessageWithTemplate2(phoneNumber, WHATSAPP_ORDER_OFD_TEMPLATE_ID, components, POSTTRANSACTIONCOMM, nil)
	return
}

func SendOrderInTransitMessage(phoneNumber string, customerName string, brandName string, daysToDeliver string, order_id string) (err error) {

	// a function that runs in a goroutine to send whatsapp message
	magicID, err := GetMagicURLFromWhatsapp(phoneNumber, "/order/"+order_id+"?trackingOpen=true")
	if err != nil {
		return err
	}

	var components = []TemplateComponent{
		{
			Type: "body",
			Parameters: []TemplateParameter{
				{
					Type: "text",
					Text: customerName,
				},
				{
					Type: "text",
					Text: brandName,
				},
				{
					Type: "text",
					Text: daysToDeliver,
				},
			},
		},
		{
			Type:    "button",
			SubType: "url",
			Index:   0,
			Parameters: []TemplateParameter{
				{
					Type: "text",
					Text: magicID,
				},
			},
		},
	}

	err = SendMessageWithTemplate2(phoneNumber, WHATSAPP_ORDER_IN_TRANSIT, components, POSTTRANSACTIONCOMM, nil)
	return

}

func SendOrderShippedMessage(phoneNumber string, customerName string, productName string, deliveryDate string, order_id string) (err error) {

	defer sentry.Recover()

	magicID, err := GetMagicURLFromWhatsapp(phoneNumber, "/order/"+order_id+"?trackingOpen=true")
	if err != nil {
		return err
	}

	var components = []TemplateComponent{
		{
			Type: "body",
			Parameters: []TemplateParameter{
				{
					Type: "text",
					Text: customerName,
				},
				{
					Type: "text",
					Text: productName,
				},
				{
					Type: "text",
					Text: deliveryDate,
				},
			},
		},
		{
			Type:    "button",
			SubType: "url",
			Index:   0,
			Parameters: []TemplateParameter{
				{
					Type: "text",
					Text: magicID,
				},
			},
		},
	}

	err = SendMessageWithTemplate2(phoneNumber, WHATSAPP_ORDER_SHIPPED_TEMPLATE_ID, components, POSTTRANSACTIONCOMM, nil)
	return
}

func SendOrderPickedUpMessage(phoneNumber string, customerName string, productName string, deliveryDate string, order_id string) (err error) {

	defer sentry.Recover()

	magicID, err := GetMagicURLFromWhatsapp(phoneNumber, "/order/"+order_id+"?trackingOpen=true")
	if err != nil {
		return err
	}

	var components = []TemplateComponent{
		{
			Type: "body",
			Parameters: []TemplateParameter{
				{
					Type: "text",
					Text: customerName,
				},
				{
					Type: "text",
					Text: productName,
				},
				{
					Type: "text",
					Text: deliveryDate,
				},
			},
		},
		{
			Type:    "button",
			SubType: "url",
			Index:   0,
			Parameters: []TemplateParameter{
				{
					Type: "text",
					Text: magicID,
				},
			},
		},
	}

	err = SendMessageWithTemplate2(phoneNumber, WHATSAPP_ORDER_PICKEDUP_TEMPLATE_ID, components, POSTTRANSACTIONCOMM, nil)
	return
}

type ACRMessageBody struct {
	CartID             string `json:"cartID"`
	ImageLink          string `json:"imageLink"`
	CustomerName       string `json:"customerName"`
	ProductName        string `json:"productName"`
	SellingPrice       string `json:"sellingPrice"`
	MRP                string `json:"mrp"`
	DiscountAmount     string `json:"discountAmount"`
	DiscountPercentage string `json:"discountPercentage"`
}

func SendACR20MinsMessage1(componentsFiller ACRMessageBody, phoneNumber string) (err error) {
	var components = []TemplateComponent{
		{
			Type: "header",
			Parameters: []TemplateParameter{
				{
					Type: "image",
					Image: MediaObject{
						Link: componentsFiller.ImageLink,
					},
				},
			},
		},
		{
			Type: "body",
			Parameters: []TemplateParameter{
				{
					Type: "text",
					Text: componentsFiller.CustomerName,
				},
				{
					Type: "text",
					Text: componentsFiller.ProductName,
				},
				{
					Type: "text",
					Text: componentsFiller.SellingPrice,
				},
			},
		},
		{
			Type:    "button",
			SubType: "quick_reply",
			Index:   0,
			Parameters: []TemplateParameter{
				{Type: "payload", Payload: fmt.Sprintf("%d//%s", RECOVERCARTBUT, componentsFiller.CartID)},
			},
		},
	}

	extra := map[string]string{
		"CartID": componentsFiller.CartID,
	}

	err = SendMessageWithTemplate2(phoneNumber, WHATSAPP_ACR_20MINS_TEMPLATE_ID, components, MARKETTINGCOMM, extra)
	return
}

// func SendACR20MinsMessage(imageLink string, customerName string, productName string, productSellingPrice string, cartID string, phoneNumber string, markettCommDisabled bool) (err error) {
// 	var components = []TemplateComponent{
// 		{
// 			Type: "header",
// 			Parameters: []TemplateParameter{
// 				{
// 					Type: "image",
// 					Image: mediaObject{
// 						Link: imageLink,
// 					},
// 				},
// 			},
// 		},
// 		{
// 			Type: "body",
// 			Parameters: []TemplateParameter{
// 				{
// 					Type: "text",
// 					Text: customerName,
// 				},
// 				{
// 					Type: "text",
// 					Text: productName,
// 				},
// 				{
// 					Type: "text",
// 					Text: productSellingPrice,
// 				},
// 			},
// 		},
// 		{
// 			Type:    "button",
// 			SubType: "url",
// 			Index:   0,
// 			Parameters: []TemplateParameter{
// 				{
// 					Type: "text",
// 					Text: cartID,
// 				},
// 			},
// 		},
// 	}

// 	err = SendMessageWithTemplate2(phoneNumber, WHATSAPP_ACR_20MINS_TEMPLATE_ID, components, MARKETTINGCOMM)
// 	return
// }

func SendACR3DaysMessage(componentsFiller ACRMessageBody, phoneNumber string) (err error) {
	var components = []TemplateComponent{
		{
			Type: "header",
			Parameters: []TemplateParameter{
				{
					Type: "image",
					Image: MediaObject{
						Link: componentsFiller.ImageLink,
					},
				},
			},
		},
		{
			Type: "body",
			Parameters: []TemplateParameter{
				{
					Type: "text",
					Text: componentsFiller.CustomerName,
				},
				{
					Type: "text",
					Text: componentsFiller.ProductName,
				},
				{
					Type: "text",
					Text: componentsFiller.DiscountPercentage,
				},
			},
		},
		{
			Type:    "button",
			SubType: "quick_reply",
			Index:   0,
			Parameters: []TemplateParameter{
				{Type: "payload", Payload: fmt.Sprintf("%d//%s", RECOVERCARTBUT, componentsFiller.CartID)},
			},
		},
	}

	extra := map[string]string{
		"CartID": componentsFiller.CartID,
	}

	err = SendMessageWithTemplate2(phoneNumber, WHATSAPP_ACR_3DAYS_TEMPLATE_ID, components, MARKETTINGCOMM, extra)
	return
}

func SendACR24HrsMessage(componentsFiller ACRMessageBody, phoneNumber string) (err error) {
	var components = []TemplateComponent{
		{
			Type: "header",
			Parameters: []TemplateParameter{
				{
					Type: "image",
					Image: MediaObject{
						Link: componentsFiller.ImageLink,
					},
				},
			},
		},
		{
			Type: "body",
			Parameters: []TemplateParameter{
				{
					Type: "text",
					Text: componentsFiller.CustomerName,
				},
				{
					Type: "text",
					Text: componentsFiller.ProductName,
				},
				{
					Type: "text",
					Text: componentsFiller.SellingPrice,
				},
				{
					Type: "text",
					Text: componentsFiller.MRP,
				},
			},
		},
		{
			Type:    "button",
			SubType: "quick_reply",
			Index:   0,
			Parameters: []TemplateParameter{
				{Type: "payload", Payload: fmt.Sprintf("%d//%s", RECOVERCARTBUT, componentsFiller.CartID)},
			},
		},
	}

	extra := map[string]string{
		"CartID": componentsFiller.CartID,
	}

	err = SendMessageWithTemplate2(phoneNumber, WHATSAPP_ACR_24HRS_TEMPLATE_ID, components, MARKETTINGCOMM, extra)
	return
}

func SendACR4HrsMessage(componentsFiller ACRMessageBody, phoneNumber string) (err error) {
	var components = []TemplateComponent{
		{
			Type: "header",
			Parameters: []TemplateParameter{
				{
					Type: "image",
					Image: MediaObject{
						Link: componentsFiller.ImageLink,
					},
				},
			},
		},
		{
			Type: "body",
			Parameters: []TemplateParameter{
				{
					Type: "text",
					Text: componentsFiller.DiscountAmount,
				},
				{
					Type: "text",
					Text: componentsFiller.CustomerName,
				},
				{
					Type: "text",
					Text: componentsFiller.DiscountAmount,
				},
				{
					Type: "text",
					Text: componentsFiller.ProductName,
				},
			},
		},
		{
			Type:    "button",
			SubType: "quick_reply",
			Index:   0,
			Parameters: []TemplateParameter{
				{Type: "payload", Payload: fmt.Sprintf("%d//%s", RECOVERCARTBUT, componentsFiller.CartID)},
			},
		},
	}

	extra := map[string]string{
		"CartID": componentsFiller.CartID,
	}

	err = SendMessageWithTemplate2(phoneNumber, WHATSAPP_ACR_4HRS_TEMPLATE_ID, components, MARKETTINGCOMM, extra)
	return
}

type ReviewMessageBody struct {
	ImageLink    string
	CustomerName string
	ProductName  string
	ProductID    string
	BrandName    string
}

type ReOrderMessageBody struct {
	ImageLink           string
	CustomerName        string
	ProductName         string
	ProductSellingPrice string
	ProductID           string
}

type ProductReviewPayload struct {
	UserID    string `json:"userID"`
	ProductID string `json:"productID"`
	Rating    int    `json:"rating"`
	RatingStr string `json:"ratingStr"`
}

func SendProductReOrderMessage(filler ReOrderMessageBody, userID string, phoneNumber string) (err error) {

	var components = []TemplateComponent{
		{
			Type: "header",
			Parameters: []TemplateParameter{
				{
					Type: "image",
					Image: MediaObject{
						Link: filler.ImageLink,
					},
				},
			},
		},
		{
			Type: "body",
			Parameters: []TemplateParameter{
				{
					Type: "text",
					Text: filler.CustomerName,
				},
				{
					Type: "text",
					Text: filler.ProductName,
				},
				{
					Type: "text",
					Text: filler.ProductSellingPrice,
				},
			},
		},
		{
			Type:    "button",
			SubType: "quick_reply",
			Index:   0,
			Parameters: []TemplateParameter{
				{Type: "payload", Payload: fmt.Sprintf("%d//%s", ReOrderBut, filler.ProductID)},
			},
		},
		{
			Type:    "button",
			SubType: "quick_reply",
			Index:   1,
			Parameters: []TemplateParameter{
				{Type: "payload", Payload: fmt.Sprintf("%d//%s", ReOrderBut, filler.ProductID)},
			},
		},
	}

	err = SendMessageWithTemplate2(phoneNumber, WHATSAPP_REORDER_TEMPLATE_ID, components, POSTTRANSACTIONCOMM, nil)
	return
}

func SendProductReOrderMessage2(components []TemplateComponent, templateID string, phoneNumber string) (err error) {

	err = SendMessageWithTemplate2(phoneNumber, templateID, components, REORDER, nil)
	return
}

func SendProductReviewMessage(filler ReviewMessageBody, userID string, phoneNumber string) (err error) {

	var components = []TemplateComponent{
		{
			Type: "header",
			Parameters: []TemplateParameter{
				{
					Type: "image",
					Image: MediaObject{
						Link: filler.ImageLink,
					},
				},
			},
		},
		{
			Type: "body",
			Parameters: []TemplateParameter{
				{
					Type: "text",
					Text: filler.CustomerName,
				},
				{
					Type: "text",
					Text: filler.ProductName,
				},
				{
					Type: "text",
					Text: filler.BrandName,
				},
			},
		},
		{
			Type:    "button",
			SubType: "quick_reply",
			Index:   0,
			Parameters: []TemplateParameter{
				{Type: "payload", Payload: fmt.Sprintf("%d//%s//%s//%d", ReviewBut, userID, filler.ProductID, 5)},
			},
		},
		{
			Type:    "button",
			SubType: "quick_reply",
			Index:   1,
			Parameters: []TemplateParameter{
				{Type: "payload", Payload: fmt.Sprintf("%d//%s//%s//%d", ReviewBut, userID, filler.ProductID, 3)},
			},
		},
		{
			Type:    "button",
			SubType: "quick_reply",
			Index:   2,
			Parameters: []TemplateParameter{
				{Type: "payload", Payload: fmt.Sprintf("%d//%s//%s//%d", ReviewBut, userID, filler.ProductID, 1)},
			},
		},
	}

	err = SendMessageWithTemplate2(phoneNumber, WHATSAPP_REVIEW_TEMPLATE_ID, components, REVIEW, nil)
	return
}

func SendNDRMessage(customerName string, productName string, paymentMethod string, phoneNumber string, awb string) (err error) {

	var components = []TemplateComponent{
		{
			Type: "body",
			Parameters: []TemplateParameter{
				{
					Type: "text",
					Text: customerName,
				},
				{
					Type: "text",
					Text: productName,
				},
				{
					Type: "text",
					Text: paymentMethod,
				},
			},
		},
		{
			Type:    "button",
			SubType: "quick_reply",
			Index:   0,
			Parameters: []TemplateParameter{
				{Type: "payload", Payload: fmt.Sprintf("%d//%s//%s//%d", NDR_BUTTON, phoneNumber, awb, 0)},
			},
		},
		{
			Type:    "button",
			SubType: "quick_reply",
			Index:   1,
			Parameters: []TemplateParameter{
				{Type: "payload", Payload: fmt.Sprintf("%d//%s//%s//%d", NDR_BUTTON, phoneNumber, awb, 1)},
			},
		},
		{
			Type:    "button",
			SubType: "quick_reply",
			Index:   2,
			Parameters: []TemplateParameter{
				{Type: "payload", Payload: fmt.Sprintf("%d//%s//%s//%d", NDR_BUTTON, phoneNumber, awb, 5)},
			},
		},
	}

	err = SendMessageWithTemplate2(phoneNumber, WHATSAPP_NDR_TEMPLATE_ID, components, SHIPMENT, nil)
	return
}

func SendAmbassdorRecruitmentMessage(phoneNumber string) (err error) {
	var components = []TemplateComponent{
		{
			Type: "header",
			Parameters: []TemplateParameter{
				{
					Type: "image",
					Image: MediaObject{
						Link: "https://storage.googleapis.com/roovo/Roovo-%20Ads%20Library/Ambassador%20Program.png",
					},
				},
			},
		},
		{
			Type:    "button",
			SubType: "quick_reply",
			Index:   0,
			Parameters: []TemplateParameter{
				{Type: "payload", Payload: fmt.Sprintf("%d", AMBASSDORJOIN)},
			},
		},
	}

	err = SendMessageWithTemplate2(phoneNumber, WHATSAPP_JOINAMBASSDOR_TEMPLATE_ID, components, RECRUITMENT, nil)
	if err != nil {
		fmt.Printf("3423423err: %v\n", err)
	}
	return
}

func SendAmbassdorTutorial(phoneNumber string, imageLink string, tutorialTemplate string, tutorialBodyFiller string) (err error) {

	var components = []TemplateComponent{
		{
			Type: "header",
			Parameters: []TemplateParameter{
				{
					Type: "image",
					Image: MediaObject{
						Link: imageLink,
					},
				},
			},
		},
		{
			Type: "body",
			Parameters: []TemplateParameter{
				{
					Type: "text",
					Text: tutorialBodyFiller,
				},
			},
		},
	}

	err = SendMessageWithTemplate2(phoneNumber, tutorialTemplate, components, RECRUITMENT, nil)
	return err
}

func SendAmbassdorDeal(imageLink string, phoneNumber string) (err error) {

	var components = []TemplateComponent{
		{
			Type: "header",
			Parameters: []TemplateParameter{
				{
					Type: "image",
					Image: MediaObject{
						Link: imageLink,
					},
				},
			},
		},
	}

	err = SendMessageWithTemplate2(phoneNumber, WHATSAPP_AMBASSADOR_DEAL, components, DEAL, nil)
	if err != nil {
		fmt.Printf("3423423err: %v\n", err)
	}
	return err
}

func SendInfluencerOnboardingTemplate(influencerName string, phoneNumber string) (err error) {

	var components = []TemplateComponent{
		{
			Type: "header",
			Parameters: []TemplateParameter{
				{
					Type: "image",
					Image: MediaObject{
						Link: "https://storage.googleapis.com/roovo-images/rawImages/influencer.png",
					},
				},
			},
		},
		{
			Type: "body",
			Parameters: []TemplateParameter{
				{
					Type: "text",
					Text: influencerName,
				},
			},
		},
	}

	err = SendMessageWithTemplate2(phoneNumber, WHATSAPP_INFLUENCER_ONBOARDING_TEMPLATE, components, INFLUENCER, nil)
	if err != nil {
		fmt.Printf("error while sending influencer onboarding template: %v\n", err)
	}
	return err
}

func SendInfluencerProfileApprovalTemplate(phoneNumber string) (err error) {
	var components = []TemplateComponent{}
	err = SendMessageWithTemplate2(phoneNumber, WHATSAPP_INFLUENCER_APPROVAL_TEMPLATE, components, INFLUENCER, nil)
	if err != nil {
		fmt.Printf("error while sending influencer profile approval template: %v\n", err)
	}
	return err
}

func SendInfluencerProfiledDisapprovalTemplate(phoneNumber string) (err error) {
	var components = []TemplateComponent{}
	err = SendMessageWithTemplate2(phoneNumber, WHATSAPP_INFLUENCER_DISAPPROVAL_TEMPLATE, components, INFLUENCER, nil)
	if err != nil {
		fmt.Printf("error while sending influencer profile disapproval template: %v\n", err)
	}
	return err
}
