# golog

[![Travis CI](https://travis-ci.com/bingoohuang/golog.svg?branch=master)](https://travis-ci.com/bingoohuang/golog)
[![Software License](https://img.shields.io/badge/License-MIT-orange.svg?style=flat-square)](https://github.com/bingoohuang/golog/blob/master/LICENSE.md)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/bingoohuang/golog)
[![Coverage Status](http://codecov.io/github/bingoohuang/golog/coverage.svg?branch=master)](http://codecov.io/github/bingoohuang/golog?branch=master)
[![goreport](https://www.goreportcard.com/badge/github.com/bingoohuang/golog)](https://www.goreportcard.com/report/github.com/bingoohuang/golog)

golog，支持:

1. 日志格式化标准
1. 日志级别颜色
1. 按天/大小滚动
1. 自动压缩
1. 自动删除
1. 自动日志文件名
1. logrus一行集成
1. 日志限速

## Integration with logrus and log

Use default settings:

```go
import "github.com/bingoohuang/golog"

func main() {
	golog.SetupLogrus()

	log.Printf("Hello, this message is logged by std log, #%d", 1) // default Info
	log.Printf("T! Hello, this message is logged by std log, #%d", 2) // Trace
	log.Printf("D! Hello, this message is logged by std log, #%d", 3) // Debug
	log.Printf("I! Hello, this message is logged by std log, #%d", 4) // Info
	log.Printf("W! Hello, this message is logged by std log, #%d", 5) // Warn
	log.Printf("F! Hello, this message is logged by std log, #%d", 6) // Fatal

	logrus.Tracef("Hello, this message is logged by std log, #%d", 7)
	logrus.Debugf("Hello, this message is logged by std log, #%d", 8)
	logrus.Infof("Hello, this message is logged by std log, #%d", 9)
	logrus.Warnf("Hello, this message is logged by std log, #%d", 10)
}
```

will output the log messages as below:

```sh
2020-12-29 08:52:43.810 [INFO ] 29490 --- [1    ] [-] logfmt.LogrusOption.Setup logrus.go:121 : log file created:~/logs/gologdemo.log
2020-12-29 08:52:43.810 [INFO ] 29490 --- [1    ] [-] main.main main.go:35 : Hello, this message is logged by std log, #1
2020-12-29 08:52:43.810 [INFO ] 29490 --- [1    ] [-] main.main main.go:38 : Hello, this message is logged by std log, #4
2020-12-29 08:52:43.810 [WARN ] 29490 --- [1    ] [-] main.main main.go:39 : Hello, this message is logged by std log, #5
2020-12-29 08:52:43.811 [FATAL] 29490 --- [1    ] [-] main.main main.go:40 : Hello, this message is logged by std log, #6
2020-12-29 08:52:43.811 [INFO ] 29490 --- [1    ] [-] main.main main.go:44 : Hello, this message is logged by std log, #8
2020-12-29 08:52:43.811 [WARN ] 29490 --- [1    ] [-] main.main main.go:45 : Hello, this message is logged by std log, #8
```

Customize the settings:

```go
golog.SetupLogrus(golog.Spec("level=debug,rotate=.yyyy-MM-dd-HH,maxAge=5d,gzipAge=1d"))
```

## Specifications

name       | env              | prerequisite    | default value          | description
-----------|------------------|-----------------|------------------------|-----------------------------------------------------------------------------------------------------
level      | GOLOG_LEVEL      | -               | info                   | log level to record (debug/info/warn/error)
file       | GOLOG_FILE       | -               | ~/logs/{bin}/{bin}.log | base log file name, if root user, default log file will be /var/log/{bin}/{bin}.log
rotate     | GOLOG_ROTATE     | -               | .yyyy-MM-dd            | time rotate pattern(full pattern: yyyy-MM-dd HH:mm)[Split according to the Settings of the last bit]
maxAge     | GOLOG_MAXAGE     | -               | 30d                    | max age to keep log files (unit m/h/d/w)
gzipAge    | GOLOG_GZIPAGE    | -               | 3d                     | gzip aged log files (unit m/h/d/w)
maxSize    | GOLOG_MAXSIZE    | -               | 100M                   | max size to rotate log files (unit K/M/K/KiB/MiB/GiB/KB/MB/GB)
stdout     | GOLOG_STDOUT     | -               | false                  | print the log to stdout at the same time or not
printColor | GOLOG_PRINTCOLOR | layout is empty | true                   | print color on the log level or not, only for stdout=true
printCall  | GOLOG_PRINTCALL  | layout is empty | false                  | print caller file:line or not (performance slow)
simple     | GOLOG_SIMPLE     | layout is empty | false                  | simple to print log (not print `PID --- [GID] [TraceID]`)
layout     | GOLOG_LAYOUT     | -               | (empty)                | log line layout customization, like `%t %5l %pid --- [%5gid] [%trace] %20caller : %fields %msg%n`
fixstd     | GOLOG_FIXSTD     | -               | true                   | improve standard log for golog format.

### file

1. If the file is an existed directory, like `/var/log/`, a log file will appended as `/var/log/{bin}.log`
1. If the file is not a valid directory or file, the suffix '.log' will be used to distinguished as directory or logfile.

so examples:

1. /var/log(existed dir) -> /var/log/{bin}.log
1. /var/log/app.log(only existed dir) ->  (same)
1. /var/log/applog(existed file) -> (same)
1. /home/bingoo/logs(not existed w/o .log) -> /home/bingoo/logs/{bin}.log
1. /home/bingoo/logs/app.log(not existed w/ .log) -> (same)

## Layout pattern

patter                 | remark
-----------------------|-----------------------------------------------------------------------
`%time`  `%t`          | `%time` same with `%time{yyyy-MM-dd HH:mm:ss.SSS}`
`%level`  `%l`         | `%level` same with `%level{printColor=false lowerCase=false length=0}`
`%pid`                 | process ID
`%gid`                 | go routine ID
`%08gid`               | go routine ID, Pad with leading zeroes(width 8)
`%5gid`                | go routine ID, Pad with spaces (width 5, right justified)
`%-10trace`            | trace ID, Pad with spaces (width 10, left justified)
`%caller`              | caller information, `%caller{sep=:,level=warn}`, `sep` defines the separator between filename and line number, `level` defines the lowest level to print caller information.
`%fields`              | fields JSON
`%message` `%msg` `%m` | log detail message, `%m{singleLine=true}`, `singleLine` indicates whether the message should merged into a single line when there are multiple newlines in the message.
`%n`                   | new line
`%%`                   | escape percent sign

## Demonstration

```log
$ go get github.com/bingoohuang/golog/gologdemo
$ ADDR=":54264" gologdemo
log file created: gologdemo.log
start to listen on :54264
2020-07-18 10:59:24.179 removed by 5m0s gologdemo.log.2020-07-18-10-48.1.gz
2020-07-18 10:59:24.179 removed by 5m0s gologdemo.log.2020-07-18-10-48.2.gz
2020-07-18 10:59:24.202 log file renamed to  gologdemo.log.2020-07-18-10-59.2
2020-07-18 10:59:36.225 log file renamed to  gologdemo.log.2020-07-18-10-59.3
2020-07-18 10:59:36.226 removed by 5m0s gologdemo.log.2020-07-18-10-48.3.gz
2020-07-18 10:59:45.293 log file renamed to  gologdemo.log.2020-07-18-10-59.4
2020-07-18 10:59:45.293 removed by 5m0s gologdemo.log.2020-07-18-10-48.4.gz
2020-07-18 10:59:57.313 log file renamed to  gologdemo.log.2020-07-18-10-59.5
2020-07-18 10:59:57.313 removed by 5m0s gologdemo.log.2020-07-18-10-48.5.gz
2020-07-18 11:00:00.297 log file renamed to  gologdemo.log.2020-07-18-11-00
2020-07-18 11:00:09.360 log file renamed to  gologdemo.log.2020-07-18-11-00.1
2020-07-18 11:00:09.361 removed by 5m0s gologdemo.log.2020-07-18-10-49.gz
2020-07-18 11:00:21.391 log file renamed to  gologdemo.log.2020-07-18-11-00.2
```

demo log file content:

```log
2020-07-23 12:26:33.144 [INFO ] 9019 --- 214   [-] main.go:101          : {"contentType":"","proto":"HTTP/1.1","workerID":207} 2020-07-23 12:26:33.098 127.0.0.1:34458 GET / dkNTliprMVGmkfLaOPqIDEwZBHUVBeukHOmAEsTDFRsGbqcuwcnhUNOQZGyGZazNwxFfOumzuUSdnzCOvIUASPlddzWhZsjyEbhU
2020-07-23 12:26:33.144 [INFO ] 9019 --- 669   [-] main.go:101          : {"contentType":"","proto":"HTTP/1.1","workerID":662} 2020-07-23 12:26:33.098 127.0.0.1:34458 GET / FMVACHSBfekZuLPiGrjPOrMfsImGEWTIcLiBbcHlTJWpuVMzhDRyvBThyOUBOllxUEPJlMGGhXhyLHZknzcNaJycUysJuBFhdQjJ
2020-07-23 12:26:33.144 [INFO ] 9019 --- 379   [-] main.go:101          : {"contentType":"","proto":"HTTP/1.1","workerID":372} 2020-07-23 12:26:33.098 127.0.0.1:34458 GET / cQbUXmEQUrJXouDNlMyDBFhykLOjCaNRbwDEdTjsUlZTCIWHsycwhnGitpRDfICIwTKeGlAMPVyMWnLQXUHFIcPtLLudqiGtfkvH
2020-07-23 12:26:33.144 [INFO ] 9019 --- 119   [-] main.go:101          : {"contentType":"","proto":"HTTP/1.1","workerID":112} 2020-07-23 12:26:33.098 127.0.0.1:34458 GET / uMAaQDwxDPlBEtrFiGbhyFNFJPDHNdQFcYzNvXgsrtBcjZQDCIDYBnDaasJKNVQSbaQPUmzCFRCCqppwIGGgzSIrBvFVopVvSZHL
2020-07-23 12:26:33.144 [INFO ] 9019 --- 244   [-] main.go:101          : {"contentType":"","proto":"HTTP/1.1","workerID":237} 2020-07-23 12:26:33.098 127.0.0.1:34458 GET / OPBsKeflLHELLpymnBqxaGrvEyAjCStmMUAKzXUPFrJQOepHQzfARmStzkTmzWJnZmEJqEjiPQxCKhnicomhYBKiOXoyimKMPipp
2020-07-23 12:26:33.144 [INFO ] 9019 --- 368   [-] main.go:101          : {"contentType":"","proto":"HTTP/1.1","workerID":361} 2020-07-23 12:26:33.098 127.0.0.1:34458 GET / vlnikHYVmbSqimicIgfpUBImaKRmPuxZnljbgbhEUONUUtAKjxsaBbTUJTdYjMVWxDkcbjWDgzUFENhfyBAHqHCLDOLjQXkYFocm
2020-07-23 12:26:33.144 [INFO ] 9019 --- 910   [-] main.go:101          : {"contentType":"","proto":"HTTP/1.1","workerID":903} 2020-07-23 12:26:33.098 127.0.0.1:34458 GET / ggcnhxToaBYtsMRWJCJEHpZViQphaRTwSMhUytFjdPzKltmjHRSReYfswPSbcrCSsUybGIibQPCRavxvwKMyQoOjelNPacinPYFK
2020-07-23 12:26:33.110 [INFO ] 9019 --- 660   [-] main.go:101          : {"contentType":"","proto":"HTTP/1.1","workerID":653} 2020-07-23 12:26:33.098 127.0.0.1:34458 GET / DcXihdRvktitsFKQmAGIpiFsDYslqnPQebbmQrqUGZGTdHAkGHvoUmCMiejCYEzfEriLFlcTjPiHDOxMaOdhyLyPaHyYmwqMCnok
2020-07-23 12:26:33.110 [INFO ] 9019 --- 440   [-] main.go:101          : {"contentType":"","proto":"HTTP/1.1","workerID":433} 2020-07-23 12:26:33.098 127.0.0.1:34458 GET / GkbfwhOuBvinWDBIdKrVWTbtKJsSDtCeJirgIoiUovmUALuCfHPdjYccdNZyGsWXFTlpmsBpbIJLlPVWuzuNxpcTAFcRQXGmOtjt
```

watch the log file rotating and gzipping and deleting `watch -c "ls -tlh gologdemo.log*"`

## Log Creation Sequence

```sh
+------------+   Write log messages    +-------------------------+
| file=a.log +------------------------>+ a.log created           |
+------------+                         +-------------------------+    rotate=.yyyy-MM-dd
                                                                    +----------------------+
+------------+                         +-------------------------+   time goes to next day |
|            +------------------------>+~/logs/{app}.log created |                         |
+------------+                         +-------------------------+                         |
                                                                                           |
                                                                                           |
                                                                                           |
                                       +-------------------------------------+             |
           maxSize=100M                |  x.log renamed to x.log.2020-07-17  |             |
   +-----------------------------------+                                     <-------------+
   |       still current day.          |  a new file x.log recreated!        |
   |                                   +-------------------------------------+
   |
   |
   |        +------------------------------------------------+
   |        | x.log renamed to x.log.2020-07-17.1(100M)      |    reached new 100M
   +-------->                                                +-----------------------------+
            | a new file x.log recreated to current writing. |    still current day        |
            +------------------------------------------------+                             |
                                                                                           |
                                                                                           |
                                         +-------------------------------------------+     |
       day goes to 2020-07-21            | x.log renamed to x.log.2020-07-17.2(100M) |     |
   +-------------------------------------+                                           <-----+
   |          gzipAge=3d                 | a new file x.log recreated!               |
   |                                     +-------------------------------------------+
   |
   | +-----------------------------+                          +----------------------------------+
   | | x.log.2020-07-17.1.log.gz   |   day goes to 2020-7-22  | x.log.2020|07|17.1.log.gz deleted|
   +->                             +------------------------->|                                  |
     | x.log.2020-07-17.2.log.gz   |      maxAge=4d           | x.log.2020|07|17.2.log.gz deleted|
     +-----------------------------+                          +----------------------------------+
```

## Log rate limiter

stdlib log

```go
logf := golog.NewLimitLog(1, 200*time.Millisecond, 2)
logf("Hello i:%d", i) // will limited to 200 lines per ms with burst 2.
```

logurs:

```go
golog.SetupLogrus()
logr := golog.NewLimitLogrus(nil, 1, 200*time.Millisecond, 2)
logr.Infof("Hello i:%d", i) // will limited to 200 lines per ms with burst 2.
```

## Help

1. `sed "s/\x1B\[([0-9]{1,2}(;[0-9]{1,2})?)?[m|K]//g" x.log` to strip color from log file.
1. [`tail -F x.log`](https://explainshell.com/explain?cmd=tail+-F+x.log) even if x.log recreated.
1. [GIN](https://github.com/gin-gonic/gin) framework extra logs will be printed?
    1. Use gin.New() instead of gin.Default()
    1. Because gin.Default() exist Logger(), gin.New() not print logs

## golog gin with trace ID

```go
import (
	"github.com/bingoohuang/golog"
	"github.com/bingoohuang/golog/pkg/ginlogrus"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	golog.SetupLogrus()

	r := gin.New()
	r.Use(ginlogrus.Logger(nil, true), gin.Recovery())

	r.GET("/ping", func(c *gin.Context) {
	ginlogrus.NewLoggerGin(c, nil).Info("pinged1")
	logrus.Info("pinged2")
	c.JSON(200, gin.H{"message": "pong"})

	fmt.Println("context trace id:", ginlogrus GetTraceIDGin(c))
})
// ...
}
```

```sh
2020-08-24 09:47:30.530 [INFO ] 68880 --- [24   ] [87513ae4-10d4-43f3-be5e-f8a11e636f4b] ginlogurs_test.go:21 : pinged1
2020-08-24 09:47:30.531 [INFO ] 68880 --- [24   ] [87513ae4-10d4-43f3-be5e-f8a11e636f4b] ginlogurs_test.go:22 : pinged2
2020-08-24 09:47:30.531 [INFO ] 68880 --- [24   ] [87513ae4-10d4-43f3-be5e-f8a11e636f4b] ginlogrus.go:64      : 127.0.0.1 GET /ping [200] 18  Go-http-client/1.1 (746.916µs)
```

## Thanks to the giant shoulders:

1. [lestrrat-go/file-rotatelogs](https://github.com/lestrrat-go/file-rotatelogs)
1. [benbjohnson/clock](https://github.com/benbjohnson/clock)
1. [rifflock/lfshook A local file system hook for golang logrus logger](https://github.com/rifflock/lfshook)
1. [mzky/utils 一个工具集，包括文件组件，日志组件](https://github.com/mzky/utils)
1. [Lumberjack writing logs to rolling files.](https://github.com/natefinch/lumberjack)
1. [应用程序日志规范](https://github.com/bingoohuang/blog/issues/151)
1. [op/go-logging format](https://github.com/op/go-logging/blob/master/format.go)
1. [log4j layout](https://logging.apache.org/log4j/2.x/manual/layouts.html)
1. [CLI 控制台颜色渲染工具库, 拥有简洁的使用API，支持16色，256色，RGB色彩渲染输出](https://github.com/gookit/color)
