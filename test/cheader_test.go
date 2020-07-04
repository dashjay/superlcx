package test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"superlcx/cc"
)

func TestCHeader(t *testing.T) {
	cc.Config.Middleware = "c_header"
	cc.Config.CustomHeaders = map[string]cc.CustomHeader{
		"req1":  {Key: req1K, Value: req1V},
		"resp1": {Key: resp1K, Value: resp1V},
	}
	modeWithMiddleware := []string{"blend", "proxy"}
	for _, m := range modeWithMiddleware {
		cc.Config.Mode = m
		ctx, cancel := context.WithCancel(context.Background())
		go testMode(ctx)
		wait()
		req, err := http.NewRequest("GET", fmt.Sprintf("http://0.0.0.0:%d%s", cc.Config.ListenPort, cHeader), strings.NewReader("req"))
		if err != nil {
			panic(err)
		}
		resp, err := hc.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.Header.Get(resp1K) != resp1V {
			t.Fatalf("[x] c_header response test tail [%s] != [%s]", resp.Header.Get(resp1K), resp1V)
		}
		t.Logf("[√] c_header response test success [%s] == [%s]", resp.Header.Get(resp1K), resp1V)
		cancel()
		wait()
	}
}
