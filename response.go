package agency

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

const RESPONSE_BUF_SIZ = 512

type Response struct {
	statusCode int
	header     http.Header
	stream     io.Writer
	headerSent bool
	buf        bytes.Buffer
}

func newResponse(stream io.Writer) *Response {
	return &Response{
		headerSent: false,
		statusCode: 200,
		header:     make(http.Header),
		stream:     stream,
		buf:        *bytes.NewBuffer(make([]byte, RESPONSE_BUF_SIZ)),
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
	fmt.Printf("debug: WriteHeader %d\n", statusCode)
	if 100 <= statusCode && statusCode < 200 {
		resp.sendHeader(statusCode)
		return
	}
	resp.statusCode = statusCode
}

func (resp *Response) Write(body []byte) (int, error) {
	if nil == body {
		return 0, nil
	}

	fmt.Printf("debug: WriteBody %d\n", len(body))
	n := len(body)
	if resp.buf.Len()+n < RESPONSE_BUF_SIZ {
		resp.buf.Write(body)
		return n, nil
	}

	return resp.Flush(body)
}

func (resp *Response) Flush(body []byte) (int, error) {
	var err error = nil
	if !resp.headerSent {
		cLen := resp.header.Get("Content-Length")
		if nil == body && "" == cLen {
			resp.header.Set("Content-Length", fmt.Sprintf("%d", resp.buf.Len()))
		}
		err = resp.sendHeader(resp.statusCode)
	}
	if nil != err {
		return 0, err
	}

	if 0 < resp.buf.Len() {
		_, err = resp.stream.Write(resp.buf.Bytes())
		if nil != err {
			return 0, err
		}
		resp.buf.Reset()
	}

	if nil == body || len(body) < 1 {
		return 0, nil
	}

	return resp.stream.Write(body)
}
