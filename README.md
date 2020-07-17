# golog

[![Travis CI](https://img.shields.io/travis/bingoohuang/golog/master.svg?style=flat-square)](https://travis-ci.com/bingoohuang/golog)
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

## Integration with logrus

Use default settings:

```go
func init() {
    golog.SetupLogrus(nil, "")
}
```

Customize the settings:

```go
func init() {
    golog.SetupLogrus(nil, "level=debug,rotate=.yyyy-MM-dd-HH,maxAge=5d,gzipAge=1d")
}
```

## Specifications

name       | default value    | description
-----------|------------------|-------------------------------------------------------------
level      | info             | log level to record (debug/info/warn/error)
file       | ~/logs/{bin}.log | base log file
rotate     | .yyyy-MM-dd      | time rotate pattern(yyyy-MM-dd HH:mm)
maxAge     | 30d              | max age to keep log files (unit s/m/h/d/w)
gzipAge    | 3d               | gzip aged log files (unit m/h/d/w)
maxSize    | 100M             | max size to rotate log files (unit K/M/K)
printColor | true             | print color on the log level or not
printCall  | true             | print caller file and line number  or not (performance slow)
stdout     | true             | print the log to stdout at the same time or not

## Demonstration

```log
$ go get github.com/bingoohuang/golog/gologdemo
$ ADDR=":54264" gologdemo
start to listen on :54264
log file created: gologdemo.log
2020-07-17 17:01:42.968    INFO 7002 --- [   19] [-]           main.go:34 : {"contemtType":"","proto":"HTTP/1.1"} [::1]:56946 GET /abc
2020-07-17 17:01:45.974    INFO 7002 --- [   34] [-]           main.go:34 : {"contemtType":"","proto":"HTTP/1.1"} [::1]:56958 GET /abc
2020-07-17 17:01:46.977    INFO 7002 --- [   21] [-]           main.go:34 : {"contemtType":"","proto":"HTTP/1.1"} [::1]:56963 GET /abc
2020-07-17 17:01:47.900    INFO 7002 --- [    5] [-]           main.go:34 : {"contemtType":"","proto":"HTTP/1.1"} [::1]:56968 GET /abc
```

## Log Creation Sequence

```
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

## Help shell

1. `sed "s/\x1B\[([0-9]{1,2}(;[0-9]{1,2})?)?[m|K]//g" x.log` to strip color from log file.
