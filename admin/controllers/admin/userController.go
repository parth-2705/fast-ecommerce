package controllers

import (
	"context"
	"encoding/csv"
	"fmt"
	"hermes/controllers"
	"hermes/db"
	"hermes/models"
	"hermes/utils/data"
	"hermes/utils/rw"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserUpdateRequest struct {
	User           models.User    `json:"user"`
	DefaultAddress models.Address `json:"defaultAddress"`
}

func GetUsers(c *gin.Context) {
	var err error
	var user models.User
	phone := c.Query("phone")
	if phone == "" {
		GetUserPage(c)
		return
	} else {
		user, err = models.GetUser(phone)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get user by phone number " + err.Error()})
			return
		}
		address := user.GetDefaultAddress()
		c.JSON(http.StatusOK, gin.H{"users": gin.H{"rows": []models.User{user}}, "defaultAddress": address})
		return
	}
}

func GetUserPage(c *gin.Context) {
	var err error
	var users []models.User
	limit := c.Query("limit")
	limitInt := 20
	if limit != "" {
		limitInt, err = strconv.Atoi(limit)
		if err != nil {
			limitInt = 20
		}
	}

	page := c.Query("page")
	pageInt := 1
	if page != "" {
		pageInt, err = strconv.Atoi(page)
		if err != nil {
			pageInt = 1
		}
	}

	var Paginater controllers.Pagination = controllers.Pagination{
		Limit: limitInt,
		Page:  pageInt,
	}

	var userPagination *controllers.Pagination
	phone := c.Query("phone")
	if phone != "" {
		userPagination, err = controllers.Paginate("user", &Paginater, users, bson.A{bson.D{{Key: "phone", Value: phone}}}, bson.A{})
	} else {
		userPagination, err = controllers.Paginate("user", &Paginater, users, bson.A{}, bson.A{})
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error in pagination": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": userPagination})
}

func GetAllUserAddressess(c *gin.Context) {
	userID := c.Query("userID")

	var user models.User
	user.ID = userID

	addressess, err := user.GetAddresses()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error in getting addressess ": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"addressess": addressess})
}

func GetDataByPincode(ctx *gin.Context) {
	// Get the pincode from the URL
	pincode := ctx.Param("pincode")
	var response db.Pincode
	// query is a map[string]interface{}
	query := map[string]interface{}{
		"pincode": pincode,
	}
	// Get the data from the database
	err := db.PincodeCollection.FindOne(context.Background(), query).Decode(&response)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Return the data as JSON
	ctx.JSON(http.StatusOK, response)
}

// func CreateOrUpdateUser(c *gin.Context) {
// 	var updateRequest UserUpdateRequest
// 	var userID string = ""
// 	phone := c.Query("phone")
// 	if phone != "new" {
// 		user, err := models.GetUser(phone)
// 		if err == nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "User doesn't exist "})
// 		}
// 		updateRequest.User.ID = user.ID
// 	}

// 	err := c.BindJSON(updateRequest)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid json " + err.Error()})
// 	}

// 	if userID == "" {
// 		user, err := updateRequest.User.Create()
// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user object " + err.Error()})
// 			return
// 		}
// 		userID = user.ID
// 	} else {
// 		err = updateRequest.User.Update()
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update user " + err.Error()})
// 			return
// 		}
// 	}

// 	updateRequest.DefaultAddress.UserID = userID
// 	if updateRequest.DefaultAddress.ID != "" {
// 		err = updateRequest.DefaultAddress.UpdateInDB()
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update default address " + err.Error()})
// 			return
// 		}
// 		err = models.UpdateDefaultAddress(updateRequest.DefaultAddress.ID, userID)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to make default address " + err.Error()})
// 			return
// 		}
// 	} else {
// 		err = updateRequest.DefaultAddress.SaveToDB()
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to make default address " + err.Error()})
// 			return
// 		}
// 		err = models.UpdateDefaultAddress(updateRequest.DefaultAddress.ID, userID)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to make default address " + err.Error()})
// 			return
// 		}
// 	}
// 	c.JSON(http.StatusOK, gin.H{"response": "success"})
// }

func CreateUser(c *gin.Context) {
	var updateRequest UserUpdateRequest
	phone := c.Query("phone")
	referralCode := c.Query("code")
	exists := models.UserExists(phone)
	if exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists "})
		return
	}
	err := c.BindJSON(&updateRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid json " + err.Error()})
		return
	}
	updateRequest.User.ID = data.GetUUIDString("user")

	var user models.User
	if len(referralCode) > 0 {
		user, err = updateRequest.User.CreateWithReferralCode(referralCode)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user object " + err.Error()})
			return
		}

	} else {
		user, err = updateRequest.User.Create()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user object " + err.Error()})
			return
		}
	}

	updateRequest.User = user
	if updateRequest.DefaultAddress.PinCode != "" {
		updateRequest.DefaultAddress.UserID = user.ID
		updateRequest.DefaultAddress.ID = data.GetUUIDString("address")
		err = updateRequest.DefaultAddress.SaveToDB()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update default address " + err.Error()})
			return
		}
		err = models.UpdateDefaultAddress(updateRequest.DefaultAddress.ID, user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to make default address " + err.Error()})
			return
		}
	}

	var profile models.Profile
	profile.UserID = user.ID

	err = profile.CreateIfNotExists()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create/update profile " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": updateRequest})
}

