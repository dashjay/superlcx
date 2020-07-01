package stdout

import (
	"net/http"
	"net/http/httputil"
	"os"
)

func HandleRequest(req *http.Request) {
	over := make(chan bool)
	go handleRequest(over, req)
	<-over
}

func handleRequest(over chan bool, req *http.Request) {
	content, err := httputil.DumpRequest(req, true)
	over <- true
	if err != nil {
		panic(err)
	}
	os.Stdout.Write(content)
	os.Stdout.Write([]byte("\n"))
}

func HandleResponse(resp *http.Response) {
	over := make(chan bool)
	go handlerResponse(over, resp)
	<-over
}

func handlerResponse(over chan bool, resp *http.Response) {
	content, err := httputil.DumpResponse(resp, true)
	over <- true
	if err != nil {
		panic(err)
	}
	os.Stdout.Write(content)
	os.Stdout.Write([]byte("\n"))
}
