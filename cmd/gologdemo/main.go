package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/bingoohuang/golog/apachelog"
	"github.com/bingoohuang/golog/rotate"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK\n"))
	})

	logf, err := rotate.New(
		"access_log.%Y%m%d%H%M",
		rotate.WithMaxAge(24*time.Hour),
	)
	if err != nil {
		log.Printf("failed to create rotatelogs: %s", err)
		return
	}

	addr := FreeAddr()
	fmt.Println("start to listen on", addr)

	// Now you must write to logf. apache-logformat library can create
	// a http.Handler that only writes the approriate logs for the request
	// to the given handle
	if err := http.ListenAndServe(addr, apachelog.CombinedLog.Wrap(mux, logf)); err != nil {
		panic(err)
	}
}

// FreeAddr asks the kernel for a free open port that is ready to use.
func FreeAddr() string {
	if v := os.Getenv("ADDR"); v != "" {
		return v
	}

	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return ":10020"
	}

	_ = l.Close()

	return fmt.Sprintf(":%d", l.Addr().(*net.TCPAddr).Port)
}
