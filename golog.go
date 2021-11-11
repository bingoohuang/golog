package golog

import (
	"github.com/bingoohuang/golog/pkg/unmask"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bingoohuang/golog/pkg/homedir"

	"github.com/bingoohuang/golog/pkg/logfmt"

	"github.com/bingoohuang/golog/pkg/spec"

	"github.com/sirupsen/logrus"
)

// SetupOption defines the options to setup.
type SetupOption struct {
	Spec    string
	Layout  string
	LogPath string
	Logger  *logrus.Logger
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

// LogPath defines the log path.
func LogPath(v string) SetupOptionFn { return func(o *SetupOption) { o.LogPath = v } }

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

// Setup setup the logrus logger with specific configuration like guava CacheBuilderSpec.
// eg: "level=info,file=a.log,rotate=yyyy-MM-dd,maxAge=30d,gzipAge=3d,maxSize=100M,printColor,stdout,printCaller"
func Setup(fns ...SetupOptionFn) *logfmt.Result {
	o := SetupOption{}
	SetupOptionFns(fns).Setup(&o)
	logrusOption := o.InitiateOption()

	return logrusOption.Setup(o.Logger)
}

func (o SetupOption) InitiateOption() logfmt.LogrusOption {
	logSpec := &LogSpec{}
	if err := spec.ParseSpec(o.Spec, "spec", logSpec, spec.WithEnvPrefix("GOLOG")); err != nil {
		panic(err)
	}

	logrusOption := logfmt.LogrusOption{
		Level:       logSpec.Level,
		LogPath:     CreateLogDir(o.LogPath, logSpec),
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
	return logrusOption
}

func CreateLogDir(logPath string, logSpec *LogSpec) string {
	logDir := ""
	if logPath == "" {
		logPath = logSpec.File
	}
	appName := filepath.Base(os.Args[0])
	wd, _ := os.Getwd()
	wd = filepath.Base(wd)
	if logPath == "" {
		if CheckPrivileges() {
			logDir = filepath.Join("/var/log/", wd)
		} else {
			logDir = filepath.Join("~/logs/" + wd)
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

	unmask.Unmask()
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

// Printf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
// If the last argument is an error, the format will be prepended with "E!"
// for error level if there is no level tag defined in it.
func Printf(format string, v ...interface{}) {
	if len(v) > 0 {
		if _, ok := v[len(v)-1].(error); ok {
			if _, _, hasLevelTag := logfmt.ParseLevelFromMsg(format); !hasLevelTag {
				format = "E! " + format
			}
		}
	}
	log.Printf(format, v...)
}

// LimitConf defines the log limit configuration.
type LimitConf struct {
	EveryNum  int
	EveryTime time.Duration
	Key       string
}

// RegisterLimiter registers a limit for the log generation frequency.
func RegisterLimiter(c LimitConf) {
	logfmt.RegisterLimitConf(logfmt.LimitConf{
		EveryNum:  c.EveryNum,
		EveryTime: c.EveryTime,
		Key:       c.Key,
	})
}
