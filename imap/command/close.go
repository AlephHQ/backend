package command

import "fmt"

type Close struct {
	Tag string
}

func NewCmdClose() *Close {
	return &Close{
		Tag: getTag(),
	}
}

func (c *Close) Command() string {
	return fmt.Sprintf("%s CLOSE", c.Tag)
}
