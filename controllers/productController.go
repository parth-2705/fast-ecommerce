package controllers

import (
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"hermes/configs"
	"hermes/db"
	"hermes/models"
	"hermes/search"
	"hermes/utils/amplitude"
	"hermes/utils/data"
	"hermes/utils/network"
)

func getVariantIDFromURL(c *gin.Context) (variantID string) {
	variantID = c.Param("variantID")
	return
}

func getCartIDFromURL(c *gin.Context) (cartID string) {
	cartID = c.Param("cartID")
	return
}

func getSimilarProducts(product models.Product) (similarProducts []models.Product, err error) {
	filter1 := bson.D{{Key: "$match", Value: bson.M{"category": product.Category, "_id": bson.M{"$not": bson.M{"$eq": product.ID}}}}}
	filter2 := bson.D{{Key: "$limit", Value: 8}}
	cur, err := getProductsWithBrand(filter1, filter2)
	if err != nil {
		return nil, err
	}
	err = cur.All(context.Background(), &similarProducts)
	return
}

// Function that joins Product Collection with Brands Collection to get Products with Brand Objects. Take s input a list of bson.D filters, the result of which should be used to join
func getProductsWithBrand(filters ...bson.D) (*mongo.Cursor, error) {
	aggregrateSearchObject := bson.A{}
	for _, filter := range filters {
		aggregrateSearchObject = append(aggregrateSearchObject, filter)
	}

	aggregrateSearchObject = append(aggregrateSearchObject,
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "brands"},
					{Key: "localField", Value: "brandID"},
					{Key: "foreignField", Value: "_id"},
					{Key: "as", Value: "brand"},
				},
			},
		},
		bson.D{
			{Key: "$unwind",
				Value: bson.D{
					{Key: "path", Value: "$brand"},
					{Key: "preserveNullAndEmptyArrays", Value: true},
				},
			},
		},
		bson.D{
			{Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "variants"},
					{Key: "localField", Value: "_id"},
					{Key: "foreignField", Value: "productID"},
					{Key: "as", Value: "variants"},
				},
			},
		},
	)

	cursor, err := db.ProductCollection.Aggregate(context.Background(), aggregrateSearchObject)
	return cursor, err
}

// func getCompleteProduct(productID string) (product models.Product, err error) {
// 	cursor, err := db.ProductCollection.Aggregate(context.Background(),
// 		bson.A{
// 			bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: productID}}}},
// 			bson.D{
// 				{Key: "$lookup",
// 					Value: bson.D{
// 						{Key: "from", Value: "brands"},
// 						{Key: "localField", Value: "brandID"},
// 						{Key: "foreignField", Value: "_id"},
// 						{Key: "as", Value: "brand"},
// 					},
// 				},
// 			},
// 			bson.D{
// 				{Key: "$unwind",
// 					Value: bson.D{
// 						{Key: "path", Value: "$brand"},
// 						{Key: "preserveNullAndEmptyArrays", Value: true},
// 					},
// 				},
// 			},
// 			bson.D{
// 				{Key: "$lookup",
// 					Value: bson.D{
// 						{Key: "from", Value: "variants"},
// 						{Key: "localField", Value: "_id"},
// 						{Key: "foreignField", Value: "productID"},
// 						{Key: "as", Value: "variants"},
// 					},
// 				},
// 			},
// 		},
// 	)

// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	cursor.Next(context.Background())
// 	err = cursor.Decode(&product)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	cursor.Close(context.Background())
// 	return
// }

func mapSubset(set map[string]string, subset map[string]string) bool {
	for key, val := range subset {
		if set[key] != val {
			return false
		}
	}

	return true
}

