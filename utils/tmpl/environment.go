package tmpl

import (
	"fmt"
	"hermes/models"
)

func IsProd() bool {
	env := GetEnvVariable("ENVIRONMENT")
	if env == "prod" {
		return true
	}
	return false
}

func GetImageURLByEnvironment(size int, url models.Image) string {
	sizeStr := fmt.Sprint(size)
	urlStr := string(url)
	env := GetEnvVariable("ENVIRONMENT")
	if env == "prod" {
		return "https://roovo.in/cdn-cgi/image/width=" + sizeStr + ",format=auto/https://storage.googleapis.com/roovo-images/rawImages/" + urlStr
	}
	return "https://storage.googleapis.com/roovo-images/rawImages/" + urlStr
}

func GetImageURLByEnvironment2(size int, id string) string {
	sizeStr := fmt.Sprint(size)
	fmt.Printf("sizeStr: %v\n", sizeStr)
	fmt.Printf("id: %v\n", id)
	env := GetEnvVariable("ENVIRONMENT")
	if env == "prod" {
		return "https://roovo.in/cdn-cgi/image/width=" + sizeStr + ",format=auto/https://storage.googleapis.com/roovo-images/rawImages/" + id
	}
	return "https://storage.googleapis.com/roovo-images/rawImages/" + id
}
