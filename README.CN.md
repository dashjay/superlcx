[EN](./README.md)[中文]

# superlcx
支持中间件的端口转发工具

# 用法
```bash
~/superlcx(master) » ./superlcx -h                                                                                               dashjay@zhaowenjies-MacBook-Pro
Usage of ./superlcx:
  -M string
        middleware, comma separated if more than one, eg: --M stdout,dumps
  -host string
        target host:port (default "0.0.0.0:8081")
  -l int
        listen port (default 8080)
  -m string
        run mode (default "proxy")
  -v    show version and about then exit.
```

### 工作模式
- 代理模式 
    - 优点: 代理模式基于http包构建，能够提供（暴露）更多的API例如，`modifyResponse(修改返回体)`,`Transport(http核心)`,`director(引导)，可根据路由进行转发等等类似proxy_pass`等。
    - 缺点: 代理模式可能会造成内存抖动，引起大量gc,不适合内存很小的机器，如果你需要很高的性能（并发量）。

- 拷贝模式
    - 优点: 拷贝模式直接工作在TCP层, 他并不关心转发的是什么. 在 `io.Copy`的帮助下，运行只需要非常少的内存。
    - 缺点: 拷贝模式并不知道应用层的内容. 所有他可以做的仅仅是全部dump出来，做其他操作。

- 融合模式（开发中）
    - 期待优点：占用非常小的内存，提供丰富的接口例如`proxy_pass`等

### -M 中间件
当工作在代理模式下，可以调用中间件来对过程中的流量进行分析。例如，系统内置stdout中间件（示例中间件），可以通过`-M stdout`来使用。
（必须在代理模式下才能生效）

如果想自己实现一个中间件请查看：
[中间件的编写规范](./docs/middleware.CN.md)

编写后可以放入middlewares文件夹下，建议文件结构保持如下。
```
middlewares
└── stdout
    └── handler.go
```

接口在`handler.go`下分别暴露如下，如需载入配置，请自行载入，后期打算引入通用map来协助配置。
```
func HandleRequest(req *http.Request)
func HandleResponse(req *http.Response)
```

**警告：** 中间件组件通常会在req和resp上做一些事情，这必然会对bufio做很多io，内存抖动可能非常严重。
我考虑过编写一个中心中间件来转储req和resp，然后流式调用所有中间件，比如pipeline，但是如果其中一些更改了req或resp，也会导致其他问题，所以现在中间件组件不会相互影响。(你可以把上下文设置到req体的header里，这样可以与之后的中间件进行一些交互)