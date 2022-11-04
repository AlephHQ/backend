package response

type Handler interface {
	Handle(resp *Response) (bool, error)
}

type HandlerFunc func(resp *Response) (bool, error)

func NewHandlerFunc(f func(resp *Response) (bool, error)) HandlerFunc {
	return HandlerFunc(f)
}

func (f HandlerFunc) Handle(resp *Response) (bool, error) {
	return f(resp)
}
