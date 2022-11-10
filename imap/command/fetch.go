package command

import (
	"fmt"
	"ncp/backend/imap"
	"strings"
)

type Fetch struct {
	Tag       string
	DataItems []*imap.DataItem
	Macro     imap.FetchMacro
	SeqSet    *imap.SeqSet
}

func NewCmdFetch(seqset *imap.SeqSet) *Fetch {
	return &Fetch{
		Tag:       getTag(),
		SeqSet:    seqset,
		DataItems: make([]*imap.DataItem, 0),
	}
}

func (f *Fetch) SetMacro(m imap.FetchMacro) *Fetch {
	f.Macro = m

	return f
}

func (f *Fetch) AppendDataItem(di *imap.DataItem) *Fetch {
	f.DataItems = append(f.DataItems, di)

	return f
}

func (f *Fetch) Command() string {
	if len(f.DataItems) == 0 {
		return fmt.Sprintf("%s FETCH %s %s", f.Tag, f.SeqSet.String(), f.Macro)
	}

	items := make([]string, 0)
	for _, item := range f.DataItems {
		name := string(item.Name)
		if item.Section != "" {
			name = fmt.Sprintf("%s[%s]", item.Name, item.Section)
		}

		items = append(items, name)
	}

	return fmt.Sprintf("%s FETCH %s (%s)", f.Tag, f.SeqSet.String(), strings.Join(items, " "))
}