func GetUserAPI(c *gin.Context) {
	phone := c.Param("phone")
	user, err := models.GetUser(phone)
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found "})
		return
	}
	c.JSON(http.StatusOK, user)
}

func UpdateUser(c *gin.Context) {
	var updateRequest UserUpdateRequest
	phone := c.Query("phone")

	err := c.BindJSON(&updateRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid json " + err.Error()})
		return
	}

	user, err := models.GetUser(phone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User doesn't exist "})
		return
	}
	updateRequest.User.ID = user.ID

	err = updateRequest.User.Update()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update user " + err.Error()})
		return
	}

	updateRequest.DefaultAddress.UserID = updateRequest.User.ID
	if updateRequest.DefaultAddress.ID != "" {
		err = updateRequest.DefaultAddress.UpdateInDB()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update default address " + err.Error()})
			return
		}
		err = models.UpdateDefaultAddress(updateRequest.DefaultAddress.ID, updateRequest.User.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to make default address " + err.Error()})
			return
		}
	} else {
		updateRequest.DefaultAddress.ID = data.GetUUIDString("address")
		err = updateRequest.DefaultAddress.SaveToDB()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to make default address " + err.Error()})
			return
		}
		err = models.UpdateDefaultAddress(updateRequest.DefaultAddress.ID, updateRequest.User.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to make default address " + err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"response": updateRequest})
}

func StopUserMarketComm(c *gin.Context) {

	type requestFormat struct {
		UserPhone string `json:"phone"`
	}

	var requestBody requestFormat

	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		c.AbortWithError(400, err)
	}

	if !strings.HasPrefix(requestBody.UserPhone, "+") {
		requestBody.UserPhone = "+" + requestBody.UserPhone
	}

	user, err := models.GetUser(requestBody.UserPhone)
	if err != nil {
		fmt.Printf("Phone Err: %v\n", err)
		c.AbortWithError(404, err)
		return
	}

	err = user.TurnOffMarketingComm()
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.JSON(204, nil)
}

func GetUserInternalStatus(c *gin.Context) {

	adminID := c.Request.Header.Get("user")

	admin, err := models.GetAdminbyID(adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "couldn't found admin by ID:" + adminID + " error: " + err.Error()})
		return
	}

	user, err := models.GetUserByEmail(admin.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "couldn't found user by email:" + admin.Email + " error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func ToggleUserInternalExternalStatus(c *gin.Context) {

	adminID := c.Request.Header.Get("user")

	admin, err := models.GetAdminbyID(adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user, err := models.GetUserByEmail(admin.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user.Internal = !user.Internal

	err = models.ToggleUserInternalExternalStatus(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func JoinReferralProgram(c *gin.Context) {

	type requestFormat struct {
		UserPhone string   `json:"userPhone"`
		Name      string   `json:"name"`
		City      string   `json:"city"`
		Responses []string `json:"responses"`
	}

	// Read Request Body
	var requestBody requestFormat

	err := c.BindJSON(&requestBody)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		rw.JSONErrorResponse(c, 400, err)
		return
	}

	// Add '+' to phone number if not already present
	if !strings.HasPrefix(requestBody.UserPhone, "+") {
		requestBody.UserPhone = "+" + requestBody.UserPhone
	}

	// Get User by Phone number
	user, err := models.GetUser(requestBody.UserPhone)
	if err != nil {
		rw.JSONErrorResponse(c, 404, err)
		return
	}

	// Mark as Referral Program Joined
	err = user.JoinReferralProgram(requestBody.Name, requestBody.City, requestBody.Responses, 0)
	if err != nil {
		rw.JSONErrorResponse(c, 500, err)
		return
	}

	c.JSON(204, nil)
}

func MakeAmbassdorsFromCSV(c *gin.Context) {

	group := c.Query("testGroup")
	if group == "" {
		rw.JSONErrorResponse(c, 400, fmt.Errorf("test Group Invalid"))
		return
	}

	groupInt, err := strconv.ParseInt(group, 10, 64)
	if err != nil {
		rw.JSONErrorResponse(c, 400, fmt.Errorf("test Group Invalid"))
		return
	}

	records, err := csv.NewReader(c.Request.Body).ReadAll()
	if err != nil {
		rw.JSONErrorResponse(c, 400, err)
		return
	}

	err = models.CSVToAmbassdor(records, int(groupInt))
	if err != nil {
		rw.JSONErrorResponse(c, 400, err)
	}

	c.JSON(204, nil)
}
