package agency

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	quic "github.com/quic-go/quic-go"
	"github.com/watsonserve/goutils"
)

type Proxy struct {
	channel      chan quic.Connection
	TlsCfg       *tls.Config
	QuicConf     *quic.Config
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func New(crt, key, ca string, rTimeout, wTimeout int) (*Proxy, error) {
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
	rTimeout := p.ReadTimeout
	wTimeout := p.WriteTimeout

	upStream := p.getAQuicConn(p.channel)
	pipe(upStream, comeFrom, rTimeout, wTimeout)
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
