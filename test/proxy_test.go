package test

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"testing"
	"time"

	"superlcx/cc"
	"superlcx/core"
)

var (
	shortResponse       []byte
	shortResponseLength = 1 << 10
	longResponse        []byte
	longResponseLength  = 1 << 20
	listenPort          = rand.Int31n(60000)
	testPort            = rand.Int31n(60000)
)
var (
	mode = strings.Split(cc.ALLMode, ",")
	hc   http.Client
)

func init() {
	if listenPort < 10000 {
		listenPort += 10000
	}
	if testPort < 10000 {
		testPort += 10000
	}
	if listenPort == testPort {
		testPort++
	}
	var th h
	go func() {
		fmt.Printf("test server listen at port [%d]", testPort)
		err := http.ListenAndServe(fmt.Sprintf(":%d", testPort), &th)
		if err != nil {
			panic(err)
		}
	}()
	for i := 0; i < shortResponseLength; i++ {
		shortResponse = append(shortResponse, byte('s'))
	}
	for i := 0; i < longResponseLength; i++ {
		longResponse = append(longResponse, byte('s'))
	}
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
}

func TestSuperLcx(t *testing.T) {
	t.Log("start proxy test")
	var cfg cc.Cfg
	cfg.ListenPort = int(listenPort)
	cfg.DefaultTarget = fmt.Sprintf("0.0.0.0:%d", testPort)
	for _, m := range mode {
		cfg.Mode = m
		ctx, cancel := context.WithCancel(context.Background())
		go testMode(ctx, cfg)

		time.Sleep(1 * time.Second)
		for _, s := range []string{"long", "short"} {
			req, err := http.NewRequest("GET", fmt.Sprintf("http://0.0.0.0:%d/%s", listenPort, s), strings.NewReader("req"))
			if err != nil {
				panic(err)
			}

			resp, err := hc.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			content, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			resp.Body.Close()
			switch {
			case s == "long" && len(content) != (longResponseLength):
				t.Fatalf("long response error length %d != %d", len(content), longResponseLength)
			case s == "short" && len(content) != (shortResponseLength):
				t.Fatalf("short response error length %d != %d", len(content), shortResponseLength)
			}
		}
		cancel()
		time.Sleep(3 * time.Second)
	}
}

func BenchmarkSuperLcx(b *testing.B) {
	b.Log("start proxy bench")
	var cfg cc.Cfg
	cfg.ListenPort = int(listenPort)
	cfg.DefaultTarget = fmt.Sprintf("0.0.0.0:%d", testPort)
	for _, m := range mode {
		cfg.Mode = m
		ctx, cancel := context.WithCancel(context.Background())
		go testMode(ctx, cfg)

		time.Sleep(1 * time.Second)
		for _, s := range []string{"long", "short"} {
			b.Run(m+" "+s, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					req, err := http.NewRequest("GET", fmt.Sprintf("http://0.0.0.0:%d/%s", listenPort, s), strings.NewReader("req"))
					if err != nil {
						panic(err)
					}
					resp, err := hc.Do(req)
					if err != nil {
						b.Fatal(err)
					}
					content, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						b.Fatal(err)
					}
					resp.Body.Close()
					switch {
					case s == "long" && len(content) != (longResponseLength):
						b.Fatalf("long response error length %d != %d", len(content), longResponseLength)
					case s == "short" && len(content) != (shortResponseLength):
						b.Fatalf("short response error length %d != %d", len(content), shortResponseLength)
					}
				}
			})
		}
		cancel()
		time.Sleep(3 * time.Second)
	}
}
func testMode(ctx context.Context, cfg cc.Cfg) {

	switch cfg.Mode {
	case "proxy":
		p := core.NewSapProxy(cfg)
		p.Serve(ctx)
	case "blend":
		p := core.NewSapBlend(cfg)
		p.Serve(ctx)
	case "copy":
		p := core.NewSapCopy(cfg)
		p.Serve(ctx)
	default:
		panic("unknown mode" + cfg.Mode)
	}
}
