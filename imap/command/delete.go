package command

import "fmt"

type Delete struct {
	Tag  string
	Name string
}

func NewCmdDelete(name string) *Delete {
	return &Delete{
		Tag:  getTag(),
		Name: name,
	}
}

func (d *Delete) Command() string {
	return fmt.Sprintf("%s DELETE %s", d.Tag, d.Name)
}
