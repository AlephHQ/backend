package command

import (
	"fmt"
	"ncp/backend/imap"
	"strings"
)

type Store struct {
	Tag          string
	SeqSet       []imap.SeqSet
	DataItemName imap.DataItemName
	Values       []imap.Flag
}

func NewCmdStore(seqset []imap.SeqSet, name imap.DataItemName, values []imap.Flag) *Store {
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

	seqset := make([]string, 0)
	for _, set := range s.SeqSet {
		seqset = append(seqset, set.SeqSet())
	}

	return fmt.Sprintf("%s STORE %v %v (%s)", s.Tag, strings.Join(seqset, ","), s.DataItemName, strings.Join(vals, " "))
}
