package golog

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/bingoohuang/golog/pkg/spec"

	"github.com/bingoohuang/golog/pkg/log"
	"github.com/sirupsen/logrus"
)

// SetupLogrus setup the logrus logger with specific configuration like guava CacheBuilderSpec.
// eg: "level=info,file=a.log,rotate=yyyy-MM-dd,maxAge=30d,gzipAge=3d,maxSize=100M,printColor,stdout,printCaller"
func SetupLogrus(ll *logrus.Logger, specification string) io.Writer {
	logSpec := &LogSpec{}

	if err := spec.ParseSpec(specification, "spec", logSpec); err != nil {
		panic(err)
	}

	logfile := logSpec.File
	if logfile == "" {
		logfile = "~/logs/" + filepath.Base(os.Args[0]) + ".log"
	}

	logrusOption := log.LogrusOption{
		Level:       logSpec.Level,
		LogPath:     logfile,
		Rotate:      string(logSpec.Rotate),
		MaxAge:      logSpec.MaxAge,
		GzipAge:     logSpec.GzipAge,
		MaxSize:     int(logSpec.MaxSize),
		PrintColor:  logSpec.PrintColor,
		PrintCaller: logSpec.PrintCaller,
		Stdout:      logSpec.Stdout,
	}

	fmt.Println("log file created:", logrusOption.LogPath)

	return logrusOption.Setup(ll)
}

// LogSpec defines the spec structure to be mapped to the log specification.
type LogSpec struct {
	Level       string        `spec:"level,info"`
	File        string        `spec:"file"`
	Rotate      spec.Layout   `spec:"rotate,yyyy-MM-dd"`
	MaxAge      time.Duration `spec:"maxAge,30d"`
	GzipAge     time.Duration `spec:"gzipAge,3d"`
	MaxSize     spec.Size     `spec:"maxSize,100M"`
	PrintColor  bool          `spec:"printColor,true"`
	PrintCaller bool          `spec:"printCall,true"`
	Stdout      bool          `spec:"stdout,true"`
}
