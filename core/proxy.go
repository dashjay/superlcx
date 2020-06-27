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

	"superlcx/middlewares/stdout"
)

type Proxy struct {
	defaultUrl   *url.URL
	lis          net.Listener
	reqHandlers  []func(req *http.Request)
	respHandlers []func(resp *http.Response)
}

func (p *Proxy) RegisterMiddleware(reqH func(req *http.Request), respH func(resp *http.Response)) {
	p.reqHandlers = append(p.reqHandlers, reqH)
	p.respHandlers = append(p.respHandlers, respH)
}

// NewSapProxy 构建一个SapProxy
func NewSapProxy(lis net.Listener, defaultHost string, middleware string) *Proxy {
	u, err := url.Parse(fmt.Sprintf("http://%s", defaultHost))
	if err != nil {
		panic(err)
	}
	log.Printf("parse default url as %s", u)
	p := &Proxy{
		defaultUrl:   u,
		lis:          lis,
		reqHandlers:  make([]func(req *http.Request), 0),
		respHandlers: make([]func(resp *http.Response), 0),
	}
	p.Register(middleware)
	return p
}

func (p *Proxy) director(req *http.Request) {
	for _, fn := range p.reqHandlers {
		fn(req)
	}
	if _, ok := req.Header["User-Agent"]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		req.Header.Set("User-Agent", "")
	}
	singleJoiningSlash := func(a, b string) string {
		aslash := strings.HasSuffix(a, "/")
		bslash := strings.HasPrefix(b, "/")
		switch {
		case aslash && bslash:
			return a + b[1:]
		case !aslash && !bslash:
			return a + "/" + b
		}
		return a + b
	}
	target := p.defaultUrl
	targetQuery := target.RawQuery
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
	if targetQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = targetQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
	}
}

type myTripper struct {
	http.RoundTripper
	p *Proxy
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
func (p *Proxy) modifyResponse(r *http.Response) error {
	return nil
}

func (p *Proxy) Serve() {
	proxy := &httputil.ReverseProxy{
		Director:       p.director,
		Transport:      &myTripper{RoundTripper: http.DefaultTransport, p: p},
		ModifyResponse: p.modifyResponse,
	}
	panic(http.Serve(p.lis, proxy))
}

func (p *Proxy) Register(middleware string) {
	if middleware != "" {
		ms := strings.Split(middleware, ",")
		for _, m := range ms {
			switch m {
			case "stdout":
				p.RegisterMiddleware(stdout.HandleRequest, stdout.HandleResponse)
			default:
				reqH, respH := find(m)
				p.RegisterMiddleware(reqH, respH)
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
