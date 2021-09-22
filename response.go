package main

type Response struct {
	body string
}

func (r *Response) Write(p []byte) (n int, err error) {
	r.body += string(p)
	return len(p), nil
}

func (r *Response) Clear() {
	r.body = ""
}

func NewResponse() *Response {
	return &Response{}
}
