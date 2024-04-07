package payments

import (
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/paymentmethod"
)

func CreateStripeCustomer(phone string) (id string, err error) {
	params := &stripe.CustomerParams{
		Phone: stripe.String(phone),
	}
	c, err := customer.New(params)
	return c.ID, err
}

func GetStripeCustomer(id string) (c *stripe.Customer, err error) {
	return customer.Get(id, nil)
}

func ListPaymentMethodsByCustomer(customerId string) (paymentMethods []*stripe.PaymentMethod, err error) {
	if customerId == "" {
		return []*stripe.PaymentMethod{}, nil
	}
	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(customerId),
		Type:     stripe.String(string(stripe.PaymentMethodTypeCard)),
	}
	i := paymentmethod.List(params)
	for i.Next() {
		pm := i.PaymentMethod()
		paymentMethods = append(paymentMethods, pm)
	}
	return paymentMethods, i.Err()
}
