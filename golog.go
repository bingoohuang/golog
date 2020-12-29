package golog

import (
	"os"
	"path/filepath"
	"time"

	"github.com/bingoohuang/golog/pkg/logfmt"

	"github.com/bingoohuang/golog/pkg/str"

	"github.com/bingoohuang/golog/pkg/spec"

	"github.com/sirupsen/logrus"
)

type SetupOption struct {
	Spec   string
	Layout string
	Logger *logrus.Logger
}

type SetupOptionFn func(*SetupOption)
type SetupOptionFns []SetupOptionFn

func Spec(v string) SetupOptionFn           { return func(o *SetupOption) { o.Spec = v } }
func Layout(v string) SetupOptionFn         { return func(o *SetupOption) { o.Layout = v } }
func Logger(v *logrus.Logger) SetupOptionFn { return func(o *SetupOption) { o.Logger = v } }

func (fns SetupOptionFns) Setup(o *SetupOption) {
	for _, f := range fns {
		f(o)
	}
}

// SetupLogrus setup the logrus logger with specific configuration like guava CacheBuilderSpec.
// eg: "level=info,file=a.log,rotate=yyyy-MM-dd,maxAge=30d,gzipAge=3d,maxSize=100M,printColor,stdout,printCaller"
func SetupLogrus(fns ...SetupOptionFn) *logfmt.Result {
	o := &SetupOption{}
	SetupOptionFns(fns).Setup(o)

	logSpec := &LogSpec{}
	if err := spec.ParseSpec(o.Spec, "spec", logSpec); err != nil {
		panic(err)
	}

	logrusOption := logfmt.LogrusOption{
		Level:       logSpec.Level,
		LogPath:     str.Or(logSpec.File, "~/logs/"+filepath.Base(os.Args[0])+".log"),
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

// LogSpec defines the spec structure to be mapped to the log specification.
type LogSpec struct {
	Level       string        `spec:"level,info"`
	File        string        `spec:"file"`
	Rotate      spec.Layout   `spec:"rotate,.yyyy-MM-dd"`
	MaxAge      time.Duration `spec:"maxAge,30d"`
	GzipAge     time.Duration `spec:"gzipAge,3d"`
	MaxSize     spec.Size     `spec:"maxSize,100M"`
	PrintColor  bool          `spec:"printColor,true"`
	PrintCaller bool          `spec:"printCall,true"`
	Stdout      bool          `spec:"stdout,true"`
	Simple      bool          `spec:"simple,false"`
	FixStd      bool          `spec:"fixstd,true"` // 是否增强log.Print...的输出
}
