package main

import (
	"fmt"
	"github.com/bingoohuang/golog"
	"github.com/bingoohuang/golog/pkg/port"
	"github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK\n"))
	})

	addr := port.FreeAddr()
	fmt.Println("start to listen on", addr)

	golog.SetupLogrus(nil, "file=gologdemo.log")

	// Now you must write to apachelog library can create
	// a http.Handler that only writes the appropriate logs for the request to the given handle
	if err := http.ListenAndServe(addr, logRequest(mux)); err != nil {
		panic(err)
	}
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.
			WithField("proto", r.Proto).
			WithField("contemtType", r.Header.Get("Content-Type")).
			Infof("%s %s %s", r.RemoteAddr, r.Method, r.URL)

		handler.ServeHTTP(w, r)
	})
}
