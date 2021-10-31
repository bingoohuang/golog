# hlog

log http client and service detail

## usage

[online demo](https://repl.it/join/hdiogkkz-bingoohuang)

```go
package main

import (
	"net/http"
	"time"

	"github.com/bingoohuang/golog"
	"github.com/bingoohuang/golog/pkg/hlog"
	"github.com/bingoohuang/sariaf"
	"github.com/sirupsen/logrus"
)

func main() {
	golog.Setup()
	r := sariaf.New()
	r.GET("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("Hello World")) })
	r.GET("/posts", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("GET: Get All Posts")) })
	server := &http.Server{Addr: ":8080", Handler: hlog.LogWrapHandler(r, logrus.StandardLogger())}

	go server.ListenAndServe()

	time.Sleep(100 * time.Millisecond)

	client := hlog.NewLoggedHTTPClient(logrus.StandardLogger())
	client.Get("http://127.0.0.1:8080")
	client.Post("http://127.0.0.1:8080/posts", "", nil)
}
```

```log
2020-12-22 10:34:55.726 [INFO ] 62201 --- [19   ] [-] hlog.DefaultLogger.LogRequest hlog.go:75 : Client Request GET / HTTP/1.1\r\nHost: 127.0.0.1:64755
2020-12-22 10:34:55.727 [INFO ] 62201 --- [5    ] [-] hlog.DefaultLogger.LogRequest hlog.go:75 : Server Request GET / HTTP/1.1\r\nHost: 127.0.0.1:64755\r\nAccept-Encoding: gzip\r\nUser-Agent: Go-http-client/1.1
2020-12-22 10:34:55.727 [INFO ] 62201 --- [5    ] [-] hlog.DefaultLogger.LogWriter server.go:108 : Server Response duration:8.55µs status:0 header:map[] payload:Hello World
2020-12-22 10:34:55.728 [INFO ] 62201 --- [19   ] [-] hlog.DefaultLogger.LogResponse hlog.go:92 : Client Response Duration:975.296µs error:<nil> Dump:HTTP/1.1 200 OK\r\nContent-Length: 11\r\nContent-Type: text/plain; charset=utf-8\r\nDate: Tue, 22 Dec 2020 02:34:55 GMT\r\n\r\nHello World
2020-12-22 10:34:55.728 [INFO ] 62201 --- [19   ] [-] hlog.DefaultLogger.LogRequest hlog.go:75 : Client Request POST /posts HTTP/1.1\r\nHost: 127.0.0.1:64755
2020-12-22 10:34:55.728 [INFO ] 62201 --- [5    ] [-] hlog.DefaultLogger.LogRequest hlog.go:75 : Server Request POST /posts HTTP/1.1\r\nHost: 127.0.0.1:64755\r\nAccept-Encoding: gzip\r\nContent-Length: 0\r\nUser-Agent: Go-http-client/1.1
2020-12-22 10:34:55.728 [INFO ] 62201 --- [5    ] [-] hlog.DefaultLogger.LogWriter server.go:108 : Server Response duration:8.856µs status:404 header:map[Content-Type:[text/plain; charset=utf-8] X-Content-Type-Options:[nosniff]] payload:Not Found
2020-12-22 10:34:55.729 [INFO ] 62201 --- [19   ] [-] hlog.DefaultLogger.LogResponse hlog.go:92 : Client Response Duration:475.722µs error:<nil> Dump:HTTP/1.1 404 Not Found\r\nContent-Length: 10\r\nContent-Type: text/plain; charset=utf-8\r\nDate: Tue, 22 Dec 2020 02:34:55 GMT\r\nX-Content-Type-Options: nosniff\r\n\r\nNot Found
```