package ginlogrus_test

import (
	"net/http"
	"testing"

	"github.com/bingoohuang/golog"
	"github.com/bingoohuang/golog/pkg/ginlogrus"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func TestGinlogrus(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	_, _ = golog.SetupLogrus(nil, "", "")

	r := gin.New()
	r.Use(ginlogrus.Logger(nil), gin.Recovery())

	r.GET("/ping", func(c *gin.Context) {
		ginlogrus.NewLoggerGin(c, nil).Info("pinged1")
		logrus.Info("pinged2")
		c.JSON(200, gin.H{"message": "pong"})
	})

	server := &http.Server{Addr: ":12345", Handler: r}

	go func() {
		_ = server.ListenAndServe()
	}()

	rsp, _ := http.Get("http://127.0.0.1:12345/ping")
	_ = rsp.Body.Close()

	_ = server.Close()
}
