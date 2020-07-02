package core

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"

	"superlcx/cc"
	"superlcx/middlewares/c_header"
	"superlcx/middlewares/js_lua"
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
	if cc.Config.ProxyUrls != nil && len(cc.Config.ProxyUrls) > 0 {
		for _, proxyUrl := range cc.Config.ProxyUrls {
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

	req.Header.Add(cc.UNIQUEID, uuid.New().String())
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
			case "js_lua":
				middle.RegisterMiddleware(js_lua.HandleRequest, js_lua.HandleResponse)
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
