package command

import "ncp/backend/imap"

type Store struct {
	Tag          string
	SeqSet       *imap.SeqSet
	DataItemName imap.DataItemName
	Values       []string
}

func NewCmdStore(seqset *imap.SeqSet, name imap.DataItemName) *Store {
	return &Store{
		Tag:          getTag(),
		SeqSet:       seqset,
		DataItemName: name,
	}
}

func (s *Store) AddValue(val string) *Store {
	s.Values = append(s.Values, val)

	return s
}
