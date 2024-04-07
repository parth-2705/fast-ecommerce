package controllers

import (
	"fmt"
	"hermes/controllers"
	"hermes/models"
	"hermes/utils/rw"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type AdminCart struct {
	models.Cart    `bson:",inline"`
	models.User    `json:"user" bson:"user"`
	models.Product `json:"product" bson:"product"`
}

type CreateCartRequest struct {
	UserID    string            `json:"userID"`
	Items     []models.Item     `json:"items"`
	Source    models.PortalType `json:"source"`
	UserPhone string            `json:"userPhone"`
}

func GetCartsList(c *gin.Context) {

	limitInt := controllers.GetLimitFromQueryValue(c.Query("limit"))
	pageInt := controllers.GetPageFromQueryValue(c.Query("page"))

	var Paginater controllers.Pagination = controllers.Pagination{
		Limit: limitInt,
		Page:  pageInt,
	}

	var cartStruct []*AdminCart
	preSkip := bson.A{bson.D{{Key: "$sort", Value: bson.D{{Key: "updatedAt", Value: -1}}}}}
	postSkip := bson.A{bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "users"},
			{Key: "localField", Value: "userID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "user"},
		}},
	},
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$user"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		bson.D{{Key: "$addFields", Value: bson.D{{Key: "internal", Value: "$user.internal"}}}},
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "products"},
				{Key: "localField", Value: "productID"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "product"},
			}},
		},
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$product"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}}}

	carts, err := controllers.Paginate("cart", &Paginater, cartStruct, preSkip, postSkip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"carts":      carts.Rows,
		"totalRows":  carts.TotalRows,
		"totalPages": carts.TotalPages,
	})
}

func RecoverCartOnWhatsapp(c *gin.Context) {

	type requestStruct struct {
		CartID string `json:"cartID"`
	}

	requestBody := requestStruct{}

	err := c.BindJSON(&requestBody)
	if err != nil {
		rw.JSONErrorResponse(c, 400, fmt.Errorf("invalid Request Format"))
		return
	}

	cart, err := models.GetCart(requestBody.CartID)
	if err != nil {
		rw.JSONErrorResponse(c, 404, err)
		return
	}

	err = cart.MarkAsRecovered()
	if err != nil {
		rw.JSONErrorResponse(c, 500, err)
		return
	}

	c.JSON(200, cart)
}

func CreateNewCart(c *gin.Context) {
	request := CreateCartRequest{}
	err := c.BindJSON(&request)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	if request.Source == models.Whatsapp {
		user, err := models.GetUser(request.UserPhone)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		request.UserID = user.ID
	}
	cart, err := models.CreateNewCart2(request.UserID, "", request.Items, request.Source)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	amount := cart.CartAmount
	amount.AddPaymentMethodDiscountAndCalculateDiscount("notCOD", cart.CartAmount.TotalAmount)
	c.JSON(http.StatusOK, gin.H{"cart": cart, "discount": amount.PaymentMethodDiscount.DiscountAmount})
}
