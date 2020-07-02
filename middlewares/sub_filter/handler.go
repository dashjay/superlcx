package sub_filter

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	. "superlcx/cc"
)

func HandleRequest(req *http.Request) {

}

func HandleResponse(resp *http.Response) {
	if Config.SubFilters == nil || len(Config.SubFilters) == 0 {
		return
	}
	ruri := resp.Request.URL.RequestURI()
	log.Printf("check invoke subFilter on url [%s]", ruri)
	var mk []SubFilter
	if Config.SubFilters != nil && len(Config.SubFilters) != 0 {
		for sub := range Config.SubFilters {
			f := Config.SubFilters[sub]
			if f.RUriMatcher.MatchString(ruri) {
				mk = append(mk, f)
			}
		}
	}
	if len(mk) == 0 {
		return
	}
	log.Printf("start invoke subFilter on url [%s] with mks [%v]", ruri, mk)

	buf := bufio.NewReader(resp.Body)
	var tempBuf bytes.Buffer
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			continue
		}
		for i := 0; i < len(mk); i++ {
			s := mk[i]
			if s.OldMatcher.Match(line) {
				log.Printf("subFilter [%s] matched [%s]", s.OldMatcher, line)
				line = s.HandleLine(line)
				log.Printf("subFilter after handle res: [%s]", line)
			}
		}
		tempBuf.Write(line)
	}
	_, err := tempBuf.ReadFrom(buf)
	if err != nil {
		log.Printf("sub_filter request body error, detail:[%s]", err)
	}
	n := tempBuf.Len()
	resp.Body = ioutil.NopCloser(&tempBuf)

	resp.ContentLength = int64(n)
	resp.Header.Set("Content-Length", strconv.Itoa(n))
}
