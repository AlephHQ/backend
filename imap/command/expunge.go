package command

import "fmt"

type Expunge struct {
	Tag string
}

func NewCmdExpunge() *Expunge {
	return &Expunge{
		Tag: getTag(),
	}
}

func (e *Expunge) Command() string {
	return fmt.Sprintf("%s EXPUNGE", e.Tag)
}
