package imap

import (
	"log"
	"time"
)

// RFC3501 Section 3
type ConnectionState int

const (
	ConnectingState       ConnectionState = 0
	NotAuthenticatedState ConnectionState = 1
	AuthenticatedState    ConnectionState = 2
	SelectedState         ConnectionState = 3
	LogoutState           ConnectionState = 4
)

func Run() {
	log.Println("Running ...")
	time.Sleep(10 * time.Second)
	log.Println("Done * BYE")
}
