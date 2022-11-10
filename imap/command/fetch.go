package command

import (
	"fmt"
	"ncp/backend/imap"
)

type Fetch struct {
	Tag   string
	Macro imap.FetchMacro
}

func NewCmdFetch(macro imap.FetchMacro) *Fetch {
	return &Fetch{
		Tag:   getTag(),
		Macro: macro,
	}
}

func (f *Fetch) Command() string {
	// return fmt.Sprintf("%s FETCH 1:15 %s", f.Tag, f.Macro)
	return fmt.Sprintf("%s FETCH 5 (BODY[TEXT])", f.Tag)
}
