package utils

import (
	"crypto/rand"
	"encoding/base64"
	"log"
)

// RandStr returns a random string of length n
func RandStr(n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		log.Panic(err)
	}

	return base64.RawURLEncoding.EncodeToString(b)
}
