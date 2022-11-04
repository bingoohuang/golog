package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bingoohuang/golog"
	"github.com/bingoohuang/golog/pkg/ginlogrus"
	"github.com/bingoohuang/golog/pkg/httpx"
	"github.com/bingoohuang/golog/pkg/logctx"
	"github.com/bingoohuang/golog/pkg/port"
	"github.com/bingoohuang/golog/pkg/randx"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const channelSize = 1000

func main() {
	ginHttp := flag.Bool("gin", false, "start gin http server for concurrent testing...")
	std := flag.Bool("std", false, "fix log.Print...")
	limit := flag.String("limit", "", "test limit, like 100,3s to limit 1 log every 100 logs or every 3s")
	features := flag.String("features", "", "features, available: fatal")
	sleep := flag.Duration("sleep", 100*time.Millisecond, "sleep duration lime, like 10s, default 10ms")
	cs := flag.Bool("cs", false, "http client and server logging")
	pprof := flag.String("pprof", "", "Profile pprof address, like localhost:6060")
	help := flag.Bool("help", false, `SPEC="file=demo.log,maxSize=300M,stdout=false,rotate=.yyyy-MM-dd,maxAge=10d,gzipAge=3d" ./golog`)
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *pprof != "" {
		go func() {
			// http://localhost:6060/debug/pprof/
			log.Printf("Starting pprof at %s", *pprof)
			log.Println(http.ListenAndServe(*pprof, nil))
		}()
	}

	golog.Setup()

	log.Printf("[L:%s] W! ignore sync %s.%s", "15s", "r.Schema", "r.Table")

	logrus.Infof("hello\nworld")
	logrus.Infof("[PRE]hello\nworld")
	log.Printf("hello\nworld")
	log.Printf("[PRE]hello\nworld")

	log.Printf("Hello, this message is logged by std log, #%d", 0)

	if strings.Contains(*features, "fatal") {
		err := fmt.Errorf("test fatal features")
		log.Fatal("E! " + err.Error())
	}

	if *limit != "" {
		for i := 0; i < 30*100; i++ {
			log.Printf("[L:"+*limit+"] limit test %d", i)
			time.Sleep(10 * time.Millisecond)
		}

		return
	}

	if *ginHttp {
		gin.SetMode(gin.ReleaseMode)
		r := gin.New()
		r.Use(ginlogrus.Logger(nil, true))

		r.GET("/", func(c *gin.Context) {
			ginlogrus.NewLoggerGin(c, nil).Info("pinged1")
			logrus.Info("pinged2")
			c.JSON(200, gin.H{"message": "pong"})

			logrus.Info("trace id:", ginlogrus.GetTraceIDGin(c))
		})

		server := &http.Server{Addr: ":12345", Handler: r}
		logrus.Info("start gin http server :12345 for gobench test")
		_ = server.ListenAndServe()
		return
	}

	if *std {
		log.Printf("Hello, this message is logged by std log, #%d", 1)
		log.Printf("T! Hello, this message is logged by std log, #%d", 2)
		log.Printf("D! Hello, this message is logged by std log, #%d", 3)
		log.Printf("I! Hello, this message is logged by std log, #%d", 4)
		log.Printf("W! Hello, this message is logged by std log, #%d", 5)
		log.Printf("F! Hello, this message is logged by std log, #%d", 6)

		logrus.Tracef("Hello, this message is logged by std log, #%d", 7)
		logrus.Debugf("Hello, this message is logged by std log, #%d", 8)
		logrus.Infof("Hello, this message is logged by std log, #%d", 9)
		logrus.Warnf("Hello, this message is logged by std log, #%d", 10)

		return
	}

	spec := "file=~/gologdemo.log,maxSize=1M,stdout=true,rotate=.yyyy-MM-dd-HH-mm,maxAge=5m,gzipAge=3m"
	if v := os.Getenv("SPEC"); v != "" {
		spec = v
	}

	layout := `%t{yyyy-MM-dd HH:mm:ss.SSS} [Watch:%context{name=WatchID}] [%5l{length=4}] PID=%pid --- [GID=%5gid] [%trace] %20caller{level=info} : %fields %msg%n`
	if v := os.Getenv("LAYOUT"); v != "" {
		layout = v
	}

	// 默认不开启，只有配置为 on 时生效
	// os.Setenv("GOLOG_DEBUG", "on")

	// 仅仅只需要一行代码，设置golog对于logrus的支持
	_ = golog.Setup(golog.Spec(spec), golog.Layout(layout))
	golog.RegisterLimiter(golog.LimitConf{EveryTime: 200 * time.Millisecond, Key: "log.hello"})

	for i := 0; i < 10; i++ {
		logctx.Set("WatchID", fmt.Sprintf("W%d", i+1))
		log.Printf("W! log context i:%d", i)
		time.Sleep(90 * time.Millisecond)
	}

	for i := 0; i < 10; i++ {
		logctx.Set("WatchID", fmt.Sprintf("W%d", i+1))
		log.Printf("[L:log.hello] W! log Hello1 i:%d", i)
		time.Sleep(90 * time.Millisecond)
	}

	golog.RegisterLimiter(golog.LimitConf{
		EveryTime: 200 * time.Millisecond,
		Key:       "log.hello2",
	})
	for i := 0; i < 10; i++ {
		log.Printf("[L:log.hello2] W! log Hello2 i:%d", i)
		time.Sleep(90 * time.Millisecond)
	}

	for i := 0; i < 10; i++ {
		log.Printf("[L:log.hello2] W! log Hello3 i:%d", i)
		time.Sleep(90 * time.Millisecond)
	}

	if !*cs {
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK\n"))
	})

	// Wrap(mux)

	log.Println("golog spec:", spec)

	logC := make(chan LogMessage, channelSize)
	for i := 0; i < channelSize; i++ {
		go logging(logC, i)
	}

	addr := port.FreeAddr()
	urlAddr := "http://127.0.0.1" + addr

	log.Println("start to listen on", addr)

	go func() {
		log.Printf("W! ListenAndServe error %+v", http.ListenAndServe(addr, logRequest(mux, logC)))
	}()

	ticker := time.NewTicker(*sleep)
	defer ticker.Stop()

	for range ticker.C {
		restclient(urlAddr)
	}
}

