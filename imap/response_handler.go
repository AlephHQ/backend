package imap

type ResponseHandler interface {
	Handle(resp *Response) error
}

type HandlerFunc func(resp *Response) error
