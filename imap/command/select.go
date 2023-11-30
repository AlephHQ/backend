package command

import "fmt"

type Select struct {
	Tag     string
	Mailbox string
}

func NewCmdSelect(mbox string) *Select {
	return &Select{
		Tag:     getTag(),
		Mailbox: mbox,
	}
}

func (s *Select) Command() string {
	return fmt.Sprintf("%s SELECT %s", s.Tag, s.Mailbox)
}
