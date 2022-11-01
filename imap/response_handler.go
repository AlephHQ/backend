package imap

type ResponseHandler interface {
	Handle(resp *Response) error
}

type HandlerFunc func(resp *Response) error

type FetchHandler struct {
	Messages chan string
}

func (fh *FetchHandler) Handle(resp *Response) error {
	return nil
}
