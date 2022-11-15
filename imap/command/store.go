package command

import (
	"fmt"
	"ncp/backend/imap"
	"strings"
)

type Store struct {
	Tag          string
	SeqSet       *imap.SeqSet
	DataItemName imap.DataItemName
	Values       []imap.Flag
}

func NewCmdStore(seqset *imap.SeqSet, name imap.DataItemName, values []imap.Flag) *Store {
	return &Store{
		Tag:          getTag(),
		SeqSet:       seqset,
		DataItemName: name,
		Values:       values,
	}
}

func (s *Store) Command() string {
	vals := make([]string, 0)
	for _, flag := range s.Values {
		vals = append(vals, string(flag))
	}

	return fmt.Sprintf("%s STORE %v %v (%s)", s.Tag, s.SeqSet, s.DataItemName, strings.Join(vals, " "))
}
