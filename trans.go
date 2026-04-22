package agency

import (
	"errors"
	"fmt"
	"io"
	"log"
	"time"
)

var logLev int = 3

const (
	LOG_LEV_ERR   = 0
	LOG_LEV_WARN  = 1
	LOG_LEV_INFO  = 2
	LOG_LEV_DEBUG = 4
)

func SetLogLev(lev int) {
	logLev = lev
}

type FullDuplexStream interface {
	io.ReadWriteCloser
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

func closeStream(foo, bar FullDuplexStream) {
	if nil != foo {
		foo.Close()
	}
	if nil != bar {
		bar.Close()
	}
	// log.Printf("Transport Closed")
}

func copyBuffer(dst io.Writer, src io.Reader) (written int64, err error) {
	bufSiz := 32 * 1024
	if l, ok := src.(*io.LimitedReader); ok {
		if l.N < 1 {
			bufSiz = 1
		} else {
			bufSiz = int(l.N)
		}
	}
	buf := make([]byte, bufSiz)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errors.New("InvalidWrite")
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
			if 4 == (logLev & 4) {
				fmt.Println(string(buf[0:nr]))
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}

func sendStream(load string, dst, src FullDuplexStream) {
	defer closeStream(dst, src)

	written, err := copyBuffer(dst, src)
	errMsg := "success"
	if nil != err {
		errMsg = "failed -- " + err.Error()
	}
	log.Printf("%s_result --length=%d --status=%s", load, written, errMsg)
}

func pipe(dst, src FullDuplexStream, rTimeout, wTimeout time.Duration) {
	if nil == dst || nil == src {
		closeStream(dst, src)
		log.Println("refused: invoid stream")
		return
	}

	if 0 < rTimeout {
		src.SetReadDeadline(time.Now().Add(rTimeout))
	}
	if 0 < wTimeout {
		dst.SetWriteDeadline(time.Now().Add(wTimeout))
	}
	go sendStream("up", dst, src)
	go sendStream("dw", src, dst)
}
