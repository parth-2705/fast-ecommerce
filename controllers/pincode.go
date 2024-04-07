package controllers

import (
	ctx "context"
	"fmt"
	"hermes/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
