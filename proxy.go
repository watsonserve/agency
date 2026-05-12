package agency

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/watsonserve/agency/quicproxy"
)

type Proxy struct {
	quicproxy.QuicProxy
	lc           *net.ListenConfig
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func New(crt, key, ca string, idleTime, rTimeout, wTimeout int) (*Proxy, error) {
	idle := time.Duration(idleTime) * time.Second

	qp, err := quicproxy.New(crt, key, ca, idle)
	if nil != err {
		return nil, err
	}
	return &Proxy{
		QuicProxy: *qp,
		lc: &net.ListenConfig{
			KeepAlive: idle,
		},
		ReadTimeout:  time.Duration(rTimeout) * time.Second,
		WriteTimeout: time.Duration(wTimeout) * time.Second,
	}, nil
}

func (p *Proxy) proxyTransportLayer(comeStream FullDuplexStream) {
	rTimeout := p.ReadTimeout
	wTimeout := p.WriteTimeout

	upStream := p.GetAQuicConn()
	pipe(upStream, comeStream, rTimeout, wTimeout)
}

func (p *Proxy) ListenAndServe(comeFrom, upStream string) error {
	go p.QuicProxy.ListenAndServe(upStream)

	network := "tcp"
	if strings.Contains(comeFrom, "/") {
		network = "unix"
	}

	lis, err := p.lc.Listen(context.Background(), network, comeFrom)
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
