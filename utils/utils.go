package utils

import (
	"encoding/base64"
	"log"
	"math/rand"
)

var lowercase_digits = "qwertyuiopasdfghjklzxcvbnm1234567890"

// RandStr returns a random string of length n
func RandStr(n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		log.Panic(err)
	}

	return base64.RawURLEncoding.EncodeToString(b)
}

// RandStrLower returns a random string of digits and lowercase
// letters
func RandStrLower(n int) string {
	b := make([]byte, n)

	for i := range b {
		b[i] = lowercase_digits[rand.Int63()%int64(len(lowercase_digits))]
	}

	return string(b)
}
