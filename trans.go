package agency

import (
	"io"
	"log"
	"time"
)

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

func sendStream(load string, dst, src FullDuplexStream, bufSiz int) {
	written, err := io.CopyBuffer(dst, src, make([]byte, bufSiz))
	errMsg := "success"
	if nil != err {
		errMsg = "failed -- " + err.Error()
	}
	log.Printf("%s_result --length=%d --status=%s", load, written, errMsg)
}

func pipe(dst, src FullDuplexStream, bufSiz int, rTimeout, wTimeout time.Duration) {
	defer closeStream(dst, src)

	if nil == dst || nil == src {
		log.Println("refused: invoid stream")
		return
	}

	if 0 < rTimeout {
		src.SetReadDeadline(time.Now().Add(rTimeout))
	}
	if 0 < wTimeout {
		dst.SetWriteDeadline(time.Now().Add(wTimeout))
	}
	go sendStream("up", dst, src, bufSiz)
	sendStream("dw", src, dst, bufSiz)
}
