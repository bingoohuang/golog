package echologrus_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/bingoohuang/golog/pkg/echologrus"

	"github.com/bingoohuang/golog"
	"github.com/sirupsen/logrus"
)

func TestEchoLogrus(t *testing.T) {
	_ = golog.Setup(golog.Spec("file=./gologdemo.log,level=debug"))

	r := echo.New()
	r.Use(echologrus.Logger(nil, true))

	r.GET("/ping", func(c echo.Context) error {
		echologrus.NewLoggerEcho(c, nil).Info("pinged1")
		logrus.Info("pinged2")
		c.JSON(200, map[string]string{"message": "test"})

		logrus.Info("trace id:", echologrus.GetTraceIDEcho(c))
		return nil
	})

	r.GET("/t.html", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"message": "test"})
	})

	r.GET("/t.jpeg", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"message": "test"})
	})

	server := &http.Server{Addr: ":12345", Handler: r}

	go func() {
		_ = server.ListenAndServe()
	}()

	time.Sleep(3 * time.Second)
	rsp, _ := http.Get("http://127.0.0.1:12345/ping")
	rsp.Body.Close()
	rsp, _ = http.Get("http://127.0.0.1:12345/t.jpeg")
	rsp.Body.Close()
	rsp, _ = http.Get("http://127.0.0.1:12345/t.html")
	rsp.Body.Close()

	server.Close()
}