func GetProductPage(c *gin.Context) {
	var products []models.Product
	var err error
	limit := c.Query("limit")
	limitInt := 20
	if limit != "" {
		limitInt, err = strconv.Atoi(limit)
	}
	limitInt = int(math.Min(float64(limitInt), 20))
	page := c.Query("page")
	pageInt := 1
	if page != "" {
		pageInt, err = strconv.Atoi(page)
	}

	var Paginater Pagination = Pagination{
		Limit: limitInt,
		Page:  pageInt,
	}

	//apply filters logic
	appliedfiltersQuery := c.Query("filter")
	appliedfiltersArr := []string{}
	if appliedfiltersQuery != "" {
		appliedfiltersArr = strings.Split(appliedfiltersQuery, ",")
	}

	appliedFilters := map[string]bool{}
	processedAppliedFilterMap := map[string]search.FilterObject{}
	for _, val := range appliedfiltersArr {
		appliedFilters[val] = true
		filterItem := strings.Split(val, `/\`)
		temp := append(processedAppliedFilterMap[filterItem[0]].Values, filterItem[1])
		processedAppliedFilterMap[filterItem[0]] = search.FilterObject{
			Values:   temp,
			Operator: "=",
		}
	}
	sortArr := []search.SortObject{}
	categoryID := c.Query("categoryID")
	productID := c.Query("productID")
	brandID := c.Query("brandID")
	if brandID != "" {
		processedAppliedFilterMap["brandID"] = search.FilterObject{
			Values:   []string{brandID},
			Operator: "=",
			Path:     "brandID",
		}
		if productID != "" {
			processedAppliedFilterMap["product"] = search.FilterObject{
				Values:   []string{productID},
				Operator: "!=",
				Path:     "id",
			}
		}
		products, err = ProductPaginate2(&Paginater, "", processedAppliedFilterMap, sortArr)
	} else if productID != "" {
		if categoryID != "" {
			categoryIDArr, err := GetChildrenCategoryIDArr(categoryID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Error fetching children categories " + err.Error()})
				return
			}
			categoryFilter := []string{categoryID}
			categoryFilter = append(categoryFilter, categoryIDArr...)
			processedAppliedFilterMap["category"] = search.FilterObject{
				Values:   categoryFilter,
				Operator: "=",
				Path:     "newCategory",
			}
			processedAppliedFilterMap["product"] = search.FilterObject{
				Values:   []string{productID},
				Operator: "!=",
				Path:     "id",
			}
			products, err = ProductPaginate2(&Paginater, "", processedAppliedFilterMap, sortArr)
		}
	} else if categoryID != "" {
		categoryIDArr, err := GetChildrenCategoryIDArr(categoryID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Error fetching children categories " + err.Error()})
			return
		}
		categoryFilter := []string{categoryID}
		categoryFilter = append(categoryFilter, categoryIDArr...)
		processedAppliedFilterMap["category"] = search.FilterObject{
			Values:   categoryFilter,
			Operator: "=",
			Path:     "newCategory",
		}
		products, err = ProductPaginate2(&Paginater, "", processedAppliedFilterMap, sortArr)
	} else {
		// categoryFilter := []string{"category-bdbff66e-11f2-4331-8c35-ca368dca08cd"}
		// processedAppliedFilterMap["category"] = search.FilterObject{
		// 	Values:   categoryFilter,
		// 	Operator: "=",
		// 	Path:     "category",
		// }
		products, err = ProductPaginate2(&Paginater, "", processedAppliedFilterMap, sortArr)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get products " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"products": products, "pagination": Paginater})
	return
}

// Define a route for serving a single product
func SingleProductPage(c *gin.Context) {

	// Get the product ID from the URL parameter
	productID := c.Param("id")

	// Get referral code from param if any and set it to the session
	referral := c.Request.URL.Query().Get("code")
	if len(referral) > 0 {
		fmt.Println("referral code set in session: ", referral)
		data.SetSessionValue(c, configs.Referral, referral)
	}

	// Amplitude tracking
	go amplitude.TrackEventWithPropertiesByAuth("Product Page", map[string]interface{}{"id": productID}, c)
	trackingMap := amplitude.GetTrackingMap(c)

	// Find the product in the database
	var product models.Product
	var attributeNameArr []string
	var finalMap map[string]interface{} = make(map[string]interface{})

	product, err := models.GetCompleteProduct(productID)
	if err != nil {
		fmt.Println(err)
		c.AbortWithError(404, err)
		return
	}

	data.SetUTMParamsInSession(c)

	attributeMap := make(map[string]string)
	for _, attribute := range product.Attributes {
		attributeNameArr = append(attributeNameArr, attribute.Name)
		val := c.Query(attribute.Name)
		if val == "" {
			continue
		}
		for _, option := range attribute.Options {
			if val == option {
				attributeMap[attribute.Name] = val
			}
		}
	}

	variants, _ := models.GetVariantsByProductID(product.ID)

	variantToShow := variants[0]
	lowestPrice := variants[0].Price.SellingPrice
	for _, variant := range variants {
		var keyMaker []string
		for _, attributeName := range attributeNameArr {
			keyMaker = append(keyMaker, variant.Attributes[attributeName])
		}

		finalMap[strings.Join(keyMaker, "-")] = variant

		if mapSubset(attributeMap, variant.Attributes) {
			variantToShow = variant
			break
		}

		if variant.Price.SellingPrice < lowestPrice {
			lowestPrice = variant.Price.SellingPrice
			variantToShow = variant
		}
	}

	attributeMap = variantToShow.Attributes

	back := c.Request.Referer()
	if back == "" {
		back = "/"
	}

	user, _ := getUserObjectFromSession(c)
	mobileNumber, _ := GetMobileNumberFromSession(c)
	UAID, _ := getUserAgentIDFromSession(c)

	// View History
	recentlyViewedProducts, err := models.GetRecentlyViewedProductsFromRedis(c, productID)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	// fmt.Printf("len(recentlyViewedProducts): %v\n", len(recentlyViewedProducts))

	contextWithTimeout, _ := context.WithTimeout(context.Background(), 30*time.Second)
	go models.CreateProductPageViewEntry(contextWithTimeout, mobileNumber.GetCompleteMobileNumber(), user.ID, productID, UAID)
	go models.AddToRecentlyViewedList(contextWithTimeout, c, productID)

	productRatingVisualizer, err := models.GetRatingVisualizationObject(product.AverageRating)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	// Get Reviews for this product
	reviews, err := product.GetReviews(5)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	var discounted bool = true // change this to false
	if product.Price.Discount == 0 {
		discounted = false
	}

	// similarProducts, err := getSimilarProducts(product)
	limit := c.Query("limit")
	limitInt := 20
	if limit != "" {
		limitInt, err = strconv.Atoi(limit)
	}
	limitInt = int(math.Min(float64(limitInt), 20))
	page := c.Query("page")
	pageInt := 1
	if page != "" {
		pageInt, err = strconv.Atoi(page)
	}

	var Paginater Pagination = Pagination{
		Limit: limitInt,
		Page:  pageInt,
	}

	var BrandPaginater Pagination = Pagination{
		Limit: limitInt,
		Page:  pageInt,
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

	processedAppliedFilterMap2 := map[string]search.FilterObject{}
	processedAppliedFilterMap2["brandID"] = search.FilterObject{
		Values:   []string{product.BrandID},
		Operator: "=",
		Path:     "brandID",
	}
	processedAppliedFilterMap2["product"] = search.FilterObject{
		Values:   []string{product.ID},
		Operator: "!=",
		Path:     "id",
	}
	brandProducts, err := ProductPaginate2(&BrandPaginater, "", processedAppliedFilterMap2, sortArr)
	if err != nil {
		fmt.Println(err)
		c.AbortWithError(500, err)
		return
	}

	productIDWishlisted := data.GetSessionValue(c, "setProductToSession")
	if user.ID != "" && productIDWishlisted != nil && productID == productIDWishlisted.(string) {
		err = addToWishlist(c, user, productID)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		data.SetSessionValue(c, "setProductToSession", nil)
	}

	productIDNotify := data.GetSessionValue(c, "setNotifyToSession")
	if user.ID != "" && productIDNotify != nil && productID == productIDNotify.(string) {
		_, err := addOrRemoveNotifyObjectFromDB(user.ID, productID)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		data.SetSessionValue(c, "setNotifyToSession", nil)
	}

	// wishlisted := models.GetIfItemIsInWishlistOfUser(user, productID)
	_, canAddReview := models.CanAddReview(product.ID, user)
	notify := models.GetIfItemIsInNotifylistOfUser(user.ID, productID)

	hasActiveDeal := true
	deal, err := models.GetDealByProductID(product.ID)
	if err != nil {
		hasActiveDeal = false
	} else {
		// deal exists on this product
		hasActiveDeal = models.DealExpired(deal)
	}

	seller, err := models.GetSellerByID(product.SellerID)
	if err != nil {
		seller = models.Seller{}
	}

	outOfStock := false
	if variantToShow.Quantity == 0 {
		outOfStock = true
	}
	disableWhatsappButton := outOfStock
	str, _ := base64.StdEncoding.DecodeString(product.DescriptionEncoded)
	product.Description = template.HTML(str)

	// timer := map[string]int{
	// 	"hour": gofakeit.IntRange(3, 9),
	// 	"min":  gofakeit.IntRange(0, 59),
	// 	"sec":  gofakeit.IntRange(0, 59),
	// }

	newUser := false
	// if user.ID == ""{
	// 	newUser = true
	// }

	// fb.SendViewContentEvent(c, product.ID)

	var referralCode string

	if user.HasJoinedReferralProgram {
		profile, err := user.GetProfile()
		if err != nil {
			fmt.Println("profile err: ", err)
		}
		referralCode = profile.ReferralCode
	}

	fmt.Println("referralCode:", referralCode)

	session := sessions.Default(c)
	referralCodeLink := session.Get(configs.Referral)
	var referredByUserCode string

	if referralCodeLink == nil {
		referredByUserCode = ""
	} else {
		code, ok := referralCodeLink.(string)
		if ok && len(code) > 0 {
			referredByUserCode = code
		} else {
			referredByUserCode = ""
		}
	}

	fmt.Printf("product.Media: %v\n", product.Media)

	if network.MobileRequest(c) {
		c.JSON(http.StatusOK, gin.H{
			"product":                    product,
			"productRatingVisualisation": productRatingVisualizer,
			"reviews":                    reviews,
			"back":                       back,
			"isDiscounted":               discounted,
			"variantToShow":              variantToShow,
			"mapmap":                     attributeMap,
			"variantMap":                 finalMap,
			"canAddReview":               canAddReview,
			"attributeNameArr":           attributeNameArr,
			"similarProducts":            similarProducts,
			"brandProducts":              brandProducts,
			"similarPaginater":           Paginater,
			"brandPaginater":             BrandPaginater,
			// "wishlisted":                 wishlisted,
			"trackingMap":            trackingMap,
			"deal":                   deal,
			"hasActiveDeal":          hasActiveDeal,
			"seller":                 seller,
			"outOfStock":             outOfStock,
			"disableWhatsappButton":  disableWhatsappButton,
			"notify":                 notify,
			"recentlyViewedProducts": recentlyViewedProducts,
			"referralCode":           referralCode,
			"referredByUserCode":     referredByUserCode,
		})

		return
	}

	// Render the product page
	c.HTML(http.StatusOK, "productPage", gin.H{
		"product":                    product,
		"productRatingVisualisation": productRatingVisualizer,
		"reviews":                    reviews,
		"back":                       back,
		"canAddReview":               canAddReview,
		"isDiscounted":               discounted,
		"variantToShow":              variantToShow,
		"mapmap":                     attributeMap,
		"variantMap":                 finalMap,
		"attributeNameArr":           attributeNameArr,
		"similarProducts":            similarProducts,
		"brandProducts":              brandProducts,
		"similarPaginater":           Paginater,
		"brandPaginater":             BrandPaginater,
		// "wishlisted":                 wishlisted,
		"trackingMap":            trackingMap,
		"deal":                   deal,
		"hasActiveDeal":          hasActiveDeal,
		"seller":                 seller,
		"outOfStock":             outOfStock,
		"disableWhatsappButton":  disableWhatsappButton,
		"notify":                 notify,
		"recentlyViewedProducts": recentlyViewedProducts,
		// "timer":                      timer,
		"referredByUserCode": referredByUserCode,
		"couponExpired":      !newUser,
		"referralCode":       referralCode,
	})

}
