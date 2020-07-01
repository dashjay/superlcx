[EN][中文](./README.CN.md)

# superlcx
port transfer tool with middleware kit

# usage
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

### mode
- proxy 
    - advantages: the proxy mode work with http proxy package which can offer(expose) more api like modifyResponse,Transport,director, etc.
    - disadvantages: the proxy mode will cause memory jitter, not suitable for limited memory machine if you need high performance

- copy
    - advantages: the copy mode directly work on tcp layer, it doesn't care about what would be transfer. With the help of `io.Copy`, it needs less RAM.
    - disadvantages: the copy mode know nothing about application layer. all things it can do is dumping them all out.
    
- blend
    - advantages: allocate low memory comparing with the proxy mode. it can run with the middleware interface.
    - disadvantages: the blend mode still need more memory(less than proxy), and could lead to memory jitter.

### -c Configuration file
> read configured with a Golang [toml library](https://github.com/BurntSushi/toml)

⚠ If you read the configuration file with -c, please write all the configuration files, do not rely partly on the command line parameters, we do not know the internal implementation of the TOML library, may overwrite the origin config.

For the documentation of the configuration file, please see: [Configuration Description](./docs/config.md)


### -M middleware
When working in the proxy mode, middleware can be invoked to analyze the traffic in the process. For example, the built-in stdout middleware (sample middleware) can be used via '-M stdout'.
(Must be in the proxy or blend mode to work)

If you want to implement your own middleware, check out:
[Middleware standard](./docs/middleware.md)

Once every thing ok, it can be placed under the middlewares folder. It is disambiguated tha that the file structure remain as follows.
```
middlewares
└── stdout
    └── handler.go
```

The interfaces are exposing as follows under 'handler.go'. If you need to load the configuration, load it yourself, or use the cc module TOML to load your config.
```
func HandleRequest(req *http.Request)
func HandleResponse(req *http.Response)
```

**WARNING:** Middleware components may do something on the req and resp body, That's bound to do a lot of io with bufio, the memory jitters can be very serious.
I thought about writing a center middleware to dump the req and resp out, and then call all middleware like pipeline, but if some of them change the req or resp, it can also cause other problems, so now middleware components do not affect each other.