package golog

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bingoohuang/golog/pkg/logfmt"

	"github.com/bingoohuang/golog/pkg/str"

	"github.com/bingoohuang/golog/pkg/spec"

	"github.com/sirupsen/logrus"
)

// SetupLogrus setup the logrus logger with specific configuration like guava CacheBuilderSpec.
// eg: "level=info,file=a.log,rotate=yyyy-MM-dd,maxAge=30d,gzipAge=3d,maxSize=100M,printColor,stdout,printCaller"
func SetupLogrus(ll *logrus.Logger, specification, layout string) (*logfmt.Result, error) {
	logSpec := &LogSpec{}

	if err := spec.ParseSpec(specification, "spec", logSpec); err != nil {
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
		Layout:      layout,
		FixStd:      logSpec.FixStd,
	}

	fmt.Println("log file created:", logrusOption.LogPath)

	return logrusOption.Setup(ll)
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
