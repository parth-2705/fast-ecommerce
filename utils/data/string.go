package data

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateRandomHex(length int) (string, error) {
	bytes := make([]byte, length/2)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}
