package common

import (
	"context"
	"fmt"
	"hermes/db"
	"hermes/models"
	"hermes/utils/num2words"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func GenerateInvoiceUtil(c *gin.Context) ([]models.Order, models.Seller, []models.Shipping, error) {
	awbCode := c.Query("awb")

	var orders []models.Order
	var seller models.Seller
	var shippings []models.Shipping

	if len(awbCode) == 0 {
		return orders, seller, shippings, fmt.Errorf("AWB empty")
	}

	curr, err := db.ShippingCollection.Find(context.Background(), bson.M{"awb": awbCode})
	if err != nil {
		return orders, seller, shippings, fmt.Errorf("error finding shipping details: %s", err.Error())
	}

	err = curr.All(context.Background(), &shippings)
	if err != nil {
		return orders, seller, shippings, fmt.Errorf("error reading shippings details: %s", err.Error())
	}

	for _, shipping := range shippings {
		var tempOrder models.Order
		err := db.OrderCollection.FindOne(context.Background(), bson.M{"_id": shipping.Id}).Decode(&tempOrder)
		if err != nil {
			return orders, seller, shippings, fmt.Errorf("error finding order by ID: %s %s", err.Error(), shipping.Id)
		}
		orders = append(orders, tempOrder)
	}

	fmt.Println(len(orders))
	err = db.SellerCollection.FindOne(context.Background(), bson.M{"_id": orders[0].Product.SellerID}).Decode(&seller)
	if err != nil {
		return orders, seller, shippings, err
	}

	return orders, seller, shippings, nil
}

func GenerateInvoice(c *gin.Context) {
	orders, seller, shippings, err := GenerateInvoiceUtil(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	totalPayment := 0
	totalCartDiscount := 0
	totalAdditionalDiscount := 0

	for _, order := range orders {
		totalPayment += int(order.Cart.CartAmount.TotalAmount)
		totalAdditionalDiscount += int(order.AdditionalDiscount)
		totalCartDiscount += int(order.Cart.CartAmount.Coupon.DiscountAmount)
	}

	netAmount := totalPayment - totalAdditionalDiscount
	amountToWords := num2words.ConvertAnd(netAmount)
	c.HTML(http.StatusOK, "invoice.html", gin.H{"orders": orders, "seller": seller, "shipping": shippings[0], "amountToWords": amountToWords, "netAmount": netAmount, "totalCartDiscount": totalCartDiscount, "order": orders[0]})
}

func GenerateSellerInvoice(c *gin.Context) {
	orders, seller, shippings, err := GenerateInvoiceUtil(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	totalPayment := 0
	totalCartDiscount := 0
	totalAdditionalDiscount := 0

	for _, order := range orders {
		totalPayment += int(order.Cart.CartAmount.TotalAmount)
		totalAdditionalDiscount += int(order.AdditionalDiscount)
		totalCartDiscount += int(order.Cart.CartAmount.Coupon.DiscountAmount)
	}

	netAmount := totalPayment - totalAdditionalDiscount
	amountToWords := num2words.ConvertAnd(netAmount)

	c.HTML(http.StatusOK, "seller-invoice.html", gin.H{"orders": orders, "seller": seller, "shipping": shippings[0], "amountToWords": amountToWords, "netAmount": netAmount, "totalCartDiscount": totalCartDiscount, "order": orders[0]})
}
