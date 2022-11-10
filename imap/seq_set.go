package imap

import (
	"fmt"
	"strconv"
)

type SeqSet struct {
	From uint64
	To   uint64
}

func NewSeqSet(from, to uint64) *SeqSet {
	return &SeqSet{
		From: from,
		To:   to,
	}
}

func (seqset *SeqSet) String() string {
	if seqset.From == seqset.To {
		return strconv.FormatUint(seqset.From, 10)
	}

	return fmt.Sprintf(
		"%s:%s",
		strconv.FormatUint(seqset.From, 10),
		strconv.FormatUint(seqset.To, 10),
	)
}
