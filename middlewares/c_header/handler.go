package c_header

import (
	"net/http"
	"strings"

	. "superlcx/cc"
)

func HandleRequest(req *http.Request) {
	if Config.CustomHeaders == nil || len(Config.CustomHeaders) == 0 {
		return
	}
	for k, v := range Config.CustomHeaders {
		if strings.HasPrefix(k, "req") {
			// log.Printf("add header kv to req k:[%s],v:[%s]", v.Key, v.Value)
			req.Header.Set(v.Key, v.Value)
		}
	}
}

func HandleResponse(resp *http.Response) {
	if Config.CustomHeaders == nil || len(Config.CustomHeaders) == 0 {
		return
	}
	for k, v := range Config.CustomHeaders {
		if strings.HasPrefix(k, "resp") {
			// log.Printf("add header kv to resp k:[%s],v:[%s]", v.Key, v.Value)
			resp.Header.Set(v.Key, v.Value)
		}
	}
}
