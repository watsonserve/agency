package agency

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	quic "github.com/quic-go/quic-go"
)

func talkWR(stream *quic.Stream, handle http.HandlerFunc) {
	defer stream.Close()

	headerList := make([]string, 0)
	reader := bufio.NewReader(stream)
	for {
		line, err := reader.ReadString('\n')
		if line == "\r\n" || nil != err && io.EOF != err {
			break
		}
		headerList = append(headerList, line)
	}

	requestTo := strings.Split(headerList[0], " ")
	req, err := http.NewRequest(requestTo[0], requestTo[1], stream)
	if nil != err {
		fmt.Fprintln(os.Stderr, "make request", err)
		return
	}

	req.Header = make(http.Header)
	for _, item := range headerList[1:] {
		kv := strings.SplitN(item, ":", 2)
		kv[1] = strings.TrimSpace(kv[1])
		req.Header.Add(kv[0], kv[1])
	}

	resp := newResponse(stream)
	handle(resp, req)
	resp.Flush(nil)
}

func ReverseServe(network string, tlsCfg *tls.Config, quicConf *quic.Config, handle http.Handler) {
	for {
		conn, err := quic.DialAddr(context.Background(), network, tlsCfg, quicConf)
		if nil != err {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		for {
			stream, err := conn.AcceptStream(context.Background())
			if nil != err {
				fmt.Fprintln(os.Stderr, "waiting on conn", err)
				break
			}

			go talkWR(stream, handle.ServeHTTP)
		}

		conn.CloseWithError(0, "")
	}
}

func transStream(upstream string, stream *quic.Stream, timeout time.Duration) {
	upConn, err := net.Dial("tcp", upstream)
	if nil != err {
		stream.Close()
		return
	}
	pipe(upConn, stream, timeout, timeout)
}

func ReverseTrans(network string, tlsCfg *tls.Config, quicConf *quic.Config, upstream string, timeout time.Duration) {
	for {
		conn, err := quic.DialAddr(context.Background(), network, tlsCfg, quicConf)
		if nil != err {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		for {
			stream, err := conn.AcceptStream(context.Background())
			if nil != err {
				fmt.Fprintln(os.Stderr, "waiting on conn", err)
				break
			}

			go transStream(upstream, stream, timeout)
		}

		conn.CloseWithError(0, "")
	}
}
