file-Rotate
==================

Provide an `io.Writer` that periodically rotates log files from within the application. Port of [File::Rotate](https://metacpan.org/release/File-Rotate) from Perl to Go.

[![Build Status](https://travis-ci.org/lestrrat-go/file-rotate.png?branch=master)](https://travis-ci.org/lestrrat-go/file-Rotate)


# SYNOPSIS

```go
import (
  "log"
  "net/http"

  apachelog "github.com/lestrrat-go/apache-logformat"
  Rotate "github.com/bingoohuang/golog"
)

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { ... })

  logf, err := rotate.New(
    "/path/to/access_log.%Y%m%d%H%M",
    rotate.WithLinkName("/path/to/access_log"),
    rotate.WithMaxAge(24 * time.Hour),
    rotate.WithRotationTime(time.Hour),
  )
  if err != nil {
    log.Printf("failed to create Rotate: %s", err)
    return
  }

  // Now you must write to logf. apache-logformat library can create
  // a http.Handler that only writes the approriate logs for the request
  // to the given handle
  http.ListenAndServe(":8080", apachelog.CombinedLog.Wrap(mux, logf))
}
```

# DESCRIPTION

When you integrate this to into your app, it automatically write to logs that
are rotated from within the app: No more disk-full alerts because you forgot
to setup logrotate!

To install, simply issue a `go get`:

```
go get github.com/bingoohuang/golog
```

It's normally expected that this library is used with some other
logging service, such as the built-in `log` library, or loggers
such as `github.com/lestrrat-go/apache-logformat`.

```go
import(
  "log"
  "github.com/bingoohuang/golog"
)

func main() {
  rl, _ := rotate.New("/path/to/access_log.%Y%m%d%H%M")

  log.SetOutput(rl)

  /* elsewhere ... */
  log.Printf("Hello, World!")
}
```

OPTIONS
====

## Pattern (Required)

The pattern used to generate actual log file names. You should use patterns
using the strftime (3) format. For example:

```go
  rotate.New("/var/log/myapp/log.%Y%m%d")
```

## Clock (default: rotate.Local)

You may specify an object that implements the roatatelogs.Clock interface.
When this option is supplied, it's used to determine the current time to
base all of the calculations on. For example, if you want to base your
calculations in UTC, you may specify rotate.UTC

```go
  rotate.New(
    "/var/log/myapp/log.%Y%m%d",
    rotate.WithClock(rotate.UTC),
  )
```

## Location

This is an alternative to the `WithClock` option. Instead of providing an
explicit clock, you can provide a location for you times. We will create
a Clock object that produces times in your specified location, and configure
the rotatelog to respect it.

## LinkName (default: "")

Path where a symlink for the actual log file is placed. This allows you to
always check at the same location for log files even if the logs were rotated

```go
  rotate.New(
    "/var/log/myapp/log.%Y%m%d",
    rotate.WithLinkName("/var/log/myapp/current"),
  )
```

```
  // Else where
  $ tail -f /var/log/myapp/current
```

If not provided, no link will be written.

## RotationTime (default: 86400 sec (24 * 60 * 60))

Interval between file rotation. By default logs are rotated every 86400 seconds.
Note: Remember to use time.Duration values.

```go
  // Rotate every hour
  rotate.New(
    "/var/log/myapp/log.%Y%m%d",
    rotate.WithRotationTime(time.Hour),
  )
```

## MaxAge (default: 7 days)

Time to wait until old logs are purged. By default no logs are purged, which
certainly isn't what you want.
Note: Remember to use time.Duration values.

```go
  // Purge logs older than 1 hour
  rotate.New(
    "/var/log/myapp/log.%Y%m%d",
    rotate.WithMaxAge(time.Hour),
  )
```

## RotationCount (default: -1)

The number of files should be kept. By default, this option is disabled.

Note: MaxAge should be disabled by specifing `WithMaxAge(-1)` explicitly.

```go
  // Purge logs except latest 7 files
  rotate.New(
    "/var/log/myapp/log.%Y%m%d",
    rotate.WithMaxAge(-1),
    rotate.WithRotationCount(7),
  )
```

## Handler (default: nil)

Sets the event handler to receive event notifications from the Rotate
object. Currently only supported event type is FiledRotated

```go
  rotate.New(
    "/var/log/myapp/log.%Y%m%d",
    rotate.Handler(rotate.HandlerFunc(func(e Event) {
      if e.Type() != rotate.FileRotatedEventType {
        return
      }

      // Do what you want with the data. This is just an idea:
      storeLogFileToRemoteStorage(e.(*FileRotatedEvent).PreviousFile())
    })),
  )
```

## ForceNewFile

Ensure a new file is created every time New() is called. If the base file name
already exists, an implicit rotation is performed.

```go
  rotate.New(
    "/var/log/myapp/log.%Y%m%d",
    rotate.ForceNewFile(),
  )
```

# Rotating files forcefully

If you want to rotate files forcefully before the actual rotation time has reached,
you may use the `Rotate()` method. This method forcefully rotates the logs, but
if the generated file name clashes, then a numeric suffix is added so that
the new file will forcefully appear on disk.

For example, suppose you had a pattern of '%Y.log' with a rotation time of
`86400` so that it only gets rotated every year, but for whatever reason you
wanted to rotate the logs now, you could install a signal handler to
trigger this rotation:

```go
rl := rotate.New(...)

signal.Notify(ch, syscall.SIGHUP)

go func(ch chan os.Signal) {
  <-ch
  rl.Rotate()
}()
```

And you will get a log file name in like `2018.log.1`, `2018.log.2`, etc.
