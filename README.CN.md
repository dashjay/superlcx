```bash
  _____ _    _ _____  ______ _____  _      _______   __
 / ____| |  | |  __ \|  ____|  __ \| |    / ____\ \ / /
| (___ | |  | | |__) | |__  | |__) | |   | |     \ V / 
 \___ \| |  | |  ___/|  __| |  _  /| |   | |      > <  
 ____) | |__| | |    | |____| | \ \| |___| |____ / . \ 
|_____/ \____/|_|    |______|_|  \_\______\_____/_/ \_\
```

<p align="center">
<a href="https://travis-ci.com/github/dashjay/superlcx"><img src="https://travis-ci.com/dashjay/superlcx.svg?branch=master" alt="Build Status"></a>
</p>

[EN](./README.md)[中文]

# 介绍
一个高性能的工具，具有丰富的中间件。
SuperLcx为请求做代理，并返回服务端的返回值，过程中调用一些中间件，来实现比较高级的操作。

# 用法
```bash
Usage of ./superlcx:
  -M string
        middleware, comma separated if more than one, eg: --M stdout,dumps
  -c string
        load config from
  -host string
        target host:port. (default "0.0.0.0:8081")
  -l int
        listen port (default 8080)
  -log string
        l -> line of code, d -> date, t -> time, order doesn't matter (default "t")
  -m string
        run mode <proxy|copy|blend>. (default "proxy")
  -pp int
        pprof port (default 8999)
  -v    show version and about then exit.
```

### -m 工作模式
- 代理模式 proxy
    - 优点: 代理模式基于http包构建，能够提供（暴露）更多的API例如，`modifyResponse(修改返回体)`,`Transport(http核心)`,`director(引导)，可根据路由进行转发等等类似proxy_pass`等。
    - 缺点: 代理模式可能会造成内存抖动，引起大量gc,不适合内存很小的机器，如果你需要很高的性能（并发量）。

- 拷贝模式 copy
    - 优点: 拷贝模式直接工作在TCP层, 他并不关心转发的是什么. 在 `io.Copy`的帮助下，运行只需要非常少的内存。
    - 缺点: 拷贝模式并不知道应用层的内容. 所有他可以做的仅仅是全部dump出来，做其他操作。

- 融合模式 blend
    - 优点：混合模式虽然工作在TCP层，但是以牺牲解析HTTP为代价，为请求和返回值的分析提供了中间间模式运行的可能性。
    - 缺点：因为IO量可能比较大，可能会引起内存抖动，大量gc，在高并发下不适合小内存的机器。

### -c 配置读取
> 配置的读取采用了一个golang的[toml库](https://github.com/BurntSushi/toml)

⚠️可能会如果使用-c读取配置文件后，请讲全部配置全部写入配置文件，不要部分依赖命令行参数，不清楚toml库的内部实现，可能出现覆盖的情况。

配置文件的文档请查看：[配置文件说明](./docs/config.CN.md)

### -M 中间件
当工作在代理模式下，可以调用中间件来对过程中的流量进行分析。例如，系统内置stdout中间件（示例中间件），可以通过`-M stdout`来使用。
（必须在代理模式和混合模式下才能生效）

如果想自己实现一个中间件请查看：
[中间件的编写规范](./docs/middleware.CN.md)

编写后可以放入middlewares文件夹下，建议文件结构保持如下。
```
middlewares
└── stdout
    └── handler.go
```

接口在`handler.go`下分别暴露如下，如需载入配置，请自行载入，或者在公共cc库中编写。
```
func HandleRequest(req *http.Request)
func HandleResponse(req *http.Response)
```

**警告：** 中间件组件通常会在req和resp上做一些事情，这必然会对bufio做很多io，内存抖动可能非常严重。
我考虑过编写一个中心中间件来转储req和resp，然后流式调用所有中间件，比如pipeline，但是如果其中一些更改了req或resp，也会导致其他问题，所以现在中间件组件不会相互影响。(你可以把上下文设置到req体的header里，这样可以与之后的中间件进行一些交互)