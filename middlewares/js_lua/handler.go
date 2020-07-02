package js_lua

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/robertkrimen/otto"
	lua "github.com/yuin/gopher-lua"

	"superlcx/cc"
)

var Vm = VM{Lua: nil, Js: nil,
	LMap: sync.Map{}, JMap: sync.Map{},
}

func init() {

	go func() {
		time.Sleep(1 * time.Second)
		if cc.Config.JsPath != "" && strings.Contains(cc.Config.Middleware, "js_lua") {
			log.Printf("load js vm from file %s", cc.Config.JsPath)
			Vm.J.Lock()
			defer Vm.J.Unlock()
			jscode, err := ioutil.ReadFile(cc.Config.JsPath)
			if err != nil {
				panic(err)
			}
			Vm.Js = otto.New()
			_, err = Vm.Js.Run(jscode)
			if err != nil {
				panic(err)
			}
		}

		if cc.Config.LuaPath != "" && strings.Contains(cc.Config.Middleware, "js_lua") {
			log.Printf("load lua vm from file %s", cc.Config.LuaPath)
			Vm.L.Lock()
			defer Vm.L.Unlock()
			Vm.Lua = lua.NewState()
			Vm.Lua.PreloadModule("utils", SkyUtils)
			err := Vm.Lua.DoFile(cc.Config.LuaPath)
			if err != nil {
				panic(err)
			}
		}
	}()
}

var emptyBody = []byte("")

func HandleRequest(req *http.Request) {
	over := make(chan bool)
	go handleRequest(over, req)
	<-over
}

func makeHeader(h http.Header) string {
	var buf strings.Builder
	for k, d := range h {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", k, strings.Join(d, ";")))
	}
	return buf.String()
}

func handleRequest(over chan bool, req *http.Request) {
	content, err := httputil.DumpRequest(req, true)
	over <- true
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(bytes.NewReader(content))
	newReq, err := http.ReadRequest(r)
	if err != nil {
		panic(err)
	}
	bodyLen := newReq.ContentLength
	body := make([]byte, bodyLen)
	n, err := r.Read(body)
	if err != nil || n <= 0 || len(body) > (1<<20) {
		body = emptyBody[:]
	}
	uniqueId := newReq.Header.Get(cc.UNIQUEID)
	newReq.Body.Close()
	header := makeHeader(req.Header)
	reqDic := map[string]string{
		cc.UNIQUEID: uniqueId,
		"method":    newReq.Method,
		"url":       newReq.URL.RequestURI(),
		"proto":     newReq.Proto,
		"header":    header,
		"body":      string(body),
	}
	var waitG sync.WaitGroup
	if Vm.Js != nil {
		go func(wg *sync.WaitGroup) {
			wg.Add(1)
			defer wg.Done()

			Vm.J.Lock()
			defer Vm.J.Unlock()
			reqJsa, err := Vm.Js.ToValue(reqDic)
			if err != nil {
				panic(err)
			}
			result, _ := Vm.Js.Call("on_http_request", nil, reqJsa)
			v, _ := result.Export()
			if res, ok := v.(bool); ok && res {
				Vm.JMap.Store(uniqueId, res)
			}
		}(&waitG)
	}

	if Vm.Lua != nil {
		go func(wg *sync.WaitGroup) {
			wg.Add(1)
			defer wg.Done()
			Vm.L.Lock()
			defer Vm.L.Unlock()
			reqTable := Vm.Lua.NewTable()
			for k, v := range reqDic {
				Vm.Lua.SetTable(reqTable, lua.LString(k), lua.LString(v))
			}
			if err := Vm.Lua.CallByParam(lua.P{
				Fn:      Vm.Lua.GetGlobal("on_http_request"),
				NRet:    1,
				Protect: true,
			}, reqTable); err != nil {
				panic(err)
			}
			ret := Vm.Lua.Get(-1)
			Vm.Lua.Pop(1)
			if ret.String() == "true" {
				Vm.JMap.Store(uniqueId, true)
			}
		}(&waitG)
	}
	waitG.Wait()
}

func HandleResponse(resp *http.Response) {
	over := make(chan bool)
	go handlerResponse(over, resp)
	<-over
}

func handlerResponse(over chan bool, resp *http.Response) {
	var buf bytes.Buffer
	n, err := buf.ReadFrom(resp.Body)
	if err != nil {
		log.Printf("[x] read body error, detail %s", err)
	}
	err = resp.Body.Close()
	if err != nil {
		log.Printf("[x] close body error, detail %s", err)
	}
	resp.ContentLength = n
	resp.Header.Set("Content-Length", strconv.Itoa(int(n)))
	resp.Body = ioutil.NopCloser(&buf)
	over <- true
	uniqueId := resp.Request.Header.Get(cc.UNIQUEID)
	body := buf.Bytes()
	header := makeHeader(resp.Header)
	respDic := map[string]string{
		cc.UNIQUEID: uniqueId,
		"status":    fmt.Sprintf("%d", resp.StatusCode),
		"header":    header,
		"body":      string(body),
	}

	var waitG sync.WaitGroup
	if Vm.Js != nil {
		go func(wg *sync.WaitGroup) {
			wg.Add(1)
			defer wg.Done()

			Vm.J.Lock()
			defer Vm.J.Unlock()
			respJsa, err := Vm.Js.ToValue(respDic)
			if err != nil {
				panic(err)
			}
			_, _ = Vm.Js.Call("on_http_response", nil, respJsa)
		}(&waitG)
	}

	if Vm.Lua != nil {
		go func(wg *sync.WaitGroup) {
			wg.Add(1)
			defer wg.Done()
			Vm.L.Lock()
			defer Vm.L.Unlock()
			respTable := Vm.Lua.NewTable()
			for k, v := range respDic {
				Vm.Lua.SetTable(respTable, lua.LString(k), lua.LString(v))
			}
			if err := Vm.Lua.CallByParam(lua.P{
				Fn:      Vm.Lua.GetGlobal("on_http_response"),
				NRet:    0,
				Protect: true,
			}, respTable); err != nil {
				panic(err)
			}
		}(&waitG)
	}
	waitG.Wait()
}

type VM struct {
	Lua  *lua.LState
	L    sync.Mutex
	LMap sync.Map

	Js   *otto.Otto
	J    sync.Mutex
	JMap sync.Map
}

func SkyUtils(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

var exports = map[string]lua.LGFunction{
	"decode": decode,
}

func decode(L *lua.LState) int {
	r := L.Get(-1)
	L.Pop(1)
	btable := make(map[string]interface{})
	err := json.Unmarshal([]byte(r.String()), &btable)
	if err != nil {
		panic(err)
	}
	table := L.NewTable()
	for k, v := range btable {
		L.SetTable(table, lua.LString(k), lua.LString(fmt.Sprintf("%s", v)))
	}
	L.Push(table)
	return 1
}
