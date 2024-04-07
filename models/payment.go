package models

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"hermes/db"
	"hermes/utils/data"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type paymentMethod string

const (
	UPI    paymentMethod = "UPI"
	WALLET paymentMethod = "wallet"
)

type Payment struct {
	ID                        string      `json:"id" bson:"_id"`
	CreatedAt                 int64       `json:"createdAt" bson:"createdAt"`
	UpdatedAt                 int64       `json:"updatedAt" bson:"updatedAt"`
	OrderID                   string      `json:"orderId" bson:"orderId"`
	Method                    string      `json:"method" bson:"method"`
	Amount                    int64       `json:"amount" bson:"amount"`
	Currency                  string      `json:"currency" bson:"currency"`
	Status                    string      `json:"status" bson:"status"`
	ThirdPartyPaymentObjectId string      `json:"thirdPartyPaymentObjectId" bson:"thirdPartyPaymentObjectId"`
	ThirdPartyPaymentObject   interface{} `json:"thirdPartyPaymentObject" bson:"thirdPartyPaymentObject"`
}

func (Payment) GormDataType() string {
	return "json"
}

func (data Payment) Value() (driver.Value, error) {
	jsonifiedData, _ := json.Marshal(data)
	return string(jsonifiedData), nil
}

func (data *Payment) Scan(value interface{}) error {

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

func (payment Payment) Create() (newPayment Payment, err error) {
	payment.CreatedAt = time.Now().Unix()
	payment.UpdatedAt = time.Now().Unix()
	_, err = db.PaymentCollection.InsertOne(context.Background(), payment)
	if err != nil {
		fmt.Println(err)
		return newPayment, err
	}
	return payment, nil
}

func (payment Payment) Update() (err error) {
	payment.UpdatedAt = time.Now().Unix()
	_, err = db.PaymentCollection.UpdateOne(context.Background(), bson.M{"_id": payment.ID}, bson.M{"$set": payment})
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func GetPaymentById(id string) (payment Payment, err error) {
	err = db.PaymentCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&payment)
	if err != nil {
		return
	}
	return
}

func GetPaymentObjectByThirdPartyPaymentObjectId(id string) (payment Payment, err error) {
	err = db.PaymentCollection.FindOne(context.Background(), bson.M{"thirdPartyPaymentObjectId": id}).Decode(&payment)
	if err != nil {
		return
	}
	return
}

func (payment *Payment) MarkPaymentAsSuccess() (err error) {
	payment.Status = "Succeeded"
	err = payment.Update()
	if err != nil {
		return err
	}

	order, err := GetOrder(payment.OrderID)
	if err != nil {
		return err
	}
	err = order.MarkOrderAsCompleted(*payment)
	if err != nil {
		return err
	}
	err = order.MarkOrderAsFulfillable()
	if err != nil {
		return
	}

	return nil
}

func (payment *Payment) MarkPaymentAsFailed() (err error) {
	payment.Status = "Failed"
	err = payment.Update()
	if err != nil {
		return err
	}

	order, err := GetOrder(payment.OrderID)
	if err != nil {
		return err
	}

	order.PaymentStatus = "Failed"
	order.Payment = *payment
	err = order.Update()
	if err != nil {
		return err
	}

	return nil
}

func (payment *Payment) MarkPaymentasExpired() (err error) {
	payment.Status = "Expired"
	err = payment.Update()
	if err != nil {
		return err
	}

	order, err := GetOrder(payment.OrderID)
	if err != nil {
		return err
	}

	order.PaymentStatus = "Failed"
	order.Payment = *payment
	err = order.Update()
	if err != nil {
		return err
	}

	return nil
}

// func CreateUPIPaymentLink(payment *Payment, message string) (string, error) {

// 	response, thirdpartyResponse, err := payments.InitiateUPIPayment(payment.ID, float64(payment.Amount), message)
// 	if err != nil {
// 		return "", err
// 	}

// 	payment.ThirdPartyPaymentObject = thirdpartyResponse
// 	payment.ThirdPartyPaymentObjectId = response.ThirdPartyTransactionID

// 	err = payment.Update()
// 	if err != nil {
// 		return "", nil
// 	}

// 	return response.UPIDeepLink, nil

// }

type UPIPay struct {
	Link      string `json:"link" bson:"link"`
	PaymentID string `json:"paymentID" bson:"paymentID"`
	ID        string `json:"id" bson:"_id"`
}

func SaveUPIPayObject(paymentLink string, paymentID string) (upiPay UPIPay, err error) {
	upiPay.Link = paymentLink
	upiPay.PaymentID = paymentID
	upiPay.ID = data.GetUUIDStringWithoutPrefix()
	_, err = db.PayCollection.InsertOne(context.Background(), upiPay)
	return
}

func GetUPIPayObject(ID string) (upiPay UPIPay, err error) {
	err = db.PayCollection.FindOne(context.Background(), bson.M{"_id": ID}).Decode(&upiPay)
	return
}

func GetUPIPayObjectByPayment(PaymentID string) (upiPay UPIPay, err error) {
	err = db.PayCollection.FindOne(context.Background(), bson.M{"paymentID": PaymentID}).Decode(&upiPay)
	return
}
