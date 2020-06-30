package core

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

const maxReadTime = 10 * time.Second

type SapBlend struct {
	lis        net.Listener
	defaultUrl *url.URL
	*middleware
}

func NewSapBlend(lis net.Listener, target string, middleware string) *SapBlend {
	defaultUrl, err := url.Parse(fmt.Sprintf("http://%s/", target))
	if err != nil {
		panic(fmt.Sprintf("default url [%s] parse error, detail:[%s]", defaultUrl, err))
	}
	log.Printf("parse default url as %s", defaultUrl)
	b := &SapBlend{lis: lis, defaultUrl: defaultUrl, middleware: newMiddleware(middleware)}

	return b
}

func (s *SapBlend) Serve() {
	log.Printf("superlcx work in blend mode!")
	tr := http.DefaultTransport
	for {
		conn, err := s.lis.Accept()
		if err != nil {
			log.Printf("[x] listener accept error, detail:[%s]", err)
		}
		go func() {
			buf := bufio.NewReader(conn)
			conn.SetDeadline(time.Now().Add(maxReadTime))
			defer conn.Close()
			wait := 100 * time.Microsecond
			for {
				ctx, cancel := context.WithCancel(context.Background())
				req, err := http.ReadRequest(buf)
				if err != nil {
					if err == io.EOF {
						time.Sleep(wait)
						wait *= 2
						continue
					} else {
						// 这里能识别已经关闭的链接
						log.Printf("[-] connection over, detail:[%s]", err)
						return
					}
				}
				newReq := req.Clone(ctx)
				organizeUrl(newReq, s.defaultUrl)

				for _, reqH := range s.reqHandlers {
					reqH(newReq)
				}
				resp, err := tr.RoundTrip(newReq)
				if err != nil {
					log.Printf("[x] default transport req error, detail:[%s]", err)
					continue
				}
				for _, respH := range s.respHandlers {
					respH(resp)
				}

				resp.Write(conn)
				cancel()
			}
		}()
	}
}
