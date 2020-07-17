package main

import (
	"fmt"
	"github.com/bingoohuang/golog/pkg/apachelog"
	"github.com/bingoohuang/golog/pkg/port"
	"github.com/bingoohuang/golog/pkg/rotate"
	"github.com/bingoohuang/golog/pkg/timex"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK\n"))
	})

	r, err := rotate.New(
		"golog_access.log",
		rotate.WithRotateLayout(".yyyy-MM-dd-HH-mm"),
		rotate.WithMaxAge(1*timex.Week),
	)
	if err != nil {
		log.Printf("failed to create Rotate: %s", err)
		return
	}

	addr := port.FreeAddr()
	fmt.Println("start to listen on", addr)

	// Now you must write to apachelog library can create
	// a http.Handler that only writes the appropriate logs for the request to the given handle
	if err := http.ListenAndServe(addr, apachelog.CombinedLog.Wrap(mux, r)); err != nil {
		panic(err)
	}
}
