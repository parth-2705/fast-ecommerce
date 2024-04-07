package payments

import (
	"fmt"
	"hermes/models"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/paymentintent"
	"github.com/stripe/stripe-go/v74/setupintent"
)

func CreateIntent(payment models.Payment, customerId string) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Customer:         stripe.String(customerId),
		SetupFutureUsage: stripe.String("off_session"),
		Amount:           stripe.Int64(payment.Amount * 100),
		Currency:         stripe.String(string(stripe.CurrencyINR)),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
	}
	result, err := paymentintent.New(params)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func CreatePaymentIntentWithPaymentMethodId(payment models.Payment, customerId string, paymentMethodId string) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Customer:         stripe.String(customerId),
		SetupFutureUsage: stripe.String("off_session"),
		Amount:           stripe.Int64(payment.Amount * 100),
		Currency:         stripe.String(string(stripe.CurrencyINR)),
		PaymentMethod:    stripe.String(paymentMethodId),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
	}
	result, err := paymentintent.New(params)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func CreateSetupIntent(customerId string) (*stripe.SetupIntent, error) {
	params := &stripe.SetupIntentParams{
		Customer: stripe.String(customerId),
		PaymentMethodTypes: []*string{
			stripe.String("card"),
		},
	}
	result, err := setupintent.New(params)
	return result, err
}

func HandlePaymentIntentSucceeded(intentId string) (payment models.Payment, err error) {
	// payment object has attribute thirdPartyPaymentObjectId, which is the id of the payment intent
	// get the payment object from db
	payment, err = models.GetPaymentObjectByThirdPartyPaymentObjectId(intentId)
	fmt.Printf("intentId: %v\n", intentId)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	// update the payment object
	payment.Status = "Succeeded"
	err = payment.Update()
	if err != nil {
		return payment, err
	}
	// update the order object
	order, err := models.GetOrder(payment.OrderID)
	if err != nil {
		return payment, err
	}
	err = order.MarkOrderAsCompleted(payment)
	if err != nil {
		return payment, err
	}
	err = order.MarkOrderAsFulfillable()
	if err != nil {
		return
	}

	// order.PaymentStatus = "Succeeded"
	// order.Payment = payment
	// err = order.Update()
	// if err != nil {
	// 	return payment, err
	// }
	// user, err := models.GetUserByID(order.UserID)
	// if err != nil {
	// 	return payment, err
	// }

	// // a function that runs in a goroutine to send whatsapp message
	// go func() {
	// 	profile, err := user.GetProfile()
	// 	if err != nil {
	// 		return
	// 	}
	// 	// profile.LastUsedPaymentMethod = "stripe_card_" + paymentIntent.PaymentMethod.ID
	// 	if profile.WhatsappEnabled {
	// 		whatsapp.SendOrderConfirmationMessage(user.Phone, order.Address.Name, order.Product.Name, order.ID, "roovo.in")
	// 	}
	// }()

	return payment, nil
}

func HandlePaymentIntentFailed(intentId string) (payment models.Payment, err error) {
	// payment object has attribute thirdPartyPaymentObjectId, which is the id of the payment intent
	// get the payment object from db
	payment, err = models.GetPaymentObjectByThirdPartyPaymentObjectId(intentId)
	if err != nil {
		return
	}
	// update the payment object
	payment.Status = "Failed"
	err = payment.Update()
	if err != nil {
		return payment, err
	}
	// update the order object
	order, err := models.GetOrder(payment.OrderID)
	if err != nil {
		return payment, err
	}
	order.PaymentStatus = "Failed"
	err = order.Update()
	if err != nil {
		return payment, err
	}
	return payment, nil
}

func HandlePaymentIntentCanceled(intentId string) (payment models.Payment, err error) {
	// payment object has attribute thirdPartyPaymentObjectId, which is the id of the payment intent
	// get the payment object from db
	payment, err = models.GetPaymentObjectByThirdPartyPaymentObjectId(intentId)
	if err != nil {
		return
	}
	// update the payment object
	payment.Status = "Canceled"
	err = payment.Update()
	if err != nil {
		return payment, err
	}
	// update the order object
	order, err := models.GetOrder(payment.OrderID)
	if err != nil {
		return payment, err
	}
	order.PaymentStatus = "Canceled"
	err = order.Update()
	if err != nil {
		return payment, err
	}
	return payment, nil
}

func HandlePaymentIntentProcessing(intentId string) (payment models.Payment, err error) {
	// payment object has attribute thirdPartyPaymentObjectId, which is the id of the payment intent
	// get the payment object from db
	payment, err = models.GetPaymentObjectByThirdPartyPaymentObjectId(intentId)
	if err != nil {
		return
	}
	// update the payment object
	payment.Status = "Processing"
	err = payment.Update()
	if err != nil {
		return payment, err
	}
	// update the order object
	order, err := models.GetOrder(payment.OrderID)
	if err != nil {
		return payment, err
	}
	order.PaymentStatus = "Processing"
	err = order.Update()
	if err != nil {
		return payment, err
	}
	return payment, nil
}
