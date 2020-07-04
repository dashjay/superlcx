package core

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"superlcx/cc"
)

type SapProxy struct {
	defaultUrl *url.URL
	lis        net.Listener
	*middleware
}

// NewSapProxy 构建一个SapProxy
func NewSapProxy() *SapProxy {
	// start listen
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cc.Config.ListenPort))
	if err != nil {
		panic(err)
	}
	log.Printf("[+] superlcx listen at [%d]", cc.Config.ListenPort)
	u, err := url.Parse(fmt.Sprintf("http://%s", cc.Config.DefaultTarget))
	if err != nil {
		panic(err)
	}
	log.Printf("parse default url as %s", u)
	p := &SapProxy{
		defaultUrl: u,
		lis:        lis,
		middleware: newMiddleware(cc.Config.Middleware),
	}
	return p
}

func (s *SapProxy) director(req *http.Request) {
	organizeUrl(req, s.defaultUrl)
	for _, fn := range s.reqHandlers {
		fn(req)
	}
}

type myTripper struct {
	http.RoundTripper
	p *SapProxy
}

func (t *myTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {

	resp, err = t.RoundTripper.RoundTrip(req)
	if err != nil {
		log.Printf("[x] default roundTrip error, detail: %s", err)
		return nil, err
	}

	for _, fn := range t.p.respHandlers {
		fn(resp)
	}

	return resp, nil
}
func (s *SapProxy) modifyResponse(r *http.Response) error {
	return nil
}

func (s *SapProxy) Serve(ctx context.Context) {
	log.Printf("superlcx work in proxy mode!")
	proxy := &httputil.ReverseProxy{
		Director:       s.director,
		Transport:      &myTripper{RoundTripper: http.DefaultTransport, p: s},
		ModifyResponse: s.modifyResponse,
	}
	go http.Serve(s.lis, proxy)
	<-ctx.Done()
	s.lis.Close()
}
