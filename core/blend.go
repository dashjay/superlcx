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
	"strings"
	"time"

	"superlcx/cc"
)

const maxReadTime = 10 * time.Second

type SapBlend struct {
	lis        net.Listener
	defaultUrl *url.URL
	*middleware
}

func NewSapBlend() *SapBlend {
	// start listen
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cc.Config.ListenPort))
	if err != nil {
		panic(err)
	}
	log.Printf("[+] superlcx listen at [%d]", cc.Config.ListenPort)
	defaultUrl, err := url.Parse(fmt.Sprintf("http://%s/", cc.Config.DefaultTarget))
	if err != nil {
		panic(fmt.Sprintf("default url [%s] parse error, detail:[%s]", defaultUrl, err))
	}
	log.Printf("parse default url as %s", defaultUrl)
	b := &SapBlend{lis: lis, defaultUrl: defaultUrl,
		middleware: newMiddleware(cc.Config.Middleware)}
	return b
}

func (s *SapBlend) Serve(ctx context.Context) {
	log.Printf("superlcx work in blend mode!")
	tr := http.DefaultTransport
	go func() {
		<-ctx.Done()
		s.lis.Close()
	}()
	for {
		conn, err := s.lis.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "closed network connection") {
				return
			}
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

				if len(s.reqHandlers) > 0 {
					for _, reqH := range s.reqHandlers {
						reqH(newReq)
					}
				}
				resp, err := tr.RoundTrip(newReq)
				if err != nil {
					log.Printf("[x] default transport req error, detail:[%s]", err)
					continue
				}
				if len(s.respHandlers) > 0 {
					for _, respH := range s.respHandlers {
						respH(resp)
					}
				}
				resp.Write(conn)
				cancel()
			}
		}()
	}
}
