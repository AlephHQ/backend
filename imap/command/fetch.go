package command

import (
	"fmt"
	"ncp/backend/imap"
)

type Fetch struct {
	Tag    string
	Macro  imap.FetchMacro
	SeqSet *imap.SeqSet
}

func NewCmdFetch(seqset *imap.SeqSet, macro imap.FetchMacro) *Fetch {
	return &Fetch{
		Tag:    getTag(),
		Macro:  macro,
		SeqSet: seqset,
	}
}

func (f *Fetch) Command() string {
	return fmt.Sprintf("%s FETCH %s %s", f.Tag, f.SeqSet.String(), f.Macro)
}
