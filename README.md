[EN][中文](./README.CN.md)

# superlcx
port transfer tool with middleware kit

# usage
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

### mode
- proxy 
    - advantages: the proxy mode work with http proxy package which can offer(expose) more api like modifyResponse,Transport,director, etc.
    - disadvantages: the proxy mode will cause memory mem jitter, not suitable for limited memory machine if you need high performance

- copy
    - advantages: the copy mode directly work on tcp layer, it doesn't care about what would be transfer. With the help of `io.Copy`, it needs less RAM.
    - disadvantages: the copy mode know nothing about application layer. all things it can do is dumping them all out.
    
- hybrid(ing)
    - (expect)advantages: allocate low memory and expose more API like proxy_pass...etc