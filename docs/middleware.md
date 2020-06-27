[EN][中文](./middleware.CN.md)

## write a middleware


Middleware based on request body（ http.Request ）And return body（ http.Response ）Because both structures contain buf, such as the body of request and response body.

Therefore, in order to prevent superlcx read body, so that the client or server can't receive the real request body or response body, we need to do some operations.

The standard middleware `stdout` of the example write follows: 
```gotemplate
func handleRequest(over chan bool, req *http.Request) {
	content, err := httputil.DumpRequest(req, true)
	over <- true
	if err != nil {
		panic(err)
	}
...
```

The request body(if exists) will be read out by `httputil.DumpRequest`, and a copy of the request body will be assign to origin req. If we write a middleware by ourselves, we should do the same.

If you can't handle the body in the request and response well, you may have serious consequences:
- Body is empty, which can cause EOF error during reading.
- The body has been closed, panic when you read it.

In order to prevent the program from crashing for these reasons, it is recommended to write the middleware in the following way.
```
content, err := httputil.DumpRequest(req, true)
newReq, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(content)))
```

In this way, it can completely avoid the influence of req request body and send it to the server to receive the reply.
