package hlog_test

import (
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/bingoohuang/golog/pkg/hlog"
	"github.com/bingoohuang/sariaf"
	"github.com/sirupsen/logrus"
)

func TestExample(t *testing.T) {
	r := sariaf.New()
	r.SetNotFound(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "Not Found", 404) })
	r.SetPanicHandler(func(w http.ResponseWriter, r *http.Request, err interface{}) {
		http.Error(w, fmt.Sprintf("Internal Server Error:%v", err), 500)
	})

	_ = r.GET("/", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("Hello World")) })
	_ = r.GET("/posts", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("GET: Get All Posts")) })

	l, _ := net.Listen("tcp", ":0")
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()

	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: hlog.WrapHandler(r, logrus.StandardLogger())}
	defer server.Close()

	go func() {
		logrus.Infof("ListenAndServe %v", server.ListenAndServe())
	}()

	base := fmt.Sprintf("http://127.0.0.1:%d", port)

	c := hlog.NewHTTPClient(logrus.StandardLogger())
	Rest(c, http.MethodGet, base)
	Rest(c, "POST", base+"/posts")
}

func Rest(c *http.Client, method, url string) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return
	}

	_, _ = c.Do(req)
}
