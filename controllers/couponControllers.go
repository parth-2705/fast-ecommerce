package controllers

import (
	"hermes/models"
	"hermes/utils/data"
	"strings"

	"github.com/gin-gonic/gin"
)

func CouponGetHandler(c *gin.Context) {
	// cartID:= data.GetCartFromSession(c)
	// cart,_:=models.GetCart(cartID)
	// coupon,_:=models.GetCouponByID(cart.CouponID)
	c.HTML(200, "coupon", gin.H{
		// "coupon": coupon,
	})

}

type CouponCodeCheckRequest struct {
	CouponCode string `json:"couponCode"`
}

func CheckCouponApplicability(c *gin.Context) {

	// Read Request
	var couponCodeCheckRequest CouponCodeCheckRequest
	err := c.BindJSON(&couponCodeCheckRequest)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	// Check Applicability
	applicability, reason, err := isCouponApplyable(c, couponCodeCheckRequest)
	if err != nil {
		c.AbortWithError(500, err)
	}

	// Write Response
	c.JSON(200, map[string]interface{}{
		"applicability": applicability,
		"reason":        reason,
	})

}

func isCouponApplyable(c *gin.Context, couponCode CouponCodeCheckRequest) (applicable bool, reason string, err error) {

	couponCode.CouponCode = strings.ToUpper(couponCode.CouponCode)

	// Check if Coupon Exists
	coupon, err := models.GetCouponByCode(couponCode.CouponCode)
	if err != nil {
		return false, "Coupon Not Found", nil
	}

	cartID := data.GetCartFromSession(c)
	cart, err := models.GetCart(cartID)
	if err != nil {
		return false, "Error Getting Cart", err
	}

	// Check if Coupon is not Expired
	valid, reason := coupon.IsValid(cart.UserID)
	if !valid {
		return false, reason, nil
	}

	// Check if Coupon is applicable on this order
	// applicable, reason, err = coupon.IsCouponApplicable(cart.ProductID)
	// if err != nil {
	// 	return false, "Error Checking Coupon", err
	// }

	// if applicable {
	// 	// data.SetCouponInSession(c, couponCode.CouponCode)
	// 	cart.ApplyCoupon(coupon)
	// }

	applicable, reason, err = cart.ApplyCoupon(coupon)

	return
}

func RemoveAppliedCoupon(c *gin.Context) {
	cart, err := getCartFromSession(c)
	if err != nil {
		c.AbortWithError(400, err)
	}
	err = cart.RemoveCoupon()
	if err != nil {
		c.JSON(400, map[string]string{
			"error": err.Error(),
		})
		return
	}
	c.JSON(204, nil)
}
