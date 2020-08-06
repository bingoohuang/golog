package ginlogrus_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/bingoohuang/golog"
	"github.com/bingoohuang/golog/pkg/ginlogrus"
	"github.com/gin-gonic/gin"
)

func TestGinlogrus(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	golog.SetupLogrus(nil, "file=~/gologdemo.log,level=debug", "")

	r := gin.New()
	r.Use(ginlogrus.Logger(nil, true))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	r.GET("/t.html", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	})

	r.GET("/t.jpeg", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	})

	go func() {
		r.Run(":9099")
	}()
	time.Sleep(time.Second * 3)
	_, _ = http.Get("http://127.0.0.1:9099/ping")
	_, _ = http.Get("http://127.0.0.1:9099/t.jpeg")
	_, _ = http.Get("http://127.0.0.1:9099/t.html")
}
