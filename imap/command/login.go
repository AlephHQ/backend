package command

import "fmt"

type Login struct {
	Tag string

	username string
	password string
}

func NewCmdLogin(username, password string) *Login {
	return &Login{
		Tag:      getTag(),
		username: username,
		password: password,
	}
}

func (l *Login) Command() string {
	return fmt.Sprintf("%s LOGIN %s %s", l.Tag, l.username, l.password)
}
