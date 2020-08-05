package ginlogrus_test

import (
	"net/http"
	"testing"

	"github.com/bingoohuang/golog"
	"github.com/bingoohuang/golog/pkg/ginlogrus"
	"github.com/gin-gonic/gin"
)

func TestGinlogrus(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	golog.SetupLogrus(nil, "", "")

	r := gin.New()
	r.Use(ginlogrus.Logger(nil, true), gin.Recovery())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	server := &http.Server{Addr: ":12345", Handler: r}

	go func() {
		server.ListenAndServe()
	}()

	_, _ = http.Get("http://127.0.0.1:12345/ping")
	//rsp.Body.Close()

	server.Close()
}
