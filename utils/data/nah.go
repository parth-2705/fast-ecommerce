package data

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

func SHA256Hash(input string) (output string) {

	hs := sha256.New()

	hs.Write([]byte(input))

	return fmt.Sprintf("%x", hs.Sum(nil))

}

func NormalizePhoneNumber(input string) (output string) {

	return strings.TrimLeft(input, "+0")

}

func NormalizePhoneAndHash(input string) (output string) {

	normalised := NormalizePhoneNumber(input)
	output = SHA256Hash(normalised)
	return
}
