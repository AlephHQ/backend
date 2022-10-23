package imap

import (
	"log"
	"time"
)

// RFC3501 Section 3
type ConnectionState int

const (
	NotAuthenticatedState ConnectionState = 0
	AuthenticatedState    ConnectionState = 1
	SelectedState         ConnectionState = 2
	LogoutState           ConnectionState = 3
)

func Run() {
	log.Println("Running ...")
	time.Sleep(10 * time.Second)
	log.Println("Done * BYE")
}
