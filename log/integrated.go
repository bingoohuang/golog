package log

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/spf13/pflag"
)

const (
	LoglevelKey = "loglevel"
	LogdirKey   = "logdir"
	LogrusKey   = "logrus"

	LogTimeFormatKey     = "logTimeFormat"
	LogMaxBackupsDaysKey = "logMaxBackupsDays"
	LogDebugKey          = "logDebug"
)

// PFlags declares the log pflags.
func PFlags() {
	pflag.StringP(LoglevelKey, "", "info", "debug/info/warn/error")
	pflag.StringP(LogdirKey, "", "", "log dir")
	pflag.BoolP(LogrusKey, "", true, "enable log to file")
}

// Flags declares the log flags.
func Flags() {
	flag.String(LoglevelKey, "info", "debug/info/warn/error")
	flag.String(LogdirKey, "", "log dir")
	flag.Bool(LogrusKey, true, "enable log to file")
}

func LogrusSetup() io.Writer {
	viper.SetDefault(LoglevelKey, "info")
	viper.SetDefault(LogrusKey, true)

	appName := filepath.Base(os.Args[0])
	logdir := viper.GetString(LogdirKey)

	if logdir == "" {
		logdir = "~/logs/" + appName
	}

	viper.SetDefault(LogMaxBackupsDaysKey, 7)
	viper.SetDefault(LogDebugKey, false)
	viper.SetDefault(LogTimeFormatKey, ".2006-01-02")

	logrusOption := LogrusOption{
		Level:               viper.GetString(LoglevelKey),
		PrintColors:         true,
		NoCaller:            false,
		Stdout:              viper.GetBool(LogrusKey),
		LogPath:             logdir + ".log",
		RotatePostfixLayout: viper.GetString(LogTimeFormatKey),
	}

	fmt.Println("log file created:", logrusOption.LogPath)

	return logrusOption.Setup()
}
