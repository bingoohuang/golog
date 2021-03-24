package golog

import (
	"context"
	"github.com/bingoohuang/golog/pkg/homedir"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/time/rate"

	"github.com/bingoohuang/golog/pkg/logfmt"

	"github.com/bingoohuang/golog/pkg/spec"

	"github.com/sirupsen/logrus"
)

// SetupOption defines the options to setup.
type SetupOption struct {
	Spec   string
	Layout string
	Logger *logrus.Logger
}

type (
	// SetupOptionFn is func type to option setter
	SetupOptionFn func(*SetupOption)
	// SetupOptionFns is the slice of SetupOptionFn
	SetupOptionFns []SetupOptionFn
)

// Spec defines the specification of log.
func Spec(v string) SetupOptionFn { return func(o *SetupOption) { o.Spec = v } }

// Layout defines the layout of log.
func Layout(v string) SetupOptionFn { return func(o *SetupOption) { o.Layout = v } }

// Logger defines the root logrus logger.
func Logger(v *logrus.Logger) SetupOptionFn { return func(o *SetupOption) { o.Logger = v } }

// Setup setups the options.
func (fns SetupOptionFns) Setup(o *SetupOption) {
	for _, f := range fns {
		f(o)
	}
}

// DisableLogging disable all logrus logging and standard logging.
func DisableLogging() {
	logrus.SetLevel(logrus.PanicLevel)
	logrus.SetOutput(io.Discard)
	log.SetOutput(io.Discard)
	log.SetFlags(0)
}

// SetupLogrus setup the logrus logger with specific configuration like guava CacheBuilderSpec.
// eg: "level=info,file=a.log,rotate=yyyy-MM-dd,maxAge=30d,gzipAge=3d,maxSize=100M,printColor,stdout,printCaller"
func SetupLogrus(fns ...SetupOptionFn) *logfmt.Result {
	o := &SetupOption{}
	SetupOptionFns(fns).Setup(o)

	logSpec := &LogSpec{}
	if err := spec.ParseSpec(o.Spec, "spec", logSpec, spec.WithEnvPrefix("GOLOG")); err != nil {
		panic(err)
	}

	logrusOption := logfmt.LogrusOption{
		Level:       logSpec.Level,
		LogPath:     createLogDir(logSpec),
		Rotate:      string(logSpec.Rotate),
		MaxAge:      logSpec.MaxAge,
		GzipAge:     logSpec.GzipAge,
		MaxSize:     int64(logSpec.MaxSize),
		PrintColor:  logSpec.PrintColor,
		PrintCaller: logSpec.PrintCaller,
		Stdout:      logSpec.Stdout,
		Simple:      logSpec.Simple,
		Layout:      o.Layout,
		FixStd:      logSpec.FixStd,
	}

	return logrusOption.Setup(o.Logger)
}

func createLogDir(logSpec *LogSpec) string {
	logDir := ""
	logPath := logSpec.File
	appName := filepath.Base(os.Args[0])

	if logPath == "" {
		if CheckPrivileges() {
			logDir = filepath.Join("/var/log/", appName)
		} else {
			logDir = filepath.Join("~/logs/" + appName)
		}

		logPath = filepath.Join(logDir, appName+".log")
	} else {
		stat, err := os.Stat(logPath)
		if err == nil && stat.IsDir() || strings.ToLower(filepath.Ext(logPath)) != ".log" {
			// treat logPath as a log directory
			logPath = filepath.Join(logPath, appName+".log")
		}

		logDir = filepath.Dir(logPath)
	}

	stat, err := os.Stat(logPath)
	if err == nil && stat.IsDir() {
		return logPath
	}

	logDir, err = homedir.Expand(logDir)
	if err != nil {
		panic(err)
	}

	syscall.Umask(0)
	if err := os.MkdirAll(logDir, os.ModeSticky|os.ModePerm); err != nil {
		panic(err)
	}

	return logPath
}

// CheckPrivileges checks root rights to use system service
func CheckPrivileges() bool {
	if out, err := exec.Command("id", "-g").Output(); err == nil {
		if gid, e := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 32); e == nil {
			return gid == 0
		}
	}
	return false
}

// LogSpec defines the spec structure to be mapped to the log specification.
type LogSpec struct {
	Level       string        `spec:"level,info"`
	File        string        `spec:"file"`
	Rotate      spec.Layout   `spec:"rotate,.yyyy-MM-dd"`
	MaxAge      time.Duration `spec:"maxAge,30d"`
	GzipAge     time.Duration `spec:"gzipAge,3d"`
	MaxSize     spec.Size     `spec:"maxSize,100M"`
	PrintColor  bool          `spec:"printColor,false"`
	PrintCaller bool          `spec:"printCall,false"`
	Stdout      bool          `spec:"stdout,false"`
	Simple      bool          `spec:"simple,false"`
	FixStd      bool          `spec:"fixstd,true"` // 是否增强log.Print...的输出
}

// NewLimitLog create a limited printf functor to log
// that allows events up to rate r and permits bursts of at most b tokens.
func NewLimitLog(logLines float64, interval time.Duration, burst int) func(string, ...interface{}) {
	rateLimiter := rate.NewLimiter(rate.Limit(logLines/interval.Seconds()), burst)

	return func(format string, v ...interface{}) {
		if rateLimiter.Allow() {
			log.Printf(format, v...)
		}
	}
}

// NewLimitLogrus create a limited logrus entry
// that allows events up to rate r and permits bursts of at most b tokens.
func NewLimitLogrus(v *logrus.Logger, logLines float64, interval time.Duration, burst int) *logrus.Entry {
	limiter := rate.NewLimiter(rate.Limit(logLines/interval.Seconds()), burst)
	if v == nil {
		v = logrus.StandardLogger()
	}

	ctx := context.WithValue(context.Background(), logfmt.RateLimiterKey, limiter)
	return v.WithContext(ctx)
}
