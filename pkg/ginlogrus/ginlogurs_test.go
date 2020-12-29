package ginlogrus_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/bingoohuang/golog"
	"github.com/bingoohuang/golog/pkg/ginlogrus"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func TestGinlogrus(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	_ = golog.SetupLogrus(golog.Spec("file=~/gologdemo.log,level=debug"))

	r := gin.New()
	r.Use(ginlogrus.Logger(nil, true))

	r.GET("/ping", func(c *gin.Context) {
		ginlogrus.NewLoggerGin(c, nil).Info("pinged1")
		logrus.Info("pinged2")
		c.JSON(200, gin.H{"message": "pong"})

		logrus.Info("trace id:", ginlogrus.GetTraceIDGin(c))
	})

	r.GET("/t.html", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	})

	r.GET("/t.jpeg", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	})

	server := &http.Server{Addr: ":12345", Handler: r}

	go func() {
		_ = server.ListenAndServe()
	}()

	go func() {
		time.Sleep(3 * time.Second)
		rsp, _ := http.Get("http://127.0.0.1:12345/ping")
		rsp.Body.Close()
		rsp, _ = http.Get("http://127.0.0.1:12345/t.jpeg")
		rsp.Body.Close()
		rsp, _ = http.Get("http://127.0.0.1:12345/t.html")
		rsp.Body.Close()

		server.Close()
	}()
}
