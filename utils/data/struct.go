package data

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
)

func InterfaceToIoReader(data interface{}) io.Reader {
	jsonData, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	return bytes.NewReader(jsonData)
}

func Contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// func to mask phone number
func MaskPhoneNumber(phoneNumber string) string {
	if len(phoneNumber) < 10 {
		return phoneNumber
	}
	return phoneNumber[:len(phoneNumber)-10] + "XXXXXX" + phoneNumber[len(phoneNumber)-4:]
}

// make 64 bit md5 hash
func Md5Hash(data string) string {
	hasher := md5.New()
	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil))
}
