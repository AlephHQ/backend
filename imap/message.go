package imap

type Message struct {
	UID uint64

	SeqNum uint64

	Flags map[string]bool

	InternalDate string

	Size uint64

	Envelope interface{}

	Body interface{}

	Text string
}
