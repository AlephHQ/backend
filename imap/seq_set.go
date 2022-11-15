package imap

import (
	"fmt"
	"strconv"
	"strings"
)

type SeqRange struct {
	From uint64
	To   uint64
}

type SeqSet []SeqRange

func (seqset *SeqSet) String() string {
	sets := make([]string, 0)
	for _, seq := range *seqset {
		if seq.From == seq.To {
			sets = append(sets, strconv.FormatUint(seq.From, 10))
		} else {
			sets = append(
				sets,
				fmt.Sprintf(
					"%s:%s",
					strconv.FormatUint(seq.From, 10),
					strconv.FormatUint(seq.To, 10),
				),
			)
		}
	}

	return strings.Join(sets, ",")
}
