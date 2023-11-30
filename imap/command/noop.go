package command

import "fmt"

type NOOP struct {
	Tag string
}

func NewCmdNoop() *NOOP {
	return &NOOP{
		Tag: getTag(),
	}
}

func (n *NOOP) Command() string {
	return fmt.Sprintf("%s NOOP", n.Tag)
}
