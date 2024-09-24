package agency

import (
	"fmt"
	"io"
	"net/http"
)

type Response struct {
	statusCode int
	header     http.Header
	stream     io.Writer
}

func newResponse(stream io.Writer) *Response {
	return &Response{
		statusCode: 200,
		header:     make(http.Header),
		stream:     stream,
	}
}

func (resp *Response) Header() http.Header {
	return resp.header
}

func (resp *Response) WriteHeader(statusCode int) {
	resp.statusCode = statusCode
}

func (resp *Response) Write(body []byte) (int, error) {
	code := resp.statusCode
	strStatus := http.StatusText(code)
	strHeader := ""
	resp.header.Set("Content-Length", fmt.Sprintf("%d", len(body)))
	for k, vals := range resp.header {
		for _, item := range vals {
			strHeader += k + ": " + item + "\r\n"
		}
	}
	var bn int = 0
	header := fmt.Sprintf("HTTP/1.1 %03d %s\r\n%s\r\n", code, strStatus, strHeader)
	hn := len(header)
	fmt.Println("response:\n", header)
	resp.stream.Write([]byte(header))
	// if nil == err {
	bn, err := resp.stream.Write(body)
	// }
	return hn + bn, err
}
