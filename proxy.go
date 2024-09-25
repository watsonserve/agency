package agency

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	quic "github.com/quic-go/quic-go"
	"github.com/watsonserve/goutils"
)

type FullDuplexStream interface {
	io.ReadWriteCloser
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

type Proxy struct {
	channel      chan quic.Connection
	TlsCfg       *tls.Config
	QuicConf     *quic.Config
	BufSiz       int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func New(crt, key, ca string, bufsiz, rTimeout, wTimeout int) (*Proxy, error) {
	quicConf := &quic.Config{
		KeepAlivePeriod: time.Duration(10) * time.Second,
		MaxIdleTimeout:  time.Duration(30) * time.Second,
	}

	tlsCfg, err := goutils.GenTlsConfig(goutils.TLSFLAG_VERIFY, crt, key, ca)
	if nil != err {
		fmt.Fprintln(os.Stderr, "GenTlsConfig "+err.Error())
		return nil, err
	}

	return &Proxy{
		channel:      make(chan quic.Connection, 5),
		TlsCfg:       tlsCfg,
		QuicConf:     quicConf,
		BufSiz:       bufsiz,
		ReadTimeout:  time.Duration(rTimeout) * time.Second,
		WriteTimeout: time.Duration(wTimeout) * time.Second,
	}, nil
}

func (p *Proxy) srv(port string) {
	for {
		lis, err := quic.ListenAddr(port, p.TlsCfg, p.QuicConf)
		if nil != err {
			fmt.Fprintln(os.Stderr, "QuicListenAddr "+err.Error())
			return
		}

		for {
			conn, err := lis.Accept(context.Background())
			if nil != err {
				fmt.Fprintln(os.Stderr, "accept connect", err)
				break
			}

			p.channel <- conn
		}

		lis.Close()
	}
}

func (p *Proxy) getAQuicConn(channel chan quic.Connection) quic.Stream {
	for 0 < len(channel) {
		quicConn := <-channel
		stream, err := quicConn.OpenStream()
		if nil == err {
			channel <- quicConn
			return stream
		}
		quicConn.CloseWithError(0, "")
	}

	return nil
}

func (p *Proxy) proxyTransportLayer(comeFrom FullDuplexStream) {
	upStream := p.getAQuicConn(p.channel)

	defer (func() {
		if nil != upStream {
			upStream.Close()
		}
		if nil != comeFrom {
			comeFrom.Close()
		}
		log.Printf("Transport Closed")
	})()

	bufSiz := p.BufSiz
	if nil == upStream || nil == comeFrom || bufSiz < 1 {
		log.Println("refused: invoid stream or bufsize")
		return
	}

	var written int64 = 0
	var err error = nil

	go io.CopyBuffer(upStream, comeFrom, make([]byte, bufSiz))

	readTimeout := p.ReadTimeout
	writeTimeout := p.WriteTimeout
	if 0 < readTimeout {
		upStream.SetReadDeadline(time.Now().Add(readTimeout))
	}
	if 0 < writeTimeout {
		comeFrom.SetReadDeadline(time.Now().Add(writeTimeout))
	}
	written, err = io.CopyBuffer(comeFrom, upStream, make([]byte, bufSiz))

	errMsg := "success"
	if nil != err {
		errMsg = "failed -- " + err.Error()
	}
	log.Printf("transfer data: length=%d, status=%s", written, errMsg)
}

func (p *Proxy) ListenAndServe(comeFrom, upStream string) error {
	go p.srv(upStream)

	network := "tcp"
	if strings.Contains(comeFrom, "/") {
		network = "unix"
	}

	lis, err := net.Listen(network, comeFrom)
	if nil != err {
		return err
	}

	defer lis.Close()

	for {
		conn, err := lis.Accept()
		if nil != err {
			return err
		}

		go p.proxyTransportLayer(conn)
	}
}
