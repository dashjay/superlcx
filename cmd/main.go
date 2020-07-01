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

	. "superlcx/cc"
	"superlcx/core"
)

const version = "1.0.2"

var (
	showVersion bool
	configFile  string
)

func init() {
	defer func() {
		err := recover()
		if err != nil {
			flag.PrintDefaults()
			os.Exit(-1)
		}
	}()
	flag.BoolVar(&showVersion, "v", false, "show version and about then exit.")
	flag.StringVar(&configFile, "c", "", "load config from")
	flag.IntVar(&C.ListenPort, "l", 8080, "listen port")
	flag.IntVar(&C.PPROFPort, "pp", 8999, "pprof port")
	flag.StringVar(&C.DefaultTarget, "host", "0.0.0.0:8081", "target host:port.")
	flag.StringVar(&C.Mode, "m", "proxy", "run mode <proxy|copy|blend>.")
	flag.StringVar(&C.Middleware, "M", "", "middleware, comma separated if more than one, eg: --M stdout,dumps")
	flag.StringVar(&C.LogFlag, "log", "t", "l -> line of code, d -> date, t -> time, order doesn't matter")
	flag.Parse()
	if showVersion {
		fmt.Printf(`
  _____ _    _ _____  ______ _____  _      _______   __
 / ____| |  | |  __ \|  ____|  __ \| |    / ____\ \ / /
| (___ | |  | | |__) | |__  | |__) | |   | |     \ V / 
 \___ \| |  | |  ___/|  __| |  _  /| |   | |      > <  
 ____) | |__| | |    | |____| | \ \| |___| |____ / . \ 
|_____/ \____/|_|    |______|_|  \_\______\_____/_/ \_\

Superlcx [%s], a tool kit for port transfer with middlewares!
`, version)
		os.Exit(0)
	}
	if configFile != "" {
		err := C.InitConfig(configFile)
		if err != nil {
			panic(err)
		}
	}
	C.SetLogFlag()
	// print config after flag was set
	C.Print()

	if C.ListenPort < 1 || C.ListenPort > 65535 {
		panic("[x] Listen Port Invalid")
	}
	checkHost(C.DefaultTarget)
}

func main() {
	// Buried point for debug
	go func() {
		http.ListenAndServe(fmt.Sprintf(":%d", C.PPROFPort), nil)
	}()

	go showMemLog()

	// start listen
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", C.ListenPort))
	if err != nil {
		panic(err)
	}
	switch C.Mode {
	case "proxy":
		c := core.NewSapProxy(lis, C)
		c.Serve()
	case "copy":
		c := core.NewSapCopy(lis, C)
		c.Serve()
	case "blend":
		c := core.NewSapBlend(lis, C)
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
