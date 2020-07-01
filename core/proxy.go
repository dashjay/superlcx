package core

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"plugin"
	"strings"

	"superlcx/cc"
	"superlcx/middlewares/stdout"
)

type SapProxy struct {
	defaultUrl *url.URL
	lis        net.Listener
	middleware
}

// NewSapProxy 构建一个SapProxy
func NewSapProxy(lis net.Listener, cfg cc.Cfg) *SapProxy {
	u, err := url.Parse(fmt.Sprintf("http://%s", cfg.DefaultTarget))
	if err != nil {
		panic(err)
	}
	log.Printf("parse default url as %s", u)
	p := &SapProxy{
		defaultUrl: u,
		lis:        lis,
	}
	p.Register(cfg.Middleware)
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

func (s *SapProxy) Serve() {
	log.Printf("superlcx work in proxy mode!")
	proxy := &httputil.ReverseProxy{
		Director:       s.director,
		Transport:      &myTripper{RoundTripper: http.DefaultTransport, p: s},
		ModifyResponse: s.modifyResponse,
	}
	panic(http.Serve(s.lis, proxy))
}

func (s *SapProxy) Register(middleware string) {
	if middleware != "" {
		ms := strings.Split(middleware, ",")
		for _, m := range ms {
			switch m {
			case "stdout":
				s.RegisterMiddleware(stdout.HandleRequest, stdout.HandleResponse)
			default:
				reqH, respH := find(m)
				s.RegisterMiddleware(reqH, respH)
			}
		}
	}
}

func find(pluginName string) (func(req *http.Request), func(resp *http.Response)) {
	p, err := plugin.Open(pluginName)
	if err != nil {
		panic(err)
	}
	req, err := p.Lookup("HandleRequest")
	if err != nil {
		panic(err)
	}
	resp, err := p.Lookup("HandleResponse")
	if err != nil {
		panic(err)
	}
	return req.(func(req *http.Request)), resp.(func(resp *http.Response))
}
