package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bingoohuang/golog"
	"github.com/bingoohuang/golog/pkg/port"
	"github.com/bingoohuang/golog/pkg/randx"
	"github.com/sirupsen/logrus"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK\n"))
	})

	golog.SetupLogrus(nil, "file=gologdemo.log,maxSize=1K,rotate=.yyyy-MM-dd-HH-mm,maxAge=5m,gzipAge=3m")

	for i := 0; i < 100; i++ {
		i := i
		go func() {
			logrus.Infof("go routine %d", i)
		}()
	}

	addr := port.FreeAddr()
	fmt.Println("start to listen on", addr)

	go func() {
		// Now you must write to apachelog library can create
		// a http.Handler that only writes the appropriate logs for the request to the given handle
		if err := http.ListenAndServe(addr, logRequest(mux)); err != nil {
			panic(err)
		}
	}()

	for {
		time.Sleep((3 * time.Second))
		http.Get("http://127.0.0.1" + addr) // nolint:errcheck
	}
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.
			WithField("proto", r.Proto).
			WithField("contentType", r.Header.Get("Content-Type")).
			Infof("%s %s %s %s", r.RemoteAddr, r.Method, r.URL, randx.String(100))

		handler.ServeHTTP(w, r)
	})
}
