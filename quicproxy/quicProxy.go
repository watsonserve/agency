package quicproxy

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"time"

	quic "github.com/quic-go/quic-go"
	"github.com/watsonserve/goutils"
)

type QuicProxy struct {
	connChannel chan *quic.Conn
	signChannel chan *interface{}
	strmChannel chan *quic.Stream
	tlsCfg      *tls.Config
	quicConf    *quic.Config
}

func New(crt, key, ca string, idle time.Duration) (*QuicProxy, error) {
	quicConf := &quic.Config{
		KeepAlivePeriod: time.Duration(10) * time.Second,
		MaxIdleTimeout:  idle,
	}

	tlsCfg, err := goutils.GenTlsConfig(goutils.TLSFLAG_VERIFY, crt, key, ca)
	if nil != err {
		return nil, errors.New("GenTlsConfig " + err.Error())
	}

	return &QuicProxy{
		connChannel: make(chan *quic.Conn, 5),
		signChannel: make(chan *interface{}, 5),
		strmChannel: make(chan *quic.Stream, 5),
		tlsCfg:      tlsCfg,
		quicConf:    quicConf,
	}, nil
}

func (qp *QuicProxy) getAQuicConn() *quic.Stream {
	for 0 < len(qp.connChannel) {
		quicConn := <-qp.connChannel
		stream, err := quicConn.OpenStream()
		if nil == err {
			qp.connChannel <- quicConn
			return stream
		}
		quicConn.CloseWithError(0, "")
	}

	return nil
}

func (qp *QuicProxy) quicConnMgr() {
	for {
		_ = <-qp.signChannel
		qp.strmChannel <- qp.getAQuicConn()
	}
}

func (qp *QuicProxy) ListenAndServe(port string) {
	go qp.quicConnMgr()
	for {
		lis, err := quic.ListenAddr(port, qp.tlsCfg, qp.quicConf)
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

			qp.connChannel <- conn
		}

		lis.Close()
	}
}

func (qp *QuicProxy) GetAQuicConn() *quic.Stream {
	qp.signChannel <- nil
	return <-qp.strmChannel
}
