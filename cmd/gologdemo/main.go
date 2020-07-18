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

	golog.SetupLogrus(nil,
		"file=gologdemo.log,maxSize=1M,stdout=false,"+
			"rotate=.yyyy-MM-dd-HH-mm,maxAge=5m,gzipAge=3m")

	logC := make(chan LogMessage, 100)
	for i := 0; i < 1000; i++ {
		go func(workerID int) {
			for {
				msg := <-logC
				logrus.
					WithField("workerID", workerID).
					WithField("proto", msg.Proto).
					WithField("contentType", msg.ContentType).
					Infof("%s %s %s %s %s",
						msg.Time, msg.RemoteAddr, msg.Method, msg.URL, randx.String(100))
			}
		}(i)
	}

	addr := port.FreeAddr()
	fmt.Println("start to listen on", addr)

	go func() {
		// Now you must write to apachelog library can create
		// a http.Handler that only writes the appropriate logs for the request to the given handle
		if err := http.ListenAndServe(addr, logRequest(mux, logC)); err != nil {
			panic(err)
		}
	}()

	for {
		time.Sleep(3 * time.Second)
		http.Get("http://127.0.0.1" + addr) // nolint:errcheck
	}
}

type LogMessage struct {
	Time        string
	Proto       string
	ContentType string
	RemoteAddr  string
	Method      string
	URL         string
}

func logRequest(handler http.Handler, logC chan LogMessage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg := LogMessage{
			Time:        time.Now().Format("2006-01-02 15:04:05.000"),
			Proto:       r.Proto,
			ContentType: r.Header.Get("Content-Type"),
			RemoteAddr:  r.RemoteAddr,
			Method:      r.Method,
			URL:         r.URL.String(),
		}

		for i := 0; i < 1000; i++ {
			logC <- msg
		}

		handler.ServeHTTP(w, r)
	})
}
