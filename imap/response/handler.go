package response

type Handler interface {
	Handle(resp *Response) error
}

type HandlerFunc func(resp *Response) error

func NewHandlerFunc(f func(resp *Response) error) HandlerFunc {
	return HandlerFunc(f)
}

func (f HandlerFunc) Handle(resp *Response) error {
	return f(resp)
}
