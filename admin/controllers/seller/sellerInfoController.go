package controllers

import (
	ctx "context"
	"fmt"
	"hermes/admin/services/auth"
	"hermes/db"
	"hermes/services/shiprocket"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SellerInfoPage(c *gin.Context) {
	seller, err := auth.GetSellerFromSession(c)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	if seller.ProfileCompleted {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	c.HTML(http.StatusOK, "seller-info", gin.H{
		"title": "Roovo", "seller": seller,
	})
}

// func NewSellerInformation(c *gin.Context) {

// 	seller, err := auth.GetSellerFromSession(c)
// 	if err != nil {
// 		c.AbortWithError(http.StatusUnauthorized, err)
// 		return
// 	}

// 	if err := c.ShouldBind(&seller); err != nil {
// 		fmt.Printf("err: %v\n", err)
// 		c.AbortWithError(http.StatusBadRequest, err)
// 		return
// 	}

// 	seller.ProfileCompleted = true

// 	err = seller.Update()
// 	if err != nil {
// 		fmt.Printf("err: %v\n", err)
// 		c.AbortWithError(http.StatusBadGateway, err)
// 		return
// 	}

// 	err = shiprocket.CreateShipRocketPickUpAddress(seller)
// 	if err != nil {
// 		fmt.Printf("err: %v\n", err)
// 		c.AbortWithError(http.StatusBadGateway, err)
// 		return
// 	}

// 	c.Redirect(http.StatusFound, "/")
// }

func SellerInfo(c *gin.Context) {
	seller, err := auth.GetSellerFromSession(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"seller": seller})
}

// func MagicLinkLogin(c *gin.Context) {
// 	code := c.Param("code")
// 	endpoint := whatsapp.WHATSAPP_SERVICE_BASE_URL + "/magic"
// 	jsonBody := map[string]string{
// 		"code": code,
// 	}
// 	body, _ := json.Marshal(jsonBody)
// 	headers := [][]string{
// 		{"Authorization", fmt.Sprintf("Bearer %s", whatsapp.WHATSAPP_SERVICE_API_KEY)},
// 	}
// 	status, resp, err := themis.HitAPIEndpoint2(endpoint, "POST", body, headers, nil)
// 	fmt.Println("RESP MAGIC LOGIN", string(resp), endpoint)
// 	if err != nil {
// 		fmt.Println("error in sending request", err)
// 		return
// 	}
// 	response := map[string]string{}
// }

func NewSellerInformationV2(c *gin.Context) {

	seller, err := auth.GetSellerFromSession(c)
	sellerMember, err := auth.GetSellerMemberFromSession(c)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	if err := c.ShouldBind(&seller); err != nil {
		fmt.Printf("err: %v\n", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	sellerMember.Name = seller.Name
	sellerMember.Phone = seller.Phone
	seller.ProfileCompleted = true

	err = seller.Update()
	if err != nil {
		fmt.Printf("err: %v\n", err)
		c.AbortWithError(http.StatusBadGateway, err)
		return
	}

	err = sellerMember.Update()
	if err != nil {
		fmt.Printf("err: %v\n", err)
		c.AbortWithError(http.StatusBadGateway, err)
		return
	}

	err = shiprocket.CreateShipRocketPickUpAddress(seller)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		c.AbortWithError(http.StatusBadGateway, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": "Information Added"})
}

func GetDataByPincode(context *gin.Context) {
	// Get the pincode from the URL
	pincode := context.Param("pincode")
	var response db.Pincode
	// query is a map[string]interface{}
	query := map[string]interface{}{
		"pincode": pincode,
	}
	// Get the data from the database
	err := db.PincodeCollection.FindOne(ctx.Background(), query).Decode(&response)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Return the data as JSON
	context.JSON(http.StatusOK, response)
}
