package core

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"superlcx/cc"
)

type SapCopy struct {
	lis    net.Listener
	target string
}

func (s *SapCopy) Serve(ctx context.Context) {
	log.Printf("superlcx work in copy mode!")
	go func() {
		<-ctx.Done()
		s.lis.Close()
	}()
	for {
		conn, err := s.lis.Accept()
		if err != nil {
			log.Printf("[x] accept error, detail: [%s]", err)
			return
		}
		conn2, err := net.Dial("tcp", s.target)
		if err != nil {
			log.Printf("[x] connect to target error, detail: [%s]", err)
			conn.Close()
			continue
		}
		log.Printf("[+] transfer between [%s] <-> [%s]", conn.LocalAddr(), conn2.RemoteAddr())
		go func() {
			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				io.Copy(conn2, conn)
				wg.Done()
			}()
			go func() {
				io.Copy(conn, conn2)
				wg.Done()
			}()
			wg.Wait()
			conn.Close()
			conn2.Close()
		}()
	}
}

func NewSapCopy() *SapCopy {
	// start listen
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cc.Config.ListenPort))
	if err != nil {
		panic(err)
	}
	log.Printf("[+] superlcx listen at [%d]", cc.Config.ListenPort)
	return &SapCopy{lis: lis, target: cc.Config.DefaultTarget}
}
