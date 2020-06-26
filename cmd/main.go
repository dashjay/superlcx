package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"superlcx/core"
)

const version = "1.0.0"

var (
	listenPort int
	hostPort   string
	mode       string
)

func init() {
	defer func() {
		err := recover()
		if err != nil {
			flag.PrintDefaults()
			os.Exit(-1)
		}
	}()
	flag.IntVar(&listenPort, "l", 8080, "listen port")
	flag.StringVar(&hostPort, "host", "0.0.0.0:8081", "target host:port")
	flag.StringVar(&mode, "m", "proxy", "run mode")
	flag.Parse()
	if listenPort < 1 || listenPort > 65535 {
		panic("[x] Listen Port Invalid")
	}
	checkHost(hostPort)
}

func main() {
	// Buried point for debug
	go func() {
		http.ListenAndServe(":8999", nil)
	}()

	go showMemLog()

	// start listen
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", listenPort))
	if err != nil {
		panic(err)
	}
	switch mode {
	case "proxy":
		c := core.NewSapProxy(lis, hostPort)
		c.Serve()
	case "copy":
		c := core.NewSapCopy(lis, hostPort)
		c.Serve()
	default:
		flag.PrintDefaults()
		os.Exit(-1)
	}
}

// checkHost check the ip:port valid
func checkHost(host string) {
	const pattern = `^(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])$`
	ipPort := strings.Split(host, ":")
	if len(ipPort) != 2 {
		panic("host should like this ip:port ")
	}
	port, err := strconv.Atoi(ipPort[1])
	if err != nil {
		panic(err)
	}
	if port < 1 || port > 65535 {
		panic(fmt.Sprintf("host port %d invalid", port))
	}
	ok, err := regexp.MatchString(pattern, ipPort[0])
	if err != nil || !ok {
		panic(fmt.Sprintf("host ip %s invalid", ipPort[0]))
	}
}

func showMemLog() {
	ticker := time.NewTicker(20 * time.Second)
	for range ticker.C {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		log.Printf("Memory: Alloc = %vMb TotalAlloc = %vMb Sys = %vMb NumGC = %v", m.Alloc/1024/1024, m.TotalAlloc/1024/1024, m.Sys/1024/1024, m.NumGC)
	}
}