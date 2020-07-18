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

watch the log file rotating and gzipping and deleting `watch -c "ls -tlh  gologdemo.log*"`

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
1. [`tail -F x.log`](https://explainshell.com/explain?cmd=tail+-F+x.log) even if x.log recreated.

