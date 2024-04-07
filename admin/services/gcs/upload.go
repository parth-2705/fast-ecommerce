package GCS

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tryamigo/themis"
)

func UploadFileToBucket(fileKey string, fileName string, c *gin.Context) {
	f, err := c.FormFile(fileKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	blobFile, err := f.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = Uploader.UploadFile(blobFile, fileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	GCS_BUCKET_NAME, err := themis.GetSecret("GCS_BUCKET_NAME")
	if err != nil {
		fmt.Println(fmt.Errorf("error: %s", err.Error()))
		c.AbortWithError(http.StatusBadGateway, err)
	}

	GCS_UPLOAD_FOLDER, err := themis.GetSecret("GCS_UPLOAD_FOLDER")
	if err != nil {
		fmt.Println(fmt.Errorf("error: %s", err.Error()))
		c.AbortWithError(http.StatusBadGateway, err)
	}

	url := "https://storage.googleapis.com/" + GCS_BUCKET_NAME + "/" + GCS_UPLOAD_FOLDER + "/" + fileName
	c.JSON(http.StatusOK, gin.H{"response": "Success", "url": url, "fileName": fileName})
}

func UploadProfileImage(c *gin.Context) {
	fileKey := "image"
	fileName := fmt.Sprint(time.Now().UTC().UnixNano() / 1e6)
	UploadFileToBucket(fileKey, fileName, c)
}

func UploadMedia(c *gin.Context) {
	fileKey := c.Query("type")
	fileName := fmt.Sprint(time.Now().UTC().UnixNano() / 1e6)
	UploadFileToBucket(fileKey, fileName, c)
}

func UploadBulkActionCSV(c *gin.Context) {
	fileKey := "csv"
	fileName := "bulk/action/product/" + fmt.Sprint(time.Now().UTC().UnixNano()/1e6)
	UploadFileToBucket(fileKey, fileName, c)
}

func UploafFileFromDiskToBucket(path string) (bktLink string, err error) {
	rf, err := os.Open(path)
	if err != nil {
		fmt.Printf("file open err: %v\n", err)
		return
	}
	defer rf.Close()

	err = Uploader.UploadFile(rf, path)
	if err != nil {
		fmt.Printf("upload err: %v\n", err)
		return
	}

	// Get bucket link of Image
	GCS_BUCKET_NAME, err := themis.GetSecret("GCS_BUCKET_NAME")
	if err != nil {
		fmt.Println(fmt.Errorf("error: %s", err.Error()))
		return
	}
	GCS_UPLOAD_FOLDER, err := themis.GetSecret("GCS_UPLOAD_FOLDER")
	if err != nil {
		fmt.Println(fmt.Errorf("error: %s", err.Error()))
		return
	}

	bktLink = "https://storage.googleapis.com/" + GCS_BUCKET_NAME + "/" + GCS_UPLOAD_FOLDER + "/" + path
	return bktLink, nil
}
