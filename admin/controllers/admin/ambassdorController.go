package controllers

import (
	"hermes/models"
	"hermes/utils/rw"
	"strings"

	"github.com/gin-gonic/gin"
)

func RecordProductViewForAmbassdor(c *gin.Context) {

	type requstStruct struct {
		ReferralCode        string `json:"code"`
		ReferredPhoneNumber string `json:"phoneNumber"`
		ReferredProduct     string `json:"productID"`
	}

	var req requstStruct
	err := c.BindJSON(&req)
	if err != nil {
		rw.JSONErrorResponse(c, 400, err)
		return
	}

	ambassdor, err := models.GetAmbassdorByReferralCode(req.ReferralCode)
	if err != nil {
		rw.JSONErrorResponse(c, 400, err)
		return
	}

	var phone string = req.ReferredPhoneNumber
	if !strings.HasPrefix(req.ReferredPhoneNumber, "+") {
		phone = "+" + req.ReferredPhoneNumber
	}

	_, err = models.GetProduct(req.ReferredProduct)
	if err != nil {
		rw.JSONErrorResponse(c, 400, err)
		return
	}

	referredUser, err := models.GetUser(phone)
	if err != nil {
		rw.JSONErrorResponse(c, 400, err)
		return
	}

	refferalRecord, err := ambassdor.AddReferralRecord(referredUser.ID, req.ReferredProduct)
	if err != nil {
		rw.JSONErrorResponse(c, 500, err)
		return
	}

	c.JSON(200, gin.H{"recordID": refferalRecord.ID})
}