func restclient(urlAddr string) {
	log.Printf("I! invoke %s", urlAddr)
	rsp, _ := http.Get(urlAddr)
	// 这里我要写一段注释，记录一下，因为下面这行代码的缺失，花了我半天纠结的问题
	// 我写这个demo，本意就是要观察，启动1000个协程并发去写日志
	// 日志文件按时间和大小的去定期定大小滚动，以及定期删除老旧日志，观察是否是期望那样
	// 但是，本demo，无论是本机(macOS)，还是Linux，都是跑着跑着，大概五六分钟的样子，日志输出就停止了
	// 对的，tail -F ~/gologdemo.log，不再更新了
	// 对的，watch "ls -lht ~/gologdemo.log*"也没有更新了
	// gologdemo程序也没有死，也还在
	// 内心感觉很慌，恰如，前几天吕勇说的，日志打着打着，就不打了，...
	// 于是我把滚动打印日志的rotate.go文件，从头到尾，从尾又到头，反复读(code review)了好几遍
	// 该处理记录error地方的，都统统补充处理记录了，尽管我怀疑是哪个地方的error被黑洞swallow了
	// 该处理多线程竞争的地方，也没有问题，尽管郑科研言之凿凿的说，就是多线程的线程锁导致的(7-18下午4:08企业微信)
	// 现象就是：日志打着打着，就停止了，没有新的日志输出了，也没有任何错误
	//（苦逼的时候就是这样子，没有告警，没有错误，外表一切正常，就是不是期望结果)
	// 唯一庆幸的是，macOS和Linux，尽皆如此😭😄（哭笑不得），说明不是偶然现象
	// 唯一不一样的是，在Idea的debug运行的时候，竟然久久也不会出问题（一顿中饭，一个午觉，至少一个小时吧）
	//（我还想着，debug模式运行也出问题的时候，我顺便debug单步调试一下）
	// 于是，我开始怀疑，是不是golog实现问题，而是logrus是否有问题，又或者别的问题
	// 于是，我在linux上开着strace来跑gologdemo
	// 跑着跑着，报告说，accept文件句柄超标了，哈哈哈，我眼睛一亮，终于感觉抓住它了，因为似曾相似
	// accept是http server的概念，难道，go http server有问题了，查了一下，没发现啥
	// 顺便看看client端吧，结果发现在死循环里，http.Get(...)一直调用，我突然想到
	// 有一个body是Closeable的，需要手动关闭
	// 于是，加上了下面这行代码，然后重新跑应用，欧耶，一切重回正轨，三观回正了.
	httpx.CloseResponse(rsp)
}

func logging(logC <-chan LogMessage, workerID int) {
	for r := range logC {
		logrus.WithFields(map[string]interface{}{
			"workerID":    workerID,
			"proto":       r.Proto,
			"contentType": r.ContentType,
		}).Infof("%s %s %s %s %s", r.Time, r.RemoteAddr, r.Method, r.URL, randx.String(100))
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
		msg := LogMessage{
			Time:        time.Now().Format("2006-01-02 15:04:05.000"),
			Proto:       r.Proto,
			ContentType: r.Header.Get("Content-Type"),
			RemoteAddr:  r.RemoteAddr,
			Method:      r.Method,
			URL:         r.URL.String(),
		}

		for i := 0; i < channelSize; i++ {
			logC <- msg
		}

		handler.ServeHTTP(w, r)
	})
}
