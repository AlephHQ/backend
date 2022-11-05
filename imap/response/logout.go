package response

type Logout struct {
	Tag string
}

func NewHandlerLogout(tag string) *Logout {
	return &Logout{tag}
}

func (l *Logout) Handle(resp *Response) (bool, error) {
	return true, nil
}
