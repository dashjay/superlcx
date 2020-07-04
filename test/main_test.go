package test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"superlcx/cc"
)

func TestSuperLcx(t *testing.T) {
	t.Log("start proxy test")
	for _, m := range mode {
		cc.Config.Mode = m
		ctx, cancel := context.WithCancel(context.Background())
		go testMode(ctx)
		wait()
		for _, s := range []string{"long", "short"} {
			req, err := http.NewRequest("POST", fmt.Sprintf("http://0.0.0.0:%d/%s", cc.Config.ListenPort, s), strings.NewReader("req"))
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
		wait()
	}
}

func BenchmarkSuperLcx(b *testing.B) {
	b.Log("start proxy bench")
	for _, m := range mode {
		for _, s := range []string{"long", "short"} {
			b.Run(m+" "+s, func(b *testing.B) {
				cc.Config.Mode = m
				ctx, cancel := context.WithCancel(context.Background())
				go testMode(ctx)
				wait()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					req, err := http.NewRequest("POST", fmt.Sprintf("http://0.0.0.0:%d/%s", cc.Config.ListenPort, s), strings.NewReader("req"))
					if err != nil {
						panic(err)
					}
					req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					resp, err := hc.Do(req)
					if err != nil {
						b.Fatal(err)
					}
					_, err = ioutil.ReadAll(resp.Body)
					if err != nil {
						b.Fatal(err)
					}
					resp.Body.Close()
				}
				cancel()
				b.StopTimer()
				wait()
			})
		}
	}
}
