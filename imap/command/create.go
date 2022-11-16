package command

import "fmt"

type Create struct {
	Tag  string
	Name string
}

func NewCmdCreate(name string) *Create {
	return &Create{
		Tag:  getTag(),
		Name: name,
	}
}

func (c *Create) Command() string {
	return fmt.Sprintf("%s CREATE %s", c.Tag, c.Name)
}
