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
	Values       []string
}

func NewCmdStore(seqset *imap.SeqSet, name imap.DataItemName, values []string) *Store {
	return &Store{
		Tag:          getTag(),
		SeqSet:       seqset,
		DataItemName: name,
		Values:       values,
	}
}

func (s *Store) Command() string {
	return fmt.Sprintf("%s STORE %v %v (%s)", s.Tag, s.SeqSet, s.DataItemName, strings.Join(s.Values, " "))
}
