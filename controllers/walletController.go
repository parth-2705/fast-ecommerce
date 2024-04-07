package controllers

// type cartAmountModificationRequest struct {
// 	CartID string `json:"cartID"`
// }

// func ApplyWalletBalanceToCart(ctx *gin.Context) {

// 	var requestBody cartAmountModificationRequest
// 	err := ctx.BindJSON(&requestBody)
// 	if err != nil {
// 		rw.JSONErrorResponse(ctx, 400, err)
// 		return
// 	}

// 	cart, err := models.GetCart(requestBody.CartID)
// 	if err != nil {
// 		rw.JSONErrorResponse(ctx, 404, err)
// 		return
// 	}

// 	err = cart.UseWalletBalance()
// 	if err != nil {
// 		rw.JSONErrorResponse(ctx, 500, err)
// 		return
// 	}

// 	ctx.JSON(200, cart.CartAmount)
// 	return
// }

// func RemoveWalletBalanceFromCart(ctx *gin.Context) {

// 	var requestBody cartAmountModificationRequest
// 	err := ctx.BindJSON(&requestBody)
// 	if err != nil {
// 		rw.JSONErrorResponse(ctx, 400, err)
// 		return
// 	}

// 	cart, err := models.GetCart(requestBody.CartID)
// 	if err != nil {
// 		rw.JSONErrorResponse(ctx, 404, err)
// 		return
// 	}

// 	err = cart.RemoveWalletBalance()
// 	if err != nil {
// 		rw.JSONErrorResponse(ctx, 500, err)
// 		return
// 	}

// 	ctx.JSON(200, cart.CartAmount)
// 	return
// }
