package agency

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	quic "github.com/quic-go/quic-go"
)

func talkWR(stream quic.Stream, handle http.HandlerFunc) {
	headerList := make([]string, 0)
	fmt.Println("recv a conn")
	reader := bufio.NewReader(stream)
	for {
		line, err := reader.ReadString('\n')
		if line == "\r\n" || nil != err && io.EOF != err {
			break
		}
		fmt.Println(len(line), line)
		headerList = append(headerList, line)
	}
	fmt.Println("read header done")

	requestTo := strings.Split(headerList[0], " ")
	req, err := http.NewRequest(requestTo[0], requestTo[1], stream)
	if nil != err {
		fmt.Fprintln(os.Stderr, "make request", err)
		return
	}

	req.Header = make(http.Header)
	for _, item := range headerList[1:] {
		kv := strings.Split(item, ":")
		kv[1] = strings.TrimSpace(kv[1])
		fmt.Printf("header: %s: %s,\n", kv[0], kv[1])
		req.Header.Add(kv[0], kv[1])
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
