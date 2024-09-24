package agency_test

import (
	"agency"
	"fmt"
	"log"
	"os"

	"github.com/watsonserve/goutils"
)

func ExampleTest() {
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
		Desc:      "ca cert",
	}})

	if len(vals) < 2 {
		fmt.Fprintln(os.Stderr, "params error")
		return
	}
	log.Println(opts)

	p, err := agency.New(opts["crt"], opts["key"], opts["ca"], 65536, 8, 4)
	if nil != err {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	p.ListenAndServe(vals[0], vals[1])
}
