package models

type paymentOption struct {
	ID        string      `json:"id" bson:"_id"`
	MethodID  string      `json:"methodID" bson:"methodID"`
	Icon      string      `json:"icon" bson:"icon"`
	Name      string      `json:"name" bson:"name"`
	Providers []string    `json:"providers" bson:"providers"`
	Amount    OrderAmount `json:"amount" bson:"amount"`
}

var paymentOptions []paymentOption = []paymentOption{
	{
		ID:        "34c8af26-bde2-4c51-9857-04024124fec8",
		MethodID:  "newCard",
		Icon:      "/static/assets/images/payment-cards.svg",
		Name:      "Credit Card/Debit Card",
		Providers: []string{"/static/icons/visaIcon.png", "/static/icons/masterCardIcon.png", "/static/icons/rupayIcon.png"},
	},
	{
		ID:        "2d021958-de34-468f-9b5b-ba1b157d5c7c",
		MethodID:  "UPI",
		Icon:      "/static/assets/images/payment-upi.svg",
		Name:      "UPI",
		Providers: []string{"/static/icons/gpayIcon.png", "/static/icons/paytmIcon.png", "/static/icons/phonepeIcon.png"},
	},
	{
		ID:        "5e3a0aba-0509-43e7-8144-b2d1c63952b3",
		MethodID:  "COD",
		Icon:      "/static/assets/images/payment-cod.svg",
		Name:      "Cash On Delivery",
		Providers: []string{},
	},
}

func GetPaymentOptionsMap(amount OrderAmount) map[string]paymentOption {
	dst := make(map[string]paymentOption)

	for _, po := range paymentOptions {
		po.Amount = amount
		dst[po.ID] = po
	}

	return dst
}

func GetPaymentOptions(amount OrderAmount, cartAmount float64) []paymentOption {
	dst := make([]paymentOption, len(paymentOptions))
	copy(dst, paymentOptions)
	
	for i, method := range dst {
		method.Amount = amount
		method.Amount.AddPaymentMethodDiscountAndCalculateDiscount(method.MethodID, cartAmount)
		dst[i] = method
	}

	return dst
}

func CopyPaymentOptions() []paymentOption {
	dst := make([]paymentOption, len(paymentOptions))
	copy(dst, paymentOptions)
	return dst
}

func GetEmptyPaymentOptions() []paymentOption {
	paymentoptions := make([]paymentOption, 0)
	return paymentoptions
}

type PaymentMethodMap map[string]PaymentMethodConfiguration

type PaymentMethodConfiguration struct {
	Available bool `json:"available" bson:"available"`
}
