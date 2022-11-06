package response

type Response struct {
	// Raw contains the original response in its raw format
	Raw string

	// Fields contains all the different fields received
	// in the response
	Fields []string

	// Tagged indicates whether this is a tagged response
	Tagged bool
}

func NewResponse(raw string) *Response {
	resp := &Response{}
	resp.Raw = raw
	resp.Fields = make([]string, 0)

	return resp
}

func (resp *Response) AddField(field string) {
	resp.Fields = append(resp.Fields, field)
}
