package golog

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bingoohuang/golog/pkg/logfmt"
	"github.com/bingoohuang/golog/pkg/rotate"
	"github.com/bingoohuang/golog/pkg/spec"
	"github.com/bingoohuang/golog/pkg/str"
	"github.com/bingoohuang/golog/pkg/term"
	"github.com/bingoohuang/golog/pkg/unmask"
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

// Setup set up the logrus logger with specific configuration like guava CacheBuilderSpec.
// eg: "level=info,file=a.log,rotate=yyyy-MM-dd,maxAge=30d,gzipAge=3d,maxSize=100M,printColor,stdout,printCaller"
func Setup(fns ...SetupOptionFn) *logfmt.Result {
	o := SetupOption{}
	SetupOptionFns(fns).Setup(&o)
	option := o.InitiateOption()
	return option.Setup(o.Logger)
}

func (o SetupOption) InitiateOption() logfmt.Option {
	l := &LogSpec{}
	if err := spec.ParseSpec(o.Spec, "spec", l, spec.WithEnvPrefix("GOLOG")); err != nil {
		panic(err)
	}

	stdout := false
	switch strings.ToLower(l.Stdout) {
	case "true", "1", "t", "yes", "y", "on":
		stdout = true
	case "false", "0", "f", "no", "n", "off":
		stdout = false
	default:
		stdout = term.IsTerminal()
	}
	opt := logfmt.Option{
		Level:        l.Level,
		LogPath:      CreateLogDir(o.LogPath, l),
		Rotate:       string(l.Rotate),
		MaxAge:       l.MaxAge,
		GzipAge:      l.GzipAge,
		MaxSize:      int64(l.MaxSize),
		TotalSizeCap: int64(l.TotalSizeCap),
		PrintColor:   l.PrintColor,
		PrintCaller:  l.PrintCaller,
		Stdout:       stdout,
		Simple:       l.Simple,
		Layout:       o.Layout,
		FixStd:       l.FixStd,
	}
	return opt
}

func CreateLogDir(logPath string, logSpec *LogSpec) string {
	if logPath == "" {
		logPath = logSpec.File
	}

	appName := filepath.Base(os.Args[0])

	if logPath == "" {
		parent := appName
		home, _ := os.UserHomeDir()
		if wd, _ := os.Getwd(); wd != "" && home != wd {
			if base := filepath.Base(wd); base != "bin" {
				parent = base
			}
		}
		logPath = filepath.Join("~/logs/", parent, appName+".log")
	} else {
		stat, err := os.Stat(logPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "logPath %s stat err: %v\n", logPath, err)
		}
		if stat != nil && stat.IsDir() || strings.ToLower(filepath.Ext(logPath)) != ".log" {
			// treat logPath as a log directory
			logPath = filepath.Join(logPath, appName+".log")
		}
	}

	if strings.HasPrefix(logPath, "~") {
		dir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "get user home directory, err: %v\n", err)
			dir = "."
		}

		logPath = dir + logPath[1:]
	}

	logPath = filepath.Clean(logPath)
	logDir := filepath.Dir(logPath)
	stat, err := os.Stat(logDir)
	if err == nil && stat.IsDir() {
		return logPath
	}

	unmask.Unmask()
	if err := os.MkdirAll(logDir, os.ModeSticky|os.ModePerm); err != nil {
		fmt.Fprintf(os.Stderr, "make log directory, err: %v\n", err)
		logPath = filepath.Base(logPath)
	}

	if rotate.GologDebug {
		fmt.Fprintf(os.Stderr, "logPath: %s\n", logPath)
	}

	return logPath
}

func ExecutableInCurrentDir() (bool, error) {
	ex, err := os.Executable()
	if err != nil {
		return false, err
	}

	workdirDir, err := os.Getwd()
	if err != nil {
		return false, err
	}

	return filepath.Dir(ex) == workdirDir, nil
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
	Level        string        `spec:"level,info"`
	File         string        `spec:"file"`
	Rotate       spec.Layout   `spec:"rotate,.yyyy-MM-dd"`
	MaxAge       time.Duration `spec:"maxAge,30d"`
	GzipAge      time.Duration `spec:"gzipAge,3d"`
	MaxSize      spec.Size     `spec:"maxSize,100M"`
	TotalSizeCap spec.Size     `spec:"totalSizeCap,1G"` // 可选，用来指定所有日志文件的总大小上限，例如设置为3GB的话，那么到了这个值，就会删除旧的日志
	PrintColor   bool          `spec:"printColor,false"`
	PrintCaller  bool          `spec:"printCall,false"`
	Stdout       string        `spec:"stdout"`
	Simple       bool          `spec:"simple,false"`
	FixStd       bool          `spec:"fixstd,true"` // 是否增强log.Print...的输出
}

// Printf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
// If the last argument is an error, the format will be prepended with "E!"
// for error level if there is no level tag defined in it.
func Printf(format string, v ...interface{}) {
	if len(v) > 0 {
		if _, ok := v[len(v)-1].(error); ok {
			if _, _, hasLevelTag := logfmt.ParseLevelFromMsg(str.ToBytes(format)); !hasLevelTag {
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
	Level     string
}

// RegisterLimiter registers a limit for the log generation frequency.
func RegisterLimiter(c LimitConf) {
	level, _ := logrus.ParseLevel(c.Level)
	logfmt.RegisterLimitConf(logfmt.LimitConf{
		EveryNum:  c.EveryNum,
		EveryTime: c.EveryTime,
		Key:       c.Key,
		Level:     level,
	})
}
