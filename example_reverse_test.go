package agency_test

import (
	"fmt"
	"net/http"
	"os"
	"time"

	quic "github.com/quic-go/quic-go"
	"github.com/watsonserve/agency"
	"github.com/watsonserve/goutils"
)

type Srv struct{}

func (s *Srv) ServeHTTP(resp http.ResponseWriter, req *http.Request) {

}

func TestExample() {
	opts, vals := goutils.GetOptions([]goutils.Option{{
		Name:      "crt",
		Opt:       'c',
		Option:    "crt",
		HasParams: true,
		Desc:      "cert",
	}, {
		Name:      "key",
		Opt:       'k',
		Option:    "key",
		HasParams: true,
		Desc:      "private key",
	}, {
		Name:      "ca",
		Opt:       'a',
		Option:    "ca",
		HasParams: true,
		Desc:      "ca crt",
	}})

	tlsCfg, err := goutils.GenTlsConfig(goutils.TLSFLAG_CLIENT, opts["crt"], opts["key"], opts["ca"])
	if nil != err {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	quicConf := &quic.Config{
		KeepAlivePeriod: time.Duration(10) * time.Second,
		MaxIdleTimeout:  time.Duration(30) * time.Second,
	}

	agency.ReverseServe(vals[0], tlsCfg, quicConf, &Srv{})
}
