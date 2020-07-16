package log

import (
	"flag"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/mitchellh/go-homedir"
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

	logdir, _ = homedir.Expand(logdir)

	viper.SetDefault(LogMaxBackupsDaysKey, 7)
	viper.SetDefault(LogDebugKey, false)
	viper.SetDefault(LogTimeFormatKey, "%Y-%m-%d")

	logrusOption := LogrusOption{
		Level:            viper.GetString(LoglevelKey),
		PrintColors:      true,
		NoCaller:         false,
		Stdout:           viper.GetBool(LogrusKey),
		LogPath:          logdir + ".log." + viper.GetString(LogTimeFormatKey),
		RotateExtPattern: "",
	}

	return logrusOption.Setup()
}
