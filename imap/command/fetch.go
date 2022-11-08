package command

import "fmt"

type Fetch struct {
	Tag string
}

func NewCmdFetch() *Fetch {
	return &Fetch{
		Tag: getTag(),
	}
}

func (f *Fetch) Command() string {
	return fmt.Sprintf("%s FETCH 6 (BODY TEXT)", f.Tag)
}
