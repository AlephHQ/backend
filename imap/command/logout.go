package command

import "fmt"

type Logout struct {
	Tag string
}

func NewCmdLogout() *Logout {
	return &Logout{
		Tag: getTag(),
	}
}

func (l *Logout) Command() string {
	return fmt.Sprintf("%s LOGOUT", l.Tag)
}
