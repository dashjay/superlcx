package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	. "superlcx/cc"
	"superlcx/core"
)

const version = "1.0.5"

var (
	showVersion bool
	configFile  string
)

func init() {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Print(err)
			flag.PrintDefaults()
			os.Exit(-1)
		}
	}()
	flag.BoolVar(&showVersion, "v", false, "show version and about then exit.")
	flag.StringVar(&configFile, "c", "", "load config from")
	flag.IntVar(&Config.ListenPort, "l", 8080, "listen port")
	flag.IntVar(&Config.PPROFPort, "pp", 8999, "pprof port")
	flag.StringVar(&Config.DefaultTarget, "host", "0.0.0.0:8081", "target host:port.")
	flag.StringVar(&Config.Mode, "m", "proxy", "run mode <proxy|copy|blend>.")
	flag.StringVar(&Config.Middleware, "M", "", "middleware, comma separated if more than one, eg: --M stdout,dumps")
	flag.StringVar(&Config.LogFlag, "log", "t", "l -> line of code, d -> date, t -> time, order doesn't matter")
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
		err := Config.InitConfig(configFile)
		if err != nil {
			panic(err)
		}
	}

	if Config.ListenPort < 1 || Config.ListenPort > 65535 {
		panic("[x] Listen Port Invalid")
	}
	checkHost(Config.DefaultTarget)
}

func main() {
	// Buried point for debug
	go http.ListenAndServe(fmt.Sprintf(":%d", Config.PPROFPort), nil)
	go showMemLog()
	ctx, cancel := context.WithCancel(context.Background())

	var (
		// sigs, when anything unexpected happened, a signal will send to
		// this chan. then server start to stop.
		sigs = make(chan os.Signal, 1)
		// when everything closed, a signal will send to done. the main Goroutine then stop.
		done = make(chan bool, 1)
	)
	go func() {
		switch Config.Mode {
		case "proxy":
			c := core.NewSapProxy(Config)
			c.Serve(ctx)
		case "copy":
			c := core.NewSapCopy(Config)
			c.Serve(ctx)
		case "blend":
			c := core.NewSapBlend(Config)
			c.Serve(ctx)
		default:
			flag.PrintDefaults()
			sigs <- os.Interrupt
		}
	}()
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
		done <- true
	}()
	<-done
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
