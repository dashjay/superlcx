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

## 内置中间件说明文档

### stdout
该中间件仅仅是将请求和返回dump出一份，输出到标准输出，作为一个示例模板的中间件。

### c_header
改中间件，可以根据配置文件，对请求和返回中的头进行添加或者修改，可根据[配置文件中的说明](./config.CN.md)，对config.toml进行编辑，达到你想要的效果

```toml
# c_header config
[CustomHeaders]
    [CustomHeaders.req1] # req开头->代表添加或者修改请求头
    Key="X-REAL-IP"
    Value="111.111.111.111"

    [CustomHeaders.resp1] # resp开头->代表添加或者修改返回头
    Key="Server"
    Value="ASP.NET"
```

### sub_filter
和nginx中的sub_filter功能完全相同，针对指定路径（支持正则匹配），对某指定路径请求的返回值进行修改，替换页面内容。

```toml
[SubFilter]
    [SubFilter.test] # 名字暂时无用
    Old="</head>" # 原来的
    Repl='<script src="/js/jquery.min.js"></script></head>' # 替换为
    Path="/" # 指定url（支持正则表达式）
```
