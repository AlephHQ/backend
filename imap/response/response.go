package response

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
