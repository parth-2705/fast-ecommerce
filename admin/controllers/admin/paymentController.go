package controllers

import (
	"fmt"
	"hermes/models"
	"hermes/services/Sentry"
	"hermes/utils/data"
	"hermes/utils/payments"
	"hermes/utils/whatsapp"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func ConvertPaymentToPrepaid(ctx *gin.Context) {
	orderID := ctx.Query("orderID")
	order, err := models.GetOrder(orderID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Order ID Invalid"})
		return
	}
	if order.Fulfillable {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Order already fulfillable"})
		return
	}
	if len(order.CancellationReason) > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Order cancelled"})
		return
	}
	if time.Now().Sub(order.CreatedAt) >= time.Hour {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Deal Expired"})
		return
	}
	amount := order.OrderAmount
	amount.AddPaymentMethodDiscountAndCalculateDiscount(string(models.UPI), order.Cart.CartAmount.TotalAmount)
	totalAmount := int64(amount.TotalAmount)
	paymentID := data.GetUUIDString("payment")
	tags := map[string]string{
		"user":   order.UserID,
		"action": "payment-convert",
		"phone":  order.Address.Phone,
	}
	upiResponse, decentroResponse, err := payments.InitiateUPIPayment(paymentID, float64(totalAmount), "Roovo Payment")
	if err != nil {
		Sentry.SendErrorToSentry(ctx, err, tags)
		fmt.Println("UPI err:", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	newPayment := models.Payment{
		ID:                      paymentID,
		OrderID:                 order.ID,
		Amount:                  int64(amount.TotalAmount),
		Status:                  "Initiated",
		Currency:                "INR",
		Method:                  string(models.UPI),
		ThirdPartyPaymentObject: decentroResponse,
	}
	newPayment, err = newPayment.Create()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Error"})
		return
	}
	link := upiResponse.UPIDeepLinks.GeneratedLink
	_, err = models.SaveUPIPayObject(link, newPayment.ID)
	if err != nil {
		Sentry.SendErrorToSentry(ctx, err, tags)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	upiPay, err := models.GetUPIPayObjectByPayment(newPayment.ID)
	if err != nil {
		Sentry.SendErrorToSentry(ctx, err, tags)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	phone := strings.ReplaceAll(order.Address.Phone, "+", "")
	err = whatsapp.SendPaymentLinkMessage(phone, upiPay.ID, newPayment.Amount)
	if err != nil {
		Sentry.SendErrorToSentry(ctx, err, tags)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}
