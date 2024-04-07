package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

type OrderAmount struct {
	ProductPrice ProductPrice `json:"productPrice" bson:"productPrice"`
	Coupon       struct {
		CouponID       string  `json:"couponID" bson:"couponID"`
		DiscountAmount float64 `json:"discountAmount" bson:"discountAmount"`
	} `json:"Coupon" bson:"Coupon"`
	PaymentMethodDiscount PaymentMethodDiscount `json:"paymentMethodDiscount" bson:"PaymentMethodDiscount"`
	TotalAmount           float64               `json:"totalAmount" bson:"totalAmount"`
}

type PaymentMethodDiscount struct {
	DiscountPercentage int     `json:"discountPercentage" bson:"discountPercentage"`
	DiscountAmount     float64 `json:"discountAmount" bson:"discountAmount"`
}

func (OrderAmount) GormDataType() string {
	return "json"
}

func (data OrderAmount) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *OrderAmount) Scan(value interface{}) error {

	if value == nil {
		return nil
	}

	var byteSlice []byte
	switch v := value.(type) {
	case []byte:
		if len(v) > 0 {
			byteSlice = make([]byte, len(v))
			copy(byteSlice, v)
		}
	case string:
		byteSlice = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	err := json.Unmarshal(byteSlice, &data)
	return err
}

func CreateNewAmountObject(productPrice ProductPrice) (amount OrderAmount) {
	amount.ProductPrice = productPrice
	amount.CalculateTotalPrice()
	return
}

func (amount *OrderAmount) CalculateTotalPrice() {
	amount.TotalAmount = amount.ProductPrice.SellingPrice - amount.Coupon.DiscountAmount
}
