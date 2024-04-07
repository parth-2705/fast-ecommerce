package controllers

import (
	"encoding/json"
	"fmt"
	"hermes/models"
	"hermes/services/Sentry"
	"hermes/utils/payments"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Payment Statu Endpoint that returns the status of a given Payment
func GetPaymentStatus(c *gin.Context) {

	paymentID := c.Param("paymentID")

	payment, err := models.GetPaymentById(paymentID)
	if err != nil {
		c.AbortWithError(404, err)
	}

	c.JSON(200, gin.H{
		"paymentStatus": payment.Status,
	})

}

// payment status endpoint which is as webhook endpoint
func PaymentStatus(c *gin.Context) {
	type PaymentStatusQuery struct {
		PaymentID string `json:"paymentId" binding:"required"`
	}

	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Failed to set websocket upgrade: %+v", err)
		return
	}
	_, msg, err := conn.ReadMessage()
	if err != nil {
		fmt.Println("Failed to read message: %+v", err)
		return
	}
	fmt.Println("Received message: ", string(msg))
	var query PaymentStatusQuery
	err = json.Unmarshal(msg, &query)
	if err != nil {
		fmt.Println("Failed to unmarshal query: %+v", err)
		return
	}
	// loop that queries the payment status and sends it to the client, waits for 1 second and repeats
	for {

		payment, err := models.GetPaymentById(query.PaymentID)
		if err != nil {
			fmt.Println("Failed to get payment: %+v", err)
			return
		}
		fmt.Printf("payment.Status: %v\n", payment.Status)
		// write the payment status to the client
		err = conn.WriteMessage(websocket.TextMessage, []byte(payment.Status))
		if err != nil {
			fmt.Println("Failed to write message: %+v", err)
			return
		}
		// if the payment is completed, close the connection
		if payment.Status == "Succeeded" || payment.Status == "Failed" || payment.Status == "Canceled" || payment.Status == "Processing" {
			conn.Close()
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func PaymentRedirect(c *gin.Context) {
	orderID := c.Query("orderId")
	if orderID == "" {
		c.AbortWithError(400, fmt.Errorf("invalid Order ID"))
		return
	}

	paymentID := c.Query("paymentId")
	if paymentID == "" {
		c.AbortWithError(400, fmt.Errorf("invalid Payment ID"))
		return
	}

	var ws_protocol string
	// check if https:// in os.Getenv("BASE_URL")
	if strings.Contains(os.Getenv("BASE_URL"), "https://") {
		ws_protocol = "wss://"
	} else {
		ws_protocol = "ws://"
	}

	fmt.Printf("ws_protocol: %v\n", ws_protocol)

	c.HTML(http.StatusOK, "payment-redirect.html", gin.H{
		"StripePK":  os.Getenv("STRIPE_PUBLISHABLE_KEY"),
		"OrderID":   orderID,
		"PaymentID": paymentID,
		"StatusUrl": ws_protocol + c.Request.Host + "/payments/status",
	})

}

func Webhook(c *gin.Context) {
	req := c.Request
	w := c.Writer
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	event := stripe.Event{}

	if err := json.Unmarshal(payload, &event); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Webhook error while parsing basic request. %v\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	endpointSecret := os.Getenv("STRIPE_WEBHOOK_ENDPOINT_SECRET")
	signatureHeader := req.Header.Get("Stripe-Signature")
	event, err = webhook.ConstructEvent(payload, signatureHeader, endpointSecret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Webhook signature verification failed. %v\n", err)
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		return
	}
	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case "payment_intent.succeeded":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Successful payment for %d.", paymentIntent.Amount)
		// Then define and call a func to handle the successful payment intent.
		payments.HandlePaymentIntentSucceeded(paymentIntent.ID)
	case "charge.succeeded":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Successful payment for %d.", paymentIntent.Amount)
		// Then define and call a func to handle the successful payment intent.
		payments.HandlePaymentIntentSucceeded(paymentIntent.ID)
	case "charge.failed":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Failed payment for %d.", paymentIntent.Amount)
		// Then define and call a func to handle the successful payment intent.
		payments.HandlePaymentIntentFailed(paymentIntent.ID)
	case "charge.processing":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Processing payment for %d.", paymentIntent.Amount)
		// Then define and call a func to handle the successful payment intent.
		payments.HandlePaymentIntentProcessing(paymentIntent.ID)
	case "payment_method.processing":
		var paymentMethod stripe.PaymentMethod
		err := json.Unmarshal(event.Data.Raw, &paymentMethod)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("PaymentMethod %s is processing.", paymentMethod.ID)
		// Then define and call a func to handle the successful payment intent.
		payments.HandlePaymentIntentProcessing(paymentMethod.ID)
	case "payment_intent.payment_failed":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Payment failed for %d.", paymentIntent.Amount)
		// Then define and call a func to handle the successful payment intent.
		payments.HandlePaymentIntentFailed(paymentIntent.ID)
	default:
		fmt.Fprintf(os.Stderr, "Unhandled event type: %s\n", event.Type)
	}

	w.WriteHeader(http.StatusOK)
}

func SaveCardGetHandler(c *gin.Context) {
	user, err := getUserObjectFromSession(c)
	if err != nil {
		fmt.Println("err getting user from session: ", err)
		fmt.Println(err)
		c.AbortWithError(401, err)
		return
	}

	userProfile, err := user.GetProfile()
	if err != nil {
		fmt.Println("err getting user profile: ", err)
		c.AbortWithError(500, err)
		return
	}
	if userProfile.StripeCustomerID == "" {
		stripeCustomerId, err := payments.CreateStripeCustomer(user.Phone)
		if err != nil {
			fmt.Printf("err stripe customer: %v\n", err)
			c.AbortWithError(500, err)
			return
		}
		userProfile.StripeCustomerID = stripeCustomerId
		err = userProfile.Update()
		if err != nil {
			fmt.Printf("err update user profile: %v\n", err)
			c.AbortWithError(500, err)
			return
		}
	}

	intent, err := payments.CreateSetupIntent(userProfile.StripeCustomerID)
	if err != nil {
		fmt.Printf("err create setup intent: %v\n", err)
		c.AbortWithError(500, err)
		return
	}

	c.HTML(http.StatusOK, "save-card.html", gin.H{
		"ClientSecret": intent.ClientSecret,
		"CustomerID":   userProfile.StripeCustomerID,
		"StripePK":     os.Getenv("STRIPE_PUBLISHABLE_KEY"),
	})
}

func SaveCardPostHandler(c *gin.Context) {
	user, err := getUserObjectFromSession(c)
	if err != nil {
		fmt.Println("err getting user from session: ", err)
		fmt.Println(err)
		c.AbortWithError(401, err)
		return
	}

	var form payments.SaveCardForm
	err = c.ShouldBind(&form)
	if err != nil {
		fmt.Println("err binding form: ", err)
		c.AbortWithError(400, err)
		return
	}

	userProfile, err := user.GetProfile()
	if err != nil {
		fmt.Println("err getting user profile: ", err)
		c.AbortWithError(500, err)
		return
	}

	_, err = payments.CreateCardPaymentMethod(userProfile.StripeCustomerID, form)
	if err != nil {
		fmt.Printf("err create card payment method: %v\n", err)
		c.AbortWithError(500, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
	})

}

var DecentroWebhook401Response gin.H = gin.H{
	"response_code": "CB_E00013",
}

var decentroWebhook400Response gin.H = gin.H{
	"response_code": "CB_E00009",
}

var decentroWebhook200Response gin.H = gin.H{
	"response_code": "CB_S00000",
}

type DecentroTransactionStatusPayload struct {
	Attempt               int     `json:"attempt"`
	TimeStamp             string  `json:"timestamp"`
	CallbackTxnID         string  `json:"callbackTxnId"`
	OriginalCallbackTxnID string  `json:"originalCallbackTxnId"`
	TransactionStatus     string  `json:"transactionStatus"`
	ReferenceID           string  `json:"referenceId"`
	DecentroTxnID         string  `json:"decentroTxnId"`
	TransactionMessage    string  `json:"transactionMessage"`
	TransferType          string  `json:"transferType"`
	BankReferenceNumber   string  `json:"bankReferenceNumber"`
	BeneficiaryName       string  `json:"beneficiaryName"`
	TransactionAmount     float64 `json:"transactionAmount"`
	ProviderMessage       string  `json:"providerMessage"`
	NPCITransactionID     string  `json:"npci_txn_id"`
}

func DecentroWebhook(c *gin.Context) {

	requestBody, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		// Handle error
		models.CreateDecentroWebhookHitLog([]byte(`["Could not read requestBody"]`))
		c.JSON(400, decentroWebhook400Response)
		return
	}

	log, _ := models.CreateDecentroWebhookHitLog(requestBody)

	// Read Payload
	var payload DecentroTransactionStatusPayload
	err = json.Unmarshal(requestBody, &payload)
	if err != nil {
		fmt.Printf("bind err: %v\n", err)
		c.JSON(400, decentroWebhook400Response)
		return
	}

	// Process Payload
	paymentObj, err := models.GetPaymentById(payload.ReferenceID)
	if err != nil {
		fmt.Printf("paymentobj err1: %v\n", err)

		// paymentObj, err = models.GetPaymentObjectByThirdPartyPaymentObjectId(payload.DecentroTxnID)
		// if err != nil {
		// 	fmt.Printf("paymentobj err2: %v\n", err)

		// 	// Could not find the Payment Obj to Update. Send Error to Sentry.
		// 	Sentry.SendErrorToSentry(c, fmt.Errorf("could not Find Payment Obj for this Decentro Callback. Log ID: %s", log.ID), map[string]string{"logID": log.ID, "paymentObjID": payload.ReferenceID, "paymentStatus": payload.TransactionStatus})
		// 	c.JSON(400, decentroWebhook400Response)
		// 	return
		// }

		// Could not find the Payment Obj to Update. Send Error to Sentry.
		Sentry.SendErrorToSentry(c, fmt.Errorf("could not Find Payment Obj for this Decentro Callback. Log ID: %s", log.ID), map[string]string{"logID": log.ID, "paymentObjID": payload.ReferenceID, "paymentStatus": payload.TransactionStatus})
		c.JSON(400, decentroWebhook400Response)
		return
	}

	// If err in processing return 400
	switch payload.TransactionStatus {
	case "SUCCESS":
		err = paymentObj.MarkPaymentAsSuccess()
	case "success":
		err = paymentObj.MarkPaymentAsSuccess()
	case "FAILURE":
		err = paymentObj.MarkPaymentAsFailed()
	case "EXPIRED":
		err = paymentObj.MarkPaymentasExpired()
	default:
		// Unsupported Transaction Status Type
		err = fmt.Errorf("unsupported Decentro Transaction status in Callback: %s, LogID : %s", payload.TransactionStatus, log.ID)
	}

	if err != nil {
		Sentry.SendErrorToSentry(c, err, map[string]string{"logID": log.ID, "paymentObjID": payload.ReferenceID, "paymentStatus": payload.TransactionStatus})
		c.JSON(400, decentroWebhook400Response)
		return
	}

	// If Success Return 200
	c.JSON(200, decentroWebhook200Response)
}

func RedirectToUPIDeeplinks(c *gin.Context) {
	ID := c.Param("payID")
	upiPay, err := models.GetUPIPayObject(ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
	}
	c.Redirect(http.StatusFound, upiPay.Link)
}
