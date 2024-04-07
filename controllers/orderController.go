package controllers

import (
	"fmt"
	"hermes/configs"
	"hermes/models"
	"hermes/models/Logs"
	"hermes/search"
	"hermes/services/Sentry"
	"hermes/services/shiprocket"
	"hermes/utils/data"
	"hermes/utils/network"
	"hermes/utils/payments"
	"hermes/utils/rw"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type Cancellation struct {
	Reason string `json:"reason" bson:"reason"`
}

func OrderPostHandlerV2(c *gin.Context) {

	cartID := c.PostForm("cartID")
	// cartID := data.GetCartFromSession(c)
	cart, err := models.GetCart(cartID)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	// Check if this was a recovered Cart
	recoveredCartStr := c.PostForm("recoveredCart")
	recoveredCart := recoveredCartStr == "true"

	// If the cart has its order already completed then redirect user to product page
	if cart.OrderID != "" && cart.Status >= models.OrderCompleted {
		c.Redirect(302, fmt.Sprintf("/product/%s", cart.ProductID))
	}

	dealID := cart.DealID
	address := c.Query("address")

	variantID := cart.VariantID
	if variantID == "" {
		fmt.Println("invalid Product ID or AddressId")
		c.AbortWithError(400, fmt.Errorf("invalid Product ID or AddressId"))
		return
	}

	variant, err := models.GetVariant(variantID)
	if err != nil {
		fmt.Printf("err variant: %v\n", err)
		c.AbortWithError(400, err)
		return
	}

	productId := variant.ProductID
	user, err := getUserObjectFromSession(c)
	if err != nil {
		fmt.Println("err getting user from session: ", err)
		fmt.Println(err)
		c.AbortWithError(401, err)
		return
	}

	userProfile, err := user.GetProfile()
	if err != nil {
		fmt.Println("err getting user profile: ", err)
		c.AbortWithError(500, err)
		return
	}

	if productId == "" {
		fmt.Println("invalid Product ID or AddressId")
		c.AbortWithError(400, fmt.Errorf("invalid Product ID or AddressId"))
	}

	product, err := models.GetCompleteProduct(productId)
	if err != nil {
		fmt.Printf("err product: %v\n", err)
		c.AbortWithError(400, err)
		return
	}

	paymentMethodsFromStripe, err := payments.ListPaymentMethodsByCustomer(userProfile.StripeCustomerID)
	if err != nil {
		fmt.Printf("err payment methods: %v \n", err)
		c.AbortWithError(400, err)
		return
	}

	type paymentMethodData struct {
		ID   string `json:"id"`
		Card struct {
			Last4 string `json:"last4"`
			Brand string `json:"brand"`
			Type  string `json:"type"`
		} `json:"card"`
		Type string `json:"type"`
	}
	paymentMethods := []paymentMethodData{}

	for _, paymentMethod := range paymentMethodsFromStripe {
		var p paymentMethodData
		p.ID = paymentMethod.ID
		p.Type = string(paymentMethod.Type)
		if p.Type == "card" {
			p.Card.Last4 = paymentMethod.Card.Last4
			p.Card.Brand = string(paymentMethod.Card.Brand)
			p.Card.Type = string(paymentMethod.Card.Funding)
		}
		paymentMethods = append(paymentMethods, p)
	}

	// The Address that was chosen by the User is supplied to this handler in the query Params.
	// Here it is set in the session.
	// Ideally, when in the Order Summary Page the Address is set, a request should be made to the server to set the address in session.
	if len(address) > 0 {
		setAddressIDInSession(c, address)
	}

	addressObj, err := getAddressObjectFromSession(c)
	if err != nil {
		fmt.Printf("err address: %v\n", err)
		c.AbortWithError(400, err)
		return
	}

	deal, err := models.GetDealByID(dealID)
	if err != nil {
		if dealID != "" {
			fmt.Printf("err deal: %v\n", err)
			c.AbortWithError(404, err)
			return
		}
	}

	coupon, _ := models.GetCouponByID(cart.CouponID) // Need empty Coupon in case of error, thats why error is ignored

	// Create a new Unpaid and Unfulfilled Order. maintain this orders state using the orderID throughout the placing order flow
	newOrder := models.Order{
		ID:                data.GetUUIDString("Order"),
		CreatedAt:         time.Now(),
		User:              user,
		UserID:            user.ID,
		Product:           product,
		Address:           addressObj,
		Variant:           variant,
		VariantID:         variant.ID,
		PaymentStatus:     "Unpaid",
		FulfillmentStatus: "Unfulfilled",
		Brand:             product.Brand,
		DealID:            dealID,
		Deal:              deal,
		CartID:            cartID,
		Cart:              cart,
		Coupon:            coupon,
		OrderAmount:       cart.CartAmount,
		Source:            models.Website,
		Session: Logs.Session{
			ID:        data.GetUserAgentIDFromSession(c),
			UserAgent: data.GetUserAgentFromSession(c),
			IPAddress: data.GetIPAddressFromSession(c),
			FBC:       data.GetFBCCookie(c),
			FBP:       data.GetFBPCookie(c),
		},
		IPAddress: data.GetIPAddressFromSession(c),
		UserAgent: data.GetUserAgentFromSession(c),
		UTMParams: data.GetUTMParamsFromSession(c),
	}

	err = newOrder.Create()
	if err != nil {
		fmt.Printf("err order: %v\n", err)
		c.AbortWithError(500, err)
		return
	}

	defaultAddress := user.GetDefaultAddress()
	if defaultAddress.ID == "" {
		addressObj.IsDefault = true
		err := addressObj.UpdateInDB()
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
	}

	cart.OrderCreated(user.ID, newOrder.ID, recoveredCart)
	// data.RemoveCartFromSession(c)

	data.SetSessionValue(c, configs.OrderKey, newOrder.ID)

	if math.Floor(newOrder.OrderAmount.TotalAmount) == 0 {
		newOrder.Payment.Amount = 0
		newOrder.Payment.Method = "Prepaid"
		err = newOrder.MarkOrderAsCompleted(newOrder.Payment)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		err = newOrder.MarkOrderAsFulfillable()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		c.Redirect(http.StatusFound, "/order/success?orderId="+newOrder.ID)
		return
	}

	// data.SetSessionValue(c, "variantId", variantID)
	// data.SetSessionValue(c, "dealID", dealID)
	c.Redirect(http.StatusFound, fmt.Sprintf("/order/select-payment?orderId=%s", newOrder.ID))
}

func OrderSuccessHandler(c *gin.Context) {

	_, err := getUserObjectFromSession(c)
	if err != nil {
		fmt.Println(err)
		c.AbortWithError(400, err)
		return
	}

	orderId := c.Query("orderId")
	if orderId == "" {
		fmt.Println("invalid Order ID")
		c.AbortWithError(400, fmt.Errorf("invalid Order ID"))
		return
	}

	order, err := models.GetOrder(orderId)
	if err != nil {
		fmt.Printf("err order: %v \n", err)
		c.AbortWithError(400, err)
		return
	}

	if order.PaymentStatus != "Paid" {
		fmt.Println("Order not Complete Yet")
		c.AbortWithStatus(400)
	}

	product, err := models.GetProduct(order.Product.ID)
	if err != nil {
		fmt.Printf("err product: %v \n", err)
		c.AbortWithError(400, err)
		return
	}

	brand, err := models.GetBrand(product.BrandID)
	if err != nil {
		fmt.Printf("err brand: %v \n", err)
		c.AbortWithError(400, err)
		return
	}

	dealApplied := true
	var deal models.Deal

	if len(order.DealID) > 0 {
		deal, err = models.GetDealByID(order.DealID)
		if err != nil {
			fmt.Printf("error: %s", err.Error())
			dealApplied = false
		}
	} else {
		dealApplied = false
	}

	data.SetSessionValue(c, configs.OrderKey, orderId)
	c.Header("Cache-Control", "no-store, must-revalidate")

	var Paginater Pagination = Pagination{
		Limit: 20,
		Page:  1,
	}

	processedAppliedFilterMap := map[string]search.FilterObject{}
	processedAppliedFilterMap["category"] = search.FilterObject{
		Values:   []string{product.Category},
		Operator: "=",
		Path:     "category",
	}
	processedAppliedFilterMap["product"] = search.FilterObject{
		Values:   []string{product.ID},
		Operator: "!=",
		Path:     "id",
	}
	sortArr := []search.SortObject{}
	similarProducts, err := ProductPaginate2(&Paginater, "", processedAppliedFilterMap, sortArr)
	if err != nil {
		fmt.Println(err)
		c.AbortWithError(500, err)
		return
	}

	if network.MobileRequest(c) {
		c.JSON(200, gin.H{
			"order":            order,
			"product":          product,
			"product_image":    product.Images[0],
			"brand":            brand,
			"dealApplied":      dealApplied,
			"deal":             deal,
			"similarProducts":  similarProducts,
			"similarPaginater": Paginater,
		})

		return
	}

	c.HTML(200, "order-confirmed", gin.H{
		"order":            order,
		"product":          product,
		"product_image":    product.Images[0],
		"brand":            brand,
		"dealApplied":      dealApplied,
		"deal":             deal,
		"similarProducts":  similarProducts,
		"similarPaginater": Paginater,
	})
}

func OrderPendingHandler(c *gin.Context) {
	orderId := c.Query("orderId")
	if orderId == "" {
		fmt.Println("invalid Order ID")
		c.AbortWithError(400, fmt.Errorf("invalid Order ID"))
		return
	}

	order, err := models.GetOrder(orderId)
	if err != nil {
		fmt.Printf("err order: %v \n", err)
		c.AbortWithError(400, err)
		return
	}

	product, err := models.GetProduct(order.Product.ID)
	if err != nil {
		fmt.Printf("err product: %v \n", err)
		c.AbortWithError(400, err)
		return
	}

	brand, err := models.GetBrand(product.BrandID)
	if err != nil {
		fmt.Printf("err brand: %v \n", err)
		c.AbortWithError(400, err)
		return
	}

	if network.MobileRequest(c) {
		c.JSON(200, gin.H{
			"order":         order,
			"product":       product,
			"product_image": product.Images[0],
			"brand":         brand,
		})

		return
	}

	c.HTML(200, "order-pending.html", gin.H{
		"order":         order,
		"product":       product,
		"product_image": product.Images[0],
		"brand":         brand,
	})
}

func OrderFailureHandler(c *gin.Context) {
	orderId := c.Query("orderId")
	if orderId == "" {
		fmt.Println("invalid Order ID")
		c.AbortWithError(400, fmt.Errorf("invalid Order ID"))
		return
	}

	order, err := models.GetOrder(orderId)
	if err != nil {
		fmt.Printf("err order: %v \n", err)
		c.AbortWithError(400, err)
		return
	}

	product, err := models.GetProduct(order.Product.ID)
	if err != nil {
		fmt.Printf("err product: %v \n", err)
		c.AbortWithError(400, err)
		return
	}

	brand, err := models.GetBrand(product.BrandID)
	if err != nil {
		fmt.Printf("err brand: %v \n", err)
		c.AbortWithError(400, err)
		return
	}

	if network.MobileRequest(c) {
		c.JSON(200, gin.H{
			"order":         order,
			"product":       product,
			"product_image": product.Images[0],
			"brand":         brand,
		})

		return
	}

	c.HTML(200, "order-failed.html", gin.H{
		"order":         order,
		"product":       product,
		"product_image": product.Images[0],
		"brand":         brand,
	})
}

func SelectPaymentPage(c *gin.Context) {
	user, err := getUserObjectFromSession(c)
	if err != nil {
		fmt.Println(err)
		c.AbortWithError(401, err)
		return
	}

	userProfile, err := user.GetProfile()
	if err != nil {
		fmt.Println("err getting user profile: ", err)
		c.AbortWithError(500, err)
		return
	}

	// Get the Order Stored in Session
	prevOrderID := data.GetSessionValue(c, configs.OrderKey).(string)
	prevOrder, err := models.GetOrder(prevOrderID)
	if err == nil { // If Order was found and payment is complete, redirect back to home
		if prevOrder.PaymentStatus == "Paid" {
			fmt.Println("This Order has been processed")
			c.Redirect(302, "/")
		}
	}

	if userProfile.StripeCustomerID == "" {
		fmt.Println("user has no stripe customer ID")
		// create a new customer
		stripeCustomerId, err := payments.CreateStripeCustomer(user.Phone)
		if err != nil {
			fmt.Printf("err stripe customer: %v\n", err)
			c.AbortWithError(500, err)
			return
		}
		userProfile.StripeCustomerID = stripeCustomerId
		err = userProfile.Update()
		if err != nil {
			fmt.Printf("err update user profile: %v\n", err)
			c.AbortWithError(500, err)
			return
		}
	}

	wallet, err := models.GetUserWallet(user.ID)
	if err != nil {
		rw.JSONErrorResponse(c, 500, err)
		return
	}

	var product models.Product
	var variant models.Variation
	var order models.Order
	var deal models.Deal
	hasActiveDeal := true

	// check if redirect_status param is present
	redirectStatus := c.Query("redirect_status")
	if redirectStatus == "" {

		// Fetch Order Details Using the Order ID
		orderId := c.Query("orderId")
		if orderId == "" {
			fmt.Println("invalid Order ID")
			c.AbortWithError(400, fmt.Errorf("invalid Order ID"))
			return
		}

		order, err = models.GetOrder(orderId)
		if err != nil {
			fmt.Printf("err order: %v \n", err)
			c.AbortWithError(400, err)
			return
		}

		// variantID := data.GetSessionValue(c, "variantId").(string)
		// dealID := data.GetSessionValue(c, "dealID").(string)
		variantID := order.VariantID
		dealID := order.DealID

		if len(dealID) > 0 {
			deal, err = models.GetDealByID(dealID)
			if err != nil {
				fmt.Printf("error: %s in GetDealByID", err.Error())
				hasActiveDeal = false
			} else {
				// deal exists on this product
				hasActiveDeal = models.DealExpired(deal)
			}
		} else {
			hasActiveDeal = false
		}

		fmt.Printf("variantID: %v\n", variantID)
		if variantID == "" {
			fmt.Println("invalid variant ID")
			c.AbortWithError(400, fmt.Errorf("invalid variant ID"))
			return
		}

		variant, err = models.GetVariant(variantID)
		if err != nil {
			fmt.Printf("err variant: %v \n", err)
			c.AbortWithError(400, err)
			return
		}

		productID := variant.ProductID

		product, err = models.GetProduct(productID)
		if err != nil {
			fmt.Printf("err product: %v \n", err)
			c.AbortWithError(400, err)
			return
		}
	} else if redirectStatus == "payment_retry" {
		orderId := c.Query("orderId")
		if orderId == "" {
			fmt.Println("invalid Order ID")
			c.AbortWithError(400, fmt.Errorf("invalid Order ID"))
			return
		}

		order, err = models.GetOrder(orderId)
		if err != nil {
			fmt.Printf("err order: %v \n", err)
			c.AbortWithError(400, err)
			return
		}

		variant, err = models.GetVariant(order.Variant.ID)
		if err != nil {
			fmt.Printf("err variant: %v \n", err)
			c.AbortWithError(400, err)
			return
		}

		if len(order.DealID) > 0 {
			deal, err = models.GetDealByID(order.DealID)
			if err != nil {
				fmt.Printf("err deal: %v \n", err)
				c.AbortWithError(400, err)
				return
			}
		}

		product, err = models.GetProduct(order.Product.ID)
		if err != nil {
			fmt.Printf("err product: %v \n", err)
			c.AbortWithError(400, err)
			return
		}
	}

	walletBalanceSufficient := order.OrderAmount.TotalAmount <= wallet.Balance

	brand, err := models.GetBrand(product.BrandID)
	if err != nil {
		fmt.Printf("err brand: %v \n", err)
		c.AbortWithError(400, err)
		return
	}

	paymentMethodsFromStripe, err := payments.ListPaymentMethodsByCustomer(userProfile.StripeCustomerID)
	if err != nil {
		fmt.Printf("err payment methods: %v \n", err)
		c.AbortWithError(400, err)
		return
	}

	type paymentMethodData struct {
		ID   string `json:"id"`
		Card struct {
			Last4 string `json:"last4"`
			Brand string `json:"brand"`
			Type  string `json:"type"`
		} `json:"card"`
		Type   string             `json:"type"`
		Amount models.OrderAmount `json:"amount"`
	}
	paymentMethods := []paymentMethodData{}

	savedCardAmount := order.OrderAmount
	savedCardAmount.AddPaymentMethodDiscountAndCalculateDiscount("saved_card", order.Cart.CartAmount.TotalAmount)

	for _, paymentMethod := range paymentMethodsFromStripe {
		var p paymentMethodData
		p.ID = paymentMethod.ID
		p.Type = string(paymentMethod.Type)
		if p.Type == "card" {
			p.Card.Last4 = paymentMethod.Card.Last4
			p.Card.Brand = string(paymentMethod.Card.Brand)
			p.Card.Type = string(paymentMethod.Card.Funding)
			p.Amount = savedCardAmount
		}
		paymentMethods = append(paymentMethods, p)
	}

	last_used_payment_method := userProfile.LastUsedPaymentMethod
	if last_used_payment_method == "" {
		last_used_payment_method = "COD"
	}

	if last_used_payment_method == "newCard" && len(paymentMethodsFromStripe) > 0 {
		last_used_payment_method = "stripe_card_" + paymentMethodsFromStripe[0].ID
	}

	address, err := getAddressObjectFromSession(c)
	if err != nil {
		fmt.Printf("err payment methods: %v \n", err)
		c.AbortWithError(400, err)
		return
	}

	// List of Payment Methods

	paymentOptions := models.GetPaymentOptions(order.OrderAmount, order.Cart.CartAmount.TotalAmount)

	availablePaymentOptions := models.GetEmptyPaymentOptions()

	for _, po := range paymentOptions {
		if product.PaymentMethods[po.ID].Available {
			availablePaymentOptions = append(availablePaymentOptions, po)
		}
	}
	c.Header("Cache-Control", "no-store, must-revalidate")

	if network.MobileRequest(c) {
		c.JSON(200, gin.H{
			"order":                    order,
			"product":                  product,
			"product_image":            product.Images[0],
			"brand":                    brand,
			"variant":                  variant,
			"payment_methods":          paymentMethods,
			"last_used_payment_method": last_used_payment_method,
			"address":                  address,
			"paymentOptions":           availablePaymentOptions,
			"cashOnDeliveryAmount":     int(order.OrderAmount.TotalAmount),
		})

		return
	}

	c.HTML(200, "select-payment", gin.H{
		"order":                    order,
		"product":                  product,
		"product_image":            product.Images[0],
		"brand":                    brand,
		"variant":                  variant,
		"payment_methods":          paymentMethods,
		"last_used_payment_method": last_used_payment_method,
		"hasActiveDeal":            hasActiveDeal,
		"deal":                     deal,
		"address":                  address,
		"internalUser":             user.Internal,
		"paymentOptions":           availablePaymentOptions,
		"cashOnDeliveryAmount":     int(order.OrderAmount.TotalAmount),
		"wallet":                   wallet,
		"walletBalanceSufficient":  walletBalanceSufficient,
	})
}

func SelectPaymentPostHandler(c *gin.Context) {

	c.Header("Cache-Control", "no-store, must-revalidate")

	// print request body
	c.Request.ParseForm()

	var pm struct {
		PaymentMethod string `form:"payment_method" json:"payment_method" binding:"required"`
		// Variant       string `form:"variant" json:"variant" binding:"required"`
		// Deal          string `form:"deal" json:"deal"`
		OrderID string `form:"order" json:"orderID"`
	}
	// bind form data to struct
	err := c.Bind(&pm)
	if err != nil {
		if c.Query("order") != "" && c.Query("payment_method") != "" {
			pm.OrderID = c.Query("order")
			pm.PaymentMethod = c.Query("payment_method")
		} else {
			data.AddSessionFlashMessage(c, "Invalid form data")
			// redirect to select payment page with get request
			c.Redirect(302, fmt.Sprintf("/select-payment?orderID=%s", c.Query("order")))
		}
	}

	user, err := getUserObjectFromSession(c)
	if err != nil {
		fmt.Println(err)
		c.AbortWithError(401, err)
		return
	}

	userProfile, err := user.GetProfile()
	if err != nil {
		fmt.Println("err getting user profile: ", err)
		c.AbortWithError(500, err)
		return
	}

	userProfile.LastUsedPaymentMethod = pm.PaymentMethod
	err = userProfile.Update()
	if err != nil {
		fmt.Println("err updating user profile: ", err)
		c.AbortWithError(500, err)
		return
	}

	order, err := models.GetOrder(pm.OrderID)
	if err != nil {
		fmt.Printf("err order: %v \n", err)
		c.AbortWithError(400, err)
		return
	}

	if order.PaymentStatus == "Paid" {
		fmt.Println("This Order has been processed")
		c.Redirect(302, "/")
	}

	if order.Coupon.ID != "" { // check if Coupon applied is valid
		couponValid := models.IsValidCoupon(order.Coupon.ID, order.UserID)
		if !couponValid {
			c.AbortWithError(400, fmt.Errorf("coupon Invalid"))
			return
		}
	}

	variant, err := models.GetVariant(order.VariantID)
	if err != nil {
		fmt.Printf("err variant: %v \n", err)
		c.AbortWithError(400, err)
		return
	}

	order.OrderAmount.AddPaymentMethodDiscountAndCalculateDiscount(pm.PaymentMethod, order.Cart.CartAmount.TotalAmount)

	payment := models.Payment{
		ID:       data.GetUUIDString("payment"),
		OrderID:  order.ID,
		Amount:   int64(order.OrderAmount.TotalAmount),
		Status:   "Initiated",
		Currency: "INR",
		Method:   pm.PaymentMethod,
	}

	payment, err = payment.Create()
	if err != nil {
		fmt.Printf("err payment: %v\n", err)
		c.AbortWithError(500, err)
		return
	}

	order.PaymentStatus = "Initiated"
	order.Payment = payment
	err = order.Update()
	if err != nil {
		fmt.Printf("err order: %v\n", err)
		c.AbortWithError(500, err)
		return
	}

	// set payment id in session
	data.SetSessionValue(c, "paymentId", payment.ID)

	// set order id in session
	data.SetSessionValue(c, "orderId", order.ID)

	if payment.Method == "COD" {
		payment.Status = "Paid"
		err = payment.Update()
		if err != nil {
			fmt.Printf("err payment: %v\n", err)
			c.AbortWithError(500, err)
			return
		}

		err = order.MarkOrderAsCompleted(payment)
		if err != nil {
			fmt.Printf("err order: %v\n", err)
			c.AbortWithError(500, err)
			return
		}

		err = order.MarkOrderAsCODUnconfirmed()
		if err != nil {
			fmt.Printf("err order: %v\n", err)
			c.AbortWithError(500, err)
			return
		}

		userProfile.LastUsedPaymentMethod = "COD"
		err = userProfile.Update()
		if err != nil {
			fmt.Printf("err user profile: %v\n", err)
			c.AbortWithError(500, err)
			return
		}

		if network.MobileRequest(c) {
			c.JSON(http.StatusOK, gin.H{
				"orderId": order.ID,
			})

			return
		}

		c.Redirect(http.StatusFound, "/order/success?product="+variant.ID+"&orderId="+order.ID)
	}

	if payment.Method == string(models.WALLET) {
		// Get Wallet
		wallet, err := models.GetUserWallet(user.ID)
		if err != nil {
			rw.JSONErrorResponse(c, 500, err)
			return
		}

		// Deduct Money From Wallet
		err = wallet.DeductAmount(float64(payment.Amount))
		if err != nil {
			rw.JSONErrorResponse(c, 400, err)
			return
		}

		// Mark Order as Completed
		payment.Status = "Paid"
		err = payment.Update()
		if err != nil {
			rw.JSONErrorResponse(c, 500, err)
			return
		}
		err = order.MarkOrderAsCompleted(payment)
		if err != nil {
			rw.JSONErrorResponse(c, 500, err)
			return
		}

		err = order.MarkOrderAsFulfillable()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		userProfile.LastUsedPaymentMethod = string(models.WALLET)
		err = userProfile.Update()
		if err != nil {
			fmt.Printf("err user profile: %v\n", err)
			c.AbortWithError(500, err)
			return
		}

		if network.MobileRequest(c) {
			c.JSON(http.StatusOK, gin.H{
				"orderId": order.ID,
			})

			return
		}

		c.Redirect(http.StatusFound, "/order/success?product="+variant.ID+"&orderId="+order.ID)
	}

	if payment.Method == string(models.UPI) {
		// Get UPI Payment Request
		response, thirdpartyResponse, err := payments.InitiateUPIPayment(payment.ID, float64(payment.Amount), "Roovo Payment")
		if err != nil {
			Sentry.SendErrorToSentry(c, err, nil)
			fmt.Printf("err UPI Payment: %v\n", err)
			c.AbortWithError(500, err)
			return
		}

		payment.ThirdPartyPaymentObject = thirdpartyResponse
		payment.ThirdPartyPaymentObjectId = response.ThirdPartyTransactionID

		err = payment.Update()
		if err != nil {
			fmt.Printf("err update payment: %v\n", err)
			c.AbortWithError(500, err)
			return
		}

		type UPIOptions struct {
			DeepLink    string
			IconPath    string
			OptionLabel string
		}

		upiOptions := []UPIOptions{
			{
				DeepLink:    response.UPIDeepLinks.PSPURI.GpayURI,
				IconPath:    "/static/icons/Gpay.png",
				OptionLabel: "Google Pay",
			},
			{
				DeepLink:    response.UPIDeepLinks.PSPURI.PaytmURI,
				IconPath:    "/static/icons/paytmIcon.png",
				OptionLabel: "Paytm",
			},
			{
				DeepLink:    response.UPIDeepLinks.PSPURI.PhonepeURI,
				IconPath:    "/static/icons/Phonepe.png",
				OptionLabel: "PhonePe",
			},
			{
				DeepLink:    response.UPIDeepLinks.GeneratedLink,
				IconPath:    "/static/icons/rupayIcon.png",
				OptionLabel: "Other UPI Apps",
			},
		}

		c.Header("Cache-Control", "no-store, must-revalidate")

		//Move to new Page with Payment Link
		c.HTML(http.StatusOK, "UPI-Payment-Page", gin.H{
			"paymentID":     payment.ID,
			"upiOptions":    upiOptions,
			"UPIDeepLink":   response.UPIDeepLinks,
			"orderID":       payment.OrderID,
			"payableAmount": payment.Amount,
		})

		return
	}

	if payment.Method == "newCard" {
		intent, err := payments.CreateIntent(payment, userProfile.StripeCustomerID)
		if err != nil {
			fmt.Printf("err intent: %v\n", err)
			c.AbortWithError(500, err)
			return
		}
		payment.ThirdPartyPaymentObject = intent
		payment.ThirdPartyPaymentObjectId = intent.ID
		err = payment.Update()
		if err != nil {
			fmt.Printf("err update payment: %v\n", err)
			c.AbortWithError(500, err)
			return
		}

		if network.MobileRequest(c) {
			c.JSON(http.StatusOK, gin.H{
				"ClientSecret": intent.ClientSecret,
				"OrderId":      order.ID,
			})

			return
		}

		c.HTML(http.StatusOK, "createIntent.html", gin.H{
			"ClientSecret": intent.ClientSecret,
			"OrderId":      order.ID,
			"Amount":       payment.Amount,
			"StripePK":     os.Getenv("STRIPE_PUBLISHABLE_KEY"),
			"RedirectURL":  os.Getenv("BASE_URL") + "/payments/redirect?orderId=" + order.ID + "&paymentId=" + payment.ID,
		})
	}

	// chekc if payment method starts with stripe
	if strings.HasPrefix(payment.Method, "stripe_card_") {
		stripeCardId := strings.Split(payment.Method, "stripe_card_")[1]
		intent, err := payments.CreatePaymentIntentWithPaymentMethodId(payment, userProfile.StripeCustomerID, stripeCardId)
		if err != nil {
			fmt.Printf("err intent: %v\n", err)
			c.AbortWithError(500, err)
			return
		}
		payment.ThirdPartyPaymentObject = intent
		payment.ThirdPartyPaymentObjectId = intent.ID
		err = payment.Update()
		if err != nil {
			fmt.Printf("err update payment: %v\n", err)
			c.AbortWithError(500, err)
			return
		}

		userProfile.LastUsedPaymentMethod = "stripe_card_" + stripeCardId

		err = userProfile.Update()
		if err != nil {
			fmt.Printf("err update user profile: %v\n", err)
			c.AbortWithError(500, err)
			return
		}

		if network.MobileRequest(c) {
			c.JSON(http.StatusOK, gin.H{
				"ClientSecret": intent.ClientSecret,
				"OrderId":      order.ID,
			})

			return
		}

		c.HTML(http.StatusOK, "createIntent2.html", gin.H{
			"ClientSecret": intent.ClientSecret,
			"OrderId":      order.ID,
			"Amount":       payment.Amount,
			"StripeCardId": stripeCardId,
			"StripePK":     os.Getenv("STRIPE_PUBLISHABLE_KEY"),
			"RedirectURL":  os.Getenv("BASE_URL") + "/payments/redirect?orderId=" + order.ID + "&paymentId=" + payment.ID,
		})
	}
}

func OrderList(c *gin.Context) {
	// get all orders for User

	user, err := getUserObjectFromSession(c)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	orders, err := models.GetAllOrders(user.ID)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	if network.MobileRequest(c) {
		c.JSON(http.StatusOK, gin.H{
			"orders": orders,
		})
		return
	}

	c.HTML(200, "ordersList", gin.H{
		"orders": orders,
	})
}

func OrderSpecificPage(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty order ID"})
		return
	}
	user, err := getUserObjectFromSession(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid User"})
		return
	}
	order, err := models.GetOrderByUserAndID(user.ID, orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to get order by id " + err.Error()})
		return
	}
	shipment, err := models.GetShipmentByID(orderID)
	if err != nil && err != mongo.ErrNoDocuments {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get shipment for order " + err.Error()})
		return
	}
	tracking, err := shiprocket.GetTrackingForAWB(shipment.AWB)
	if err != nil && err != mongo.ErrNoDocuments {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get tracking for awb " + err.Error()})
		return
	}

	finalTrackingData, err := shiprocket.MakeTrackingData(tracking, order.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to make tracking data for awb " + err.Error()})
		return
	}

	recentlyViewedProducts, err := models.GetRecentlyViewedProductsFromRedis(c, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get recently viewed products " + err.Error()})
		return
	}

	var Paginater Pagination = Pagination{
		Limit: 20,
		Page:  1,
	}

	processedAppliedFilterMap := map[string]search.FilterObject{}
	processedAppliedFilterMap["category"] = search.FilterObject{
		Values:   []string{order.Product.Category},
		Operator: "=",
		Path:     "category",
	}
	processedAppliedFilterMap["product"] = search.FilterObject{
		Values:   []string{order.Product.ID},
		Operator: "!=",
		Path:     "id",
	}
	sortArr := []search.SortObject{}
	similarProducts, err := ProductPaginate2(&Paginater, "", processedAppliedFilterMap, sortArr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get similar products " + err.Error()})
		return
	}
	brand, err := models.GetBrand(order.Product.BrandID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get brand " + err.Error()})
		return
	}
	tempTime, _ := time.Parse("02 01 2006 15:04:05", tracking.CurrentTimestamp)
	latestTimestamp := tempTime.Format("Monday, Jan 02")
	etd, err := time.Parse("2006-01-02 15:04:05", tracking.Etd)
	etdText := ""
	if err == nil {
		etdText = etd.Format("Monday, Jan 02")
	}

	cancellable := !finalTrackingData[3].Completed

	c.HTML(200, "orderSpecific", gin.H{
		"order":                  order,
		"shipment":               shipment,
		"recentlyViewedProducts": recentlyViewedProducts,
		"similarProducts":        similarProducts,
		"similarPaginater":       Paginater,
		"finalTrackingData":      finalTrackingData,
		"brand":                  brand,
		"tracking":               tracking,
		"estimatedTime":          etdText,
		"latestTimestamp":        latestTimestamp,
		"cancellable":            cancellable,
	})
}

