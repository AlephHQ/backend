package client

import (
	"crypto/rand"
	"encoding/base64"
	"log"
)

func getTag() string {
	b := make([]byte, 7)
	_, err := rand.Read(b)
	if err != nil {
		log.Panic(err)
	}

	return base64.RawURLEncoding.EncodeToString(b)
}
