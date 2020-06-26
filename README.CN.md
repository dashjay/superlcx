[EN](./README.md)[中文]

# superlcx
支持中间件的端口转发工具

# 用法
```bash
~/superlcx(master*) » ./superlcx -h                                                                                              dashjay@zhaowenjies-MacBook-Pro
Usage of ./superlcx:
  -host string
        target host:port (default "0.0.0.0:8081")
  -l int
        listen port (default 8080)
  -m string
        run mode (default "proxy")
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
