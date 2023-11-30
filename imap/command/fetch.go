package command

import (
	"aleph/backend/imap"
	"fmt"
	"strings"
)

type Fetch struct {
	Tag       string
	DataItems []*imap.DataItem
	Macro     imap.FetchMacro
	SeqSet    []imap.SeqSet
}

func NewCmdFetch(seqset []imap.SeqSet) *Fetch {
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
	seqset := make([]string, 0)
	for _, set := range f.SeqSet {
		seqset = append(seqset, set.SeqSet())
	}

	if len(f.DataItems) == 0 {
		return fmt.Sprintf("%s FETCH %s %s", f.Tag, strings.Join(seqset, ","), f.Macro)
	}

	items := make([]string, 0)
	for _, item := range f.DataItems {
		name := string(item.Name)
		if item.Section != "" {
			name = name + "[" + string(item.Section) + "]"
		}

		if item.Partial != "" {
			name = name + "<" + item.Partial + ">"
		}

		items = append(items, name)
	}

	return fmt.Sprintf("%s FETCH %s (%s)", f.Tag, strings.Join(seqset, ","), strings.Join(items, " "))
}
