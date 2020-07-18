package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/bingoohuang/golog"
	"github.com/bingoohuang/golog/pkg/port"
	"github.com/bingoohuang/golog/pkg/randx"
	"github.com/sirupsen/logrus"
)

const ChannelSize = 1000

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK\n"))
	})

	spec := "file=~/gologdemo.log,maxSize=1M,stdout=false," +
		"rotate=.yyyy-MM-dd-HH-mm,maxAge=5m,gzipAge=3m"

	if v := os.Getenv("SPEC"); v != "" {
		spec = v
	}

	fmt.Println("golog spec:", spec)

	golog.SetupLogrus(nil, spec)

	logC := make(chan LogMessage, ChannelSize)
	for i := 0; i < ChannelSize; i++ {
		go func(workerID int) {
			for r := range logC {
				logrus.WithFields(map[string]interface{}{
					"workerID":    workerID,
					"proto":       r.Proto,
					"contentType": r.ContentType,
				}).Infof("%s %s %s %s %s",
					r.Time, r.RemoteAddr, r.Method, r.URL, randx.String(100))
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

	for i := 0; ; i++ {
		time.Sleep(3 * time.Second)

		msg := LogMessage{
			Time:        time.Now().Format("2006-01-02 15:04:05.000"),
			Proto:       fmt.Sprintf("Proto%d", i),
			ContentType: fmt.Sprintf("Content-Type%d", i),
		}

		for i := 0; i < ChannelSize; i++ {
			logC <- msg
		}
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
		logC <- LogMessage{
			Time:        time.Now().Format("2006-01-02 15:04:05.000"),
			Proto:       r.Proto,
			ContentType: r.Header.Get("Content-Type"),
			RemoteAddr:  r.RemoteAddr,
			Method:      r.Method,
			URL:         r.URL.String(),
		}

		handler.ServeHTTP(w, r)
	})
}
