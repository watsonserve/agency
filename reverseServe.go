package agency

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"

	quic "github.com/quic-go/quic-go"
)

func talkWR(stream quic.Stream, handle http.HandlerFunc) {
	var strHeader string
	_, err := fmt.Fscanf(stream, "%s\r\n\r\n", &strHeader)
	if nil != err {
		fmt.Fprintln(os.Stderr, "read data", err)
		return
	}

	lines := strings.Split(strHeader, "\r\n")
	requestTo := strings.Split(lines[0], " ")
	req, err := http.NewRequest(requestTo[0], requestTo[1], stream)
	if nil != err {
		fmt.Fprintln(os.Stderr, "make request", err)
		return
	}

	req.Header = make(http.Header)
	for _, item := range lines[1:] {
		kv := strings.Split(item, ":")
		req.Header.Add(kv[0], strings.TrimSpace(kv[1]))
	}

	handle(newResponse(stream), req)
	stream.Close()
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
