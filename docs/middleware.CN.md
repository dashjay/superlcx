## 中间件的编写
[EN](./middleware.md)[中文]

中间件基于请求体(http.Request)和返回体(http.Response)来编写，由于两个结构都包含一些buf，例如请求和返回体的body.

因此为了防止body被中间件读取后，客户端或者服务端无法收到真实的请求体和返回体，我们需要做一些操作。

示例的中间件stdout中这样写到：
```gotemplate
func handleRequest(over chan bool, req *http.Request) {
	content, err := httputil.DumpRequest(req, true)
	over <- true
	if err != nil {
		panic(err)
	}
...
```
`httputil.DumpRequest`会将请求body读出（如果存在的话），并且将body还给req这个请求体一份。如果我们自己写中间件也应该这样来写。

如果你不能很好的处理请求和返回中body，可能会造成严重的后果：
- body为空，读时EOF error
- body已经close，读时panic

为了避免程序因为这些原因崩溃，建议使用如下方式来编写中间件。

```
content, err := httputil.DumpRequest(req, true)
newReq, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(content)))
```

这样可以完全避免req请求体收到影响，能够正常发送到服务端以收到回复。