package command

import "fmt"

type FetchMacro string

const (
	FetchMacroAll  FetchMacro = "ALL"
	FetchMacroFast FetchMacro = "FAST"
	FetchMacroFull FetchMacro = "FULL"
)

type Fetch struct {
	Tag   string
	Macro FetchMacro
}

func NewCmdFetch(macro FetchMacro) *Fetch {
	return &Fetch{
		Tag:   getTag(),
		Macro: macro,
	}
}

func (f *Fetch) Command() string {
	return fmt.Sprintf("%s FETCH 1:15 %s", f.Tag, f.Macro)
}