func OrderConfirmedHandler(c *gin.Context) {
	c.HTML(200, "order-confirmed", gin.H{})
}

func OrderCancellationPage(c *gin.Context) {
	orderID, _ := c.Params.Get("orderID")
	order, err := models.GetOrder(orderID)
	if err != nil {
		c.AbortWithError(404, err)
		return
	}
	c.HTML(200, "cancelOrder", gin.H{
		"order": order,
	})
}

func OrderCancellation(c *gin.Context) {
	orderID, _ := c.Params.Get("orderID")
	order, err := models.GetOrder(orderID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Unable to find order with given id" + err.Error()})
		return
	}
	var cancellation Cancellation
	err = c.BindJSON(&cancellation)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body " + err.Error()})
		return
	}
	order.FulfillmentStatus = "Cancelled"
	order.CancellationReason = cancellation.Reason
	err = order.Update()
	if err != nil {
		c.JSON(400, gin.H{"error": "Could not update order " + err.Error()})
		return
	}
	shipment, err := models.GetShipmentByID(order.ID)
	if err == nil {
		err := shiprocket.CancelShiprocketOrder(fmt.Sprint(shipment.OrderId))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Couldn't cancel shiprocket order: " + err.Error()})
			return
		}
	} else if err != mongo.ErrNoDocuments {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Couldn't get shipment: " + err.Error()})
		return
	}
	fmt.Println("ORDER CANCELLED")
	c.HTML(200, "cancelledOrder", gin.H{
		"order": order,
	})
}

func TrackOrder(c *gin.Context) {
	orderID := c.Query("orderID")
	shipment, err := models.GetShipmentByID(orderID)
	if err != nil && err != mongo.ErrNoDocuments {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get shipment for order " + err.Error()})
		return
	}
	order, err := models.GetOrder(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get order for orderID " + err.Error()})
		return
	}
	tracking, err := shiprocket.GetTrackingForAWB(shipment.AWB)
	if err != nil && err != mongo.ErrNoDocuments {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get tracking for awb " + err.Error()})
		return
	}
	finalTrackingData, err := shiprocket.MakeTrackingData(tracking, order.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to make tracking data for awb " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": finalTrackingData, "trackingID": shipment.AWB})
}
