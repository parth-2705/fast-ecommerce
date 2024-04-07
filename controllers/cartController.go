package controllers

import (
	"fmt"
	"hermes/configs/Redis"
	"hermes/models"
	fb "hermes/services/fbAds"
	"hermes/services/shiprocket"
	"hermes/utils/data"
	"hermes/utils/network"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CartPage(context *gin.Context) {
	context.HTML(http.StatusOK, "cartPage", gin.H{})
}

type CartCreateBody struct {
	Quantity  int    `json:"quantity"`
	VariantID string `json:"variantID"`
}

// Creating a Single Product Cart
func CreateCart(c *gin.Context) {

	// dealID := c.Query("deal")

	variantID := getVariantIDFromURL(c)
	if variantID == "" {
		fmt.Println("invalid variant id")
		c.AbortWithError(400, fmt.Errorf("invalid varaint id"))
		return
	}

	userAgentID := data.GetUserAgentIDFromSession(c)

	variant, err := models.GetVariant(variantID)
	if err != nil {
		fmt.Printf("err variant: %v\n", err)
		c.AbortWithError(400, err)
		return
	}

	body := CartCreateBody{}
	err = c.BindJSON(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	if variant.Quantity < body.Quantity {
		fmt.Println("Not enough Quantity")
		c.Redirect(302, fmt.Sprintf("/product/%s", variant.ProductID))
		return
	}

	productId := variant.ProductID
	product, err := models.GetProduct(productId)
	if err != nil {
		fmt.Printf("err product: %v\n", err)
		c.AbortWithError(400, err)
		return
	}

	items := []models.Item{{
		Product:   product,
		Variant:   variant,
		ProductID: productId,
		VariantID: variantID,
		Quantity:  body.Quantity,
	}}

	userID := data.GetUserIDFromSession(c)
	cart, err := models.CreateNewCart2(userID, userAgentID, items, models.Website)
	if err != nil {
		fmt.Printf("err cart: %v\n", err)
		c.AbortWithError(500, err)
		return
	}

	data.SetCartInSession(c, cart.ID)

	go fb.SendAddToCartEvent(cart.ID, data.GetUserAgentFromSession(c), data.GetIPAddressFromSession(c), data.GetFBPCookie(c), data.GetFBCCookie(c))

	c.Redirect(302, "/cart/summary")
}

func CartSummaryPage(c *gin.Context) {

	// dealID := c.Query("deal")

	// Check if Cart ID was supplied in the URL
	cartID, cartIDPresent := c.GetQuery("cartID")
	recoveredCart := false
	if !cartIDPresent {
		cartID = data.GetCartFromSession(c)
	} else {
		data.SetCartInSession(c, cartID)
		recoveredCart = true
	}

	fmt.Printf("cartID: %v\n", cartID)

	cart, err := models.GetCart(cartID)
	if err != nil {
		c.AbortWithError(404, fmt.Errorf("cart not found"))
		return
	}

	user, err := getUserObjectFromSession(c)
	if err != nil {
		fmt.Println(err)
		c.Redirect(http.StatusTemporaryRedirect, "/auth/sign-out")
		// c.AbortWithError(401, err)
		return
	}

	// Check if this Cart belongs to the Logged in User. Log out otherwise
	if cart.UserID != "" && cart.UserID != user.ID {
		fmt.Println("Cart does not belong to User")
		c.Redirect(302, fmt.Sprintf("/product/%s", cart.ProductID))
	}

	// If the cart in session has its order already completed then redirect user to product page
	if cart.OrderID != "" && cart.Status >= models.OrderCompleted {
		c.Redirect(302, fmt.Sprintf("/product/%s", cart.ProductID))
	}

	variant, err := models.GetVariant(cart.Items[0].VariantID)
	if err != nil {
		fmt.Printf("err variant: %v\n", err)
		c.AbortWithError(400, err)
		return
	}

	productId := variant.ProductID
	// addressID := getAddressIDFromURL(c)

	back := "/product/" + productId

	if productId == "" {
		fmt.Println("invalid Product ID")
		c.AbortWithError(400, fmt.Errorf("invalid Product ID"))
		return
	}

	userProfile, err := user.GetProfile()
	if err != nil {
		fmt.Println("err getting user profile: ", err)
		c.AbortWithError(500, err)
		return
	}

	if userProfile.DefaultPaymentMethod == "" {
		err = userProfile.UpdateDefaultPaymentMethod("COD")
		if err != nil {
			fmt.Println("err updating user profile: ", err)
			c.AbortWithError(500, err)
			return
		}

		userProfile.DefaultPaymentMethod = "COD"
		// update session
	}

	data.SetSessionValue(c, "DEFAULT_PAYMENT_METHOD", userProfile.DefaultPaymentMethod)

	if c.Query("payment_method") != "" {
		err = userProfile.UpdateLastUsedPaymentMethod(c.Query("payment_method"))
		if err != nil {
			fmt.Println("err updating user profile: ", err)
			c.AbortWithError(500, err)
			return
		}

		// if payment method is not COD, or CARD --> error
		if c.Query("payment_method") != "COD" && c.Query("payment_method") != "CARD" {
			fmt.Println("invalid payment method")
			c.AbortWithError(400, fmt.Errorf("invalid payment method"))
			return
		}

		// update session
		data.SetSessionValue(c, "DEFAULT_PAYMENT_METHOD", c.Query("payment_method"))

	}

	// Get all te addresses for this User
	addresses, err := user.GetAddresses()
	if err != nil {
		fmt.Println("err getting user addresses: ", err)
		c.AbortWithError(500, err)
		return
	}

	dealID := cart.DealID

	// If User has no addresses saved, redirect him to Address Page
	if len(addresses) == 0 {
		fmt.Println("User has no addresses saved. Redirecting to Address Page")
		if len(dealID) == 0 {
			c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/addresses/new?redirect=cart/summary"))
		} else {
			c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/addresses/new?redirect=cart/summary"))
		}
		return
	}

	addressIDToUse := c.Query("address") //Check if there was an address refered in the URL
	addressToUse := addresses[0]         //Set address to use as the First Address in the User's address list, (or default address)

	if addressIDToUse == "" {
		// Case when no address param was supplied
		addressToUse, err = getAddressObjectFromSession(c) // Get Address stored in session
		if err != nil {
			addressToUse = addresses[0] // if can't pick address from session use the default Adress
		}
	} else {
		// Case when address param was supplied
		// Match the address ID given and use that address
		for _, address := range addresses {
			if address.ID == addressIDToUse {
				addressToUse = address
				break
			}
		}
	}
	setAddressIDInSession(c, addressToUse.ID) // Set the Address in the session

	product, err := models.GetCompleteProduct(productId)
	if err != nil {
		fmt.Printf("err product: %v\n", err)
		c.AbortWithError(400, err)
		return
	}

	seller, err := models.GetSellerByID(product.SellerID)
	if err != nil {
		fmt.Printf("err product: %v\n", err)
		c.AbortWithError(400, err)
		return
	}
	pincodeServicable := true
	pincodeServicable, err = Redis.CheckIfPincodesExistForBrand(product.BrandID, addressToUse.PinCode)
	if err != nil {
		fmt.Printf("pincode redis err: %v\n", err)
	}
	if !product.Brand.SelfFulfilled {
		_, err := shiprocket.CheckServiceability(addressToUse.PinCode, seller.PinCode, variant.Weight)
		if err != nil {
			pincodeServicable = false
		}
	}
	fmt.Printf("pincodeServicable: %v\n", pincodeServicable)
	hasActiveDeal := true
	var deal models.Deal

	if len(dealID) > 0 {
		deal, err = models.GetDealByID(dealID)
		if err != nil {
			fmt.Printf("error: %s", err.Error())
			hasActiveDeal = false
		} else {
			// deal exists on this product
			hasActiveDeal = models.DealExpired(deal)
		}
	} else {
		hasActiveDeal = false
	}

	var coupon models.Coupon
	if cart.CouponID != "" {
		coupon, err = models.GetCouponByID(cart.CouponID)
		if err != nil {
			c.AbortWithError(400, err)
			return
		}
	}

	c.Header("Cache-Control", "no-store, must-revalidate")

	if network.MobileRequest(c) {
		c.JSON(200, gin.H{
			"product":           product,
			"address":           addressToUse,
			"addressOptions":    addresses,
			"back":              back,
			"variant":           variant,
			"paymentMethod":     data.GetSessionValue(c, "DEFAULT_PAYMENT_METHOD"),
			"hasActiveDeal":     hasActiveDeal,
			"deal":              deal,
			"coupon":            coupon,
			"price":             cart.CartAmount,
			"quantity":          cart.Items[0].Quantity,
			"pincodeServicable": pincodeServicable,
		})

		return
	}

	c.HTML(200, "order-summary", gin.H{
		"product":           product,
		"address":           addressToUse,
		"addressOptions":    addresses,
		"back":              back,
		"variant":           variant,
		"paymentMethod":     data.GetSessionValue(c, "DEFAULT_PAYMENT_METHOD"),
		"hasActiveDeal":     hasActiveDeal,
		"deal":              deal,
		"coupon":            coupon,
		"price":             cart.CartAmount,
		"cartID":            cart.ID,
		"quantity":          cart.Items[0].Quantity,
		"recoveredCart":     recoveredCart,
		"pincodeServicable": pincodeServicable,
	})

}

func UpdateItemQuantityInCart(c *gin.Context) {
	cartID := c.Param("cartID")
	body := CartCreateBody{}
	err := c.BindJSON(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
	}
	cart, err := models.GetCart(cartID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
	}

	variant, err := models.GetVariant(body.VariantID)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "variant not found",
		})
		return
	}

	if variant.Quantity < body.Quantity {
		c.JSON(400, gin.H{
			"error": "not enough Quantity",
		})
		return
	}

	err = cart.ChangeQuantityForanItem(body.VariantID, body.Quantity)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	// for idx, item := range cart.Items {
	// 	if item.VariantID == body.VariantID {
	// 		item.Quantity = body.Quantity
	// 		cart.Items[idx] = item
	// 		break
	// 	}
	// }
	// err = cart.Update()
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
	// }
	c.JSON(http.StatusOK, gin.H{
		"price": cart.CartAmount,
	})
}

func getCartFromSession(c *gin.Context) (cart models.Cart, err error) {
	cartID := data.GetCartFromSession(c)
	cart, err = models.GetCart(cartID)
	return
}
