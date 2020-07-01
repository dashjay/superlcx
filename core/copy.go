package core

import (
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

func (s *SapCopy) Serve() {
	log.Printf("superlcx work in copy mode!")
	for {
		conn, err := s.lis.Accept()
		if err != nil {
			log.Printf("[x] accept error, detail: [%s]", conn)
			return
		}
		conn2, err := net.Dial("tcp", s.target)
		if err != nil {
			log.Printf("[x] connect to target error, detail: [%s]", err)
			return
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

func NewSapCopy(lis net.Listener, cfg cc.Cfg) *SapCopy {
	return &SapCopy{lis: lis, target: cfg.DefaultTarget}
}
