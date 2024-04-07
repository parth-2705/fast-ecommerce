package controllers

// func GetShiprocketOrders(c *gin.Context) {
// 	var shiprocketOrders []map[string]interface{}
// 	seller, err := auth.GetSellerFromSession(c)
// 	if err != nil {
// 		c.AbortWithError(http.StatusBadGateway, err)
// 		return
// 	}
// 	if !seller.ProfileCompleted {
// 		c.Redirect(http.StatusFound, "/info")
// 		return
// 	}

// 	preSkip := utils.GetPaymentPreSkipBySellerID(seller.ID, true, []string{})

// 	postSkip := utils.PaymentsPostSkip

// 	additionalFilter := bson.M{}
// 	startDate := c.Query("startDate")
// 	endDate := c.Query("endDate")
// 	startDateTime := time.Now()
// 	endDateTime := time.Now()
// 	if startDate != "" {
// 		startDateTime, _ = time.Parse("2006-01-02", startDate)
// 	}
// 	if endDate != "" {
// 		endDateTime, _ = time.Parse("2006-01-02", endDate)
// 		endDateTime = endDateTime.Add(24 * time.Hour)
// 	}

// 	statusFilter := c.Query("status")
// 	if statusFilter != "" {
// 		additionalFilter["orders.shipmentStatus"] = statusFilter
// 	}

// 	if len(additionalFilter) > 0 {
// 		fmt.Println("additionalFilter:", additionalFilter)
// 		preSkip = append(preSkip, bson.D{{Key: "$match", Value: additionalFilter}})
// 	}

// 	limitInt := controllers.GetLimitFromQueryValue(c.Query("limit"))
// 	pageInt := controllers.GetPageFromQueryValue(c.Query("page"))

// 	var Paginater controllers.Pagination = controllers.Pagination{
// 		Limit: limitInt,
// 		Page:  pageInt,
// 	}
// 	costExtra := utils.GetShipmentsWithCostQuery(startDateTime, endDateTime)
// 	postSkip = append(postSkip, costExtra...)
// 	orders, err := controllers.Paginate("shipping", &Paginater, shiprocketOrders, preSkip, postSkip)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error": err.Error(),
// 		})
// 		return
// 	}
// 	c.JSON(http.StatusOK, gin.H{"orders": orders})

// }

// func GetAllShiprocketOrders(c *gin.Context) {
// 	var shiprocketOrders []map[string]interface{}
// 	seller, err := auth.GetSellerFromSession(c)
// 	if err != nil {
// 		c.AbortWithError(http.StatusBadGateway, err)
// 		return
// 	}
// 	if !seller.ProfileCompleted {
// 		c.Redirect(http.StatusFound, "/info")
// 		return
// 	}

// 	preSkip := utils.GetPaymentPreSkipBySellerID(seller.ID, true, []string{})

// 	postSkip := utils.PaymentsPostSkip

// 	additionalFilter := bson.M{}
// 	startDate := c.Query("startDate")
// 	endDate := c.Query("endDate")
// 	startDateTime := time.Now()
// 	endDateTime := time.Now()
// 	loc, _ := time.LoadLocation("Asia/Calcutta")
// 	if startDate != "" {

// 		startDateTime, _ = time.ParseInLocation("2006-01-02", startDate, loc)
// 	}
// 	if endDate != "" {
// 		endDateTime, _ = time.ParseInLocation("2006-01-02", endDate, loc)
// 		endDateTime = endDateTime.Add(24 * time.Hour)
// 	}

// 	statusFilter := c.Query("status")
// 	if statusFilter != "" {
// 		additionalFilter["orders.shipmentStatus"] = statusFilter
// 	}

// 	if len(additionalFilter) > 0 {
// 		fmt.Println("additionalFilter:", additionalFilter)
// 		preSkip = append(preSkip, bson.D{{Key: "$match", Value: additionalFilter}})
// 	}

// 	chargesArr := append(preSkip, postSkip...)
// 	chargesExtra := utils.GetChargesExtraByCheckTime(startDateTime, endDateTime)
// 	chargesArr = append(chargesArr, chargesExtra...)
// 	var charges []map[string]interface{}
// 	var chargesItem interface{}
// 	chargesCur, err := db.ShippingCollection.Aggregate(context.Background(), chargesArr)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get charges " + err.Error()})
// 		return
// 	}
// 	err = chargesCur.All(context.Background(), &charges)

// 	if len(charges) > 0 {
// 		chargesItem = charges[0]
// 	} else {
// 		chargesItem = []interface{}{""}
// 	}
// 	fmt.Println("This is chargesItem:", chargesItem, charges)

// 	limitInt := controllers.GetLimitFromQueryValue(c.Query("limit"))
// 	pageInt := controllers.GetPageFromQueryValue(c.Query("page"))

// 	var Paginater controllers.Pagination = controllers.Pagination{
// 		Limit: limitInt,
// 		Page:  pageInt,
// 	}
// 	costExtra := utils.GetShipmentsWithCostQuery(startDateTime, endDateTime)
// 	postSkip = append(postSkip, costExtra...)
// 	orders, err := controllers.Paginate("shipping", &Paginater, shiprocketOrders, preSkip, postSkip)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error": err.Error(),
// 		})
// 		return
// 	}

// 	var statusList interface{}
// 	var listResp []map[string]interface{}
// 	statusFilterArr := bson.A{
// 		bson.D{
// 			{"$match",
// 				bson.D{
// 					{"product.sellerID", seller.ID},
// 				},
// 			},
// 		},
// 		bson.D{
// 			{"$group",
// 				bson.D{
// 					{"_id", ""},
// 					{"statusList", bson.D{{"$addToSet", "$shipmentStatus"}}},
// 				},
// 			},
// 		},
// 	}
// 	cur, err := db.OrderCollection.Aggregate(context.Background(), statusFilterArr)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to get status list " + err.Error()})
// 		return
// 	}
// 	err = cur.All(context.Background(), &listResp)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to decode status list " + err.Error()})
// 		return
// 	}
// 	if len(listResp) > 0 {
// 		statusList = listResp[0]["statusList"]
// 	} else {
// 		statusList = []interface{}{""}
// 	}
// 	c.JSON(http.StatusOK, gin.H{"orders": orders, "statusList": statusList, "charges": chargesItem})
// }
