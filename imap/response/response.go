package response

import (
	"errors"
	"strings"
)

type Response struct {
	// Raw contains the original response in its raw format
	Raw string

	// Fields contains all the different fields received
	// in the response
	Fields []interface{}
}

func NewResponse() *Response {
	resp := &Response{
		Fields: make([]interface{}, 0),
	}

	return resp
}

func (resp *Response) AddField(field interface{}) {
	resp.Fields = append(resp.Fields, field)
}

func (resp *Response) Error() error {
	err := make([]string, 0)
	for _, word := range resp.Fields[2:] {
		err = append(err, word.(string))
	}

	return errors.New(strings.Join(err, " "))
}
