package main

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/gin-gonic/gin"
)

// Wrap adds several routes from package `net/http/pprof` to *gin.Engine object
func Wrap(router interface{}) {
	if r, ok := router.(*gin.Engine); ok {
		WrapGroup(&r.RouterGroup)
	} else if r, ok := router.(*gin.RouterGroup); ok {
		WrapGroup(r)
	} else if r, ok := router.(*http.ServeMux); ok {
		WrapServeMux(r)
	} else {
		panic(fmt.Errorf("please wrap *gin.Engine or *gin.RouterGroup"))
	}
}

// WrapGroup adds several routes from package `net/http/pprof` to *gin.RouterGroup object
func WrapGroup(router *gin.RouterGroup) {
	basePath := strings.TrimSuffix(router.BasePath(), "/")

	var prefix string

	switch {
	case basePath == "":
		prefix = ""
	case strings.HasSuffix(basePath, "/debug"):
		prefix = "/debug"
	case strings.HasSuffix(basePath, "/debug/pprof"):
		prefix = "/debug/pprof"
	}

	for _, r := range routers {
		router.Handle(r.Method, strings.TrimPrefix(r.Path, prefix), r.Handler)
	}
}

// WrapServeMux adds several routes from package `net/http/pprof` to *http.ServeMux object
func WrapServeMux(mux *http.ServeMux) {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	mux.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))
	mux.Handle("/debug/pprof/block", pprof.Handler("block"))
	mux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	mux.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))
}

var routers = []struct {
	Method  string
	Path    string
	Handler gin.HandlerFunc
}{
	{"GET", "/debug/pprof/", func(c *gin.Context) {
		pprof.Index(c.Writer, c.Request)
	}},
	{"GET", "/debug/pprof/cmdline", func(c *gin.Context) {
		pprof.Cmdline(c.Writer, c.Request)
	}},
	{"GET", "/debug/pprof/profile", func(c *gin.Context) {
		pprof.Profile(c.Writer, c.Request)
	}},
	{"GET", "/debug/pprof/symbol", func(c *gin.Context) {
		pprof.Symbol(c.Writer, c.Request)
	}},
	{"POST", "/debug/pprof/symbol", func(c *gin.Context) {
		pprof.Symbol(c.Writer, c.Request)
	}},
	{"GET", "/debug/pprof/trace", func(c *gin.Context) {
		pprof.Trace(c.Writer, c.Request)
	}},
	{"GET", "/debug/pprof/heap", func(c *gin.Context) {
		pprof.Handler("heap").ServeHTTP(c.Writer, c.Request)
	}},
	{"GET", "/debug/pprof/goroutine", func(c *gin.Context) {
		pprof.Handler("goroutine").ServeHTTP(c.Writer, c.Request)
	}},
	{"GET", "/debug/pprof/allocs", func(c *gin.Context) {
		pprof.Handler("allocs").ServeHTTP(c.Writer, c.Request)
	}},
	{"GET", "/debug/pprof/block", func(c *gin.Context) {
		pprof.Handler("block").ServeHTTP(c.Writer, c.Request)
	}},
	{"GET", "/debug/pprof/threadcreate", func(c *gin.Context) {
		pprof.Handler("threadcreate").ServeHTTP(c.Writer, c.Request)
	}},
	{"GET", "/debug/pprof/mutex", func(c *gin.Context) {
		pprof.Handler("mutex").ServeHTTP(c.Writer, c.Request)
	}},
}
