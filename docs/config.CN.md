## 配置文件说明
[EN](./config.md)[中文]

> toml的注释为#开头，建议在toml文件中写注释方便他人阅读，因为国际化，我不在我的默认配置文件中加入任何语言的注释。


```toml
ListenPort = 8081 # 监听的端口
DefaultTarget = "0.0.0.0:8080" # 默认的目标，如果没有经过任何中间件处理，请求将会发送到这里。
PPROFPort = 8999 # golang的调试端口

# 日志格式 
# l 代表文件行，会打印 blend.go:37: 在每条日志之前。
# t 代表时间 ，会打印 11:26:24 在每条日志之前。
# d 日期，会打印日期，在每条日志之前。
LogFlag = "ltd" 

# 中间件，用逗号分隔，具体请查看README中-M关于中间件的说明
Middleware = "c_header,stdout"

# 工作模式，具体请查看README中关于-m工作模式的说明
Mode = "blend"

# 自定义头部
# 会根据[CustomHeaders.req1]中的.req**或者.resp**来决定将此KV加在请求或是返回值。
# ⚠️：此操作，也会影响其他中间件的行为（如果按顺序执行），请小心使用此插件，避免造成错误的头引起其他操作失败。
[CustomHeaders]
    [CustomHeaders.req1] # req*** 指定次头部加给请求体的headr
    Key="X-REAL-IP"
    Value="111.111.111.111"

    [CustomHeaders.resp1] # resp*** 指定次头部加给返回体的headr
    Key="Server"
    Value="ASP.NET"

# 代理Url
# 会按照Path（正则匹配），对匹配的路由进行转发例如
# 当请求到/superlcx/xxx的时候，会将次请求转发给指定Host
[ProxyUrls]
    [ProxyUrls.portrait]
    Scheme="http"
    Host="0.0.0.0:8989"
    Path="/superlcx/*"
```
