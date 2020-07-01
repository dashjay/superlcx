package core

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"superlcx/cc"
	"superlcx/middlewares/c_header"
	"superlcx/middlewares/stdout"
	"superlcx/middlewares/sub_filter"
)

func organizeUrl(req *http.Request, defaultT *url.URL) {
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
	var target *url.URL = nil
	if cc.C.ProxyUrls != nil && len(cc.C.ProxyUrls) > 0 {
		for _, proxyUrl := range cc.C.ProxyUrls {
			if proxyUrl.Re.MatchString(req.URL.RequestURI()) {
				target = proxyUrl.U
				break
			}
		}
	}
	if target == nil {
		target = defaultT
	}

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

type middleware struct {
	reqHandlers  []func(req *http.Request)
	respHandlers []func(resp *http.Response)
}

func newMiddleware(mid string) *middleware {
	middle := &middleware{
		reqHandlers:  []func(req *http.Request){},
		respHandlers: []func(resp *http.Response){},
	}
	if mid != "" {
		ms := strings.Split(mid, ",")
		for _, m := range ms {
			log.Printf("try load [%s] middleware.", m)
			switch strings.TrimSpace(m) {
			case "stdout":
				middle.RegisterMiddleware(stdout.HandleRequest, stdout.HandleResponse)
			case "c_header":
				middle.RegisterMiddleware(c_header.HandleRequest, c_header.HandleResponse)
			case "sub_filter":
				middle.RegisterMiddleware(sub_filter.HandleRequest, sub_filter.HandleResponse)
			default:
				reqH, respH := find(m)
				middle.RegisterMiddleware(reqH, respH)
			}
		}
	}
	return middle
}

func (m *middleware) RegisterMiddleware(reqH func(req *http.Request), respH func(resp *http.Response)) {
	m.reqHandlers = append(m.reqHandlers, reqH)
	m.respHandlers = append(m.respHandlers, respH)
}
