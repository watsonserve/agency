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
	headerSent bool
}

func newResponse(stream io.Writer) *Response {
	return &Response{
		headerSent: false,
		statusCode: 200,
		header:     make(http.Header),
		stream:     stream,
	}
}

func (resp *Response) Header() http.Header {
	return resp.header
}

func (resp *Response) sendHeader(statusCode int) error {
	strStatus := http.StatusText(statusCode)
	strHeader := ""
	// resp.header.Set("Content-Length", fmt.Sprintf("%d", len(body)))
	for k, vals := range resp.header {
		for _, item := range vals {
			strHeader += k + ": " + item + "\r\n"
		}
	}
	_, err := fmt.Fprintf(resp.stream, "HTTP/1.1 %03d %s\r\n%s\r\n", statusCode, strStatus, strHeader)
	if nil == err {
		resp.headerSent = true
	}
	return err
}

func (resp *Response) WriteHeader(statusCode int) {
	if 100 <= statusCode && statusCode < 200 {
		resp.sendHeader(statusCode)
		return
	}
	resp.statusCode = statusCode
}

func (resp *Response) Write(body []byte) (int, error) {
	var err error = nil
	var bn int = 0

	if !resp.headerSent {
		err = resp.sendHeader(resp.statusCode)
	}
	if nil == err {
		bn, err = resp.stream.Write(body)
	}
	return bn, err
}
