## Configuration description

[EN][中文](./config.CN.md)
> TOML's comments start with #(hash)，It is recommended that comments in the TOML file, making it easy to read. Because of internationalization, I do not include comments in any language in my default configuration file.


```toml
ListenPort = 8081 # listen sock port
DefaultTarget = "0.0.0.0:8080" # The default destination, if not processed by any middleware, is where the request will be sent.
PPROFPort = 8999 # debug port for golang

# Log format 
# l for line of code, print eg (blend.go:37:) before every line.
# t for time, print eg (11:26:24) beforeevery line.
# d for date, print date eg (2020/07/01 ) before every line.
LogFlag = "ltd" 

# middlewares, comma separated, See the -M instructions on middleware in README for details
Middleware = "c_header,stdout"

# working mode, please refer to the README for details about -m working mode
Mode = "blend"

# Custom header
# The decision to add this KV to the request or response value will be made according to [.Req**] or [.Resp**] in [CustomHeaders. Req1].
# ⚠️：This action also affects the behavior of other middleware (if performed sequentially), so be careful with this middlware to avoid other actions that fail due to faulty headers.
[CustomHeaders]
    [CustomHeaders.req1] # req*** Specifies add this kv to HEADER of request header.
    Key="X-REAL-IP"
    Value="111.111.111.111"

    [CustomHeaders.resp1] # resp*** Specifies add this kv to HEADER of response header.
    Key="Server"
    Value="ASP.NET"

# The proxy Url
# According to Path (regular matching), the matching route is forwarded 
# for example:
# When a request is made to/superlcx/XXX, the request is forwarded to the specified Host
[ProxyUrls]
    [ProxyUrls.portrait]
    Scheme="http"
    Host="0.0.0.0:8989"
    Path="/superlcx/*"

# sub_filter config
# When the user accesses the `Path`, the content(line) response from the server and matches(regexp) the `Old` will be replaced with Repl by line.
[SubFilter]
    [SubFilter.test]
    Old="</head>"
    Repl='<script src="/js/jquery.min.js"></script></head>'
    Path="/"
```
