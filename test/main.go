package test

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"superlcx/cc"
	"superlcx/core"
)

func wait() {
	time.Sleep(1600 * time.Millisecond)
}

type h struct{}

func (i *h) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/long" {
		rw.Write(longResponse[:])
		return
	}
	if r.RequestURI == "/short" {
		rw.Write(shortResponse[:])
		return
	}
	if r.RequestURI == cHeader {
		if r.Header.Get(req1K) == req1V {
			log.Printf("[+] request key add success [%s] == [%s]", r.Header.Get(req1K), req1V)
			rw.Write(shortResponse[:])
			return
		} else {
			panic("[×] c_header request test fail [" + r.Header.Get(req1K) + "] != [" + req1V + "]")
		}
	}
}

var (
	shortResponse       []byte
	shortResponseLength = 1 << 10
	longResponse        []byte
	longResponseLength        = 1 << 20
	listenPort          int32 = 0
	testPort            int32 = 0
)

const (
	cHeader = "/c_header"
	req1K   = "User-Agent1"
	req1V   = "curl/*.*"
	resp1K  = "Server"
	resp1V  = "ASP.NET"
)

var (
	mode = strings.Split(cc.ALLMode, ",")
	hc   http.Client
)

func init() {
	rand.Seed(time.Now().Unix())
	testPort = rand.Int31n(60000)
	if testPort < 10000 {
		testPort += 10000
	}
	cc.Config.DefaultTarget = fmt.Sprintf("0.0.0.0:%d", testPort)

	for i := 0; i < shortResponseLength; i++ {
		shortResponse = append(shortResponse, byte('s'))
	}
	for i := 0; i < longResponseLength; i++ {
		longResponse = append(longResponse, byte('L'))
	}

	var th h
	go func() {
		fmt.Printf("test server listen at port [%d]\n", testPort)
		err := http.ListenAndServe(fmt.Sprintf(":%d", testPort), &th)
		if err != nil {
			panic(err)
		}
	}()
}
func testMode(ctx context.Context) {
	listenPort = rand.Int31n(60000)
	if listenPort < 10000 {
		listenPort += 10000
	}
	if listenPort == testPort {
		testPort++
	}
	cc.Config.ListenPort = int(listenPort)
	switch cc.Config.Mode {
	case "proxy":
		p := core.NewSapProxy()
		p.Serve(ctx)
	case "blend":
		p := core.NewSapBlend()
		p.Serve(ctx)
	case "copy":
		p := core.NewSapCopy()
		p.Serve(ctx)
	default:
		panic("unknown mode" + cc.Config.Mode)
	}
}
