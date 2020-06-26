package core

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Proxy struct {
	defaultUrl *url.URL
	lis        net.Listener
}

// NewSapProxy 构建一个SapProxy
func NewSapProxy(lis net.Listener, defaultHost string) *Proxy {
	u, err := url.Parse(fmt.Sprintf("http://%s", defaultHost))
	if err != nil {
		panic(err)
	}
	log.Printf("parse default url as %s", u)
	return &Proxy{
		defaultUrl: u,
		lis:        lis,
	}
}

func (s *Proxy) director(req *http.Request) {
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
	target := s.defaultUrl
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
	s *Proxy
}

func (t *myTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {

	resp, err = t.RoundTripper.RoundTrip(req)
	if err != nil {
		log.Printf("[x] default roundTrip error, detail: %s", err)
		return nil, err
	}

	return resp, nil
}
func (s *Proxy) modifyResponse(r *http.Response) error {
	return nil
}

func (s *Proxy) Serve() {
	p := &httputil.ReverseProxy{
		Director:       s.director,
		Transport:      &myTripper{RoundTripper: http.DefaultTransport, s: s},
		ModifyResponse: s.modifyResponse,
	}
	panic(http.Serve(s.lis, p))
}
