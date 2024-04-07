package payments

import (
	"hermes/utils/data"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/paymentmethod"
	"github.com/stripe/stripe-go/v74/token"
)

type SaveCardForm struct {
	Number   string `json:"card_number" binding:"required" form:"card_number"`
	ExpMonth string `json:"card_expiry_month" binding:"required" form:"card_expiry_month"`
	ExpYear  string `json:"card_expiry_year" binding:"required" form:"card_expiry_year"`
	Nickname string `json:"card_nickname" form:"card_nickname"`
}

func CreateCardToken(customerId string, card SaveCardForm) (*stripe.Token, error) {
	params := &stripe.TokenParams{
		Card: &stripe.CardParams{
			Number:   stripe.String(card.Number),
			ExpMonth: stripe.String(card.ExpMonth),
			ExpYear:  stripe.String(card.ExpYear),
			Customer: stripe.String(customerId),
		},
	}
	return token.New(params)
}

func CreateCardPaymentMethod(customerId string, card SaveCardForm) (*stripe.PaymentMethod, error) {
	params := &stripe.PaymentMethodParams{
		Type: stripe.String(string(stripe.PaymentMethodTypeCard)),
		Card: &stripe.PaymentMethodCardParams{
			Number:   stripe.String(card.Number),
			ExpMonth: stripe.Int64(data.StringToInteger(card.ExpMonth)),
			ExpYear:  stripe.Int64(data.StringToInteger(card.ExpYear)),
		},
	}
	return paymentmethod.New(params)
}
