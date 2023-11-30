package imap

import (
	"fmt"
	"strconv"
)

type SeqSet interface {
	SeqSet() string
}

type SeqRange struct {
	From uint64
	To   uint64
}

func (sr *SeqRange) SeqSet() string {
	return fmt.Sprintf(
		"%s:%s",
		strconv.FormatUint(sr.From, 10),
		strconv.FormatUint(sr.To, 10),
	)
}

type SeqNumber struct {
	Val uint64
}

func (sn *SeqNumber) SeqSet() string {
	return strconv.FormatUint(sn.Val, 10)
}
