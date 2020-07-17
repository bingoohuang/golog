package golog_test

import (
	"github.com/bingoohuang/golog"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"testing"
	"time"
)

func TestSetupLogrus(t *testing.T) {
	viper.Set(golog.LoglevelKey, "DEBUG")
	viper.Set(golog.LogTimeFormatKey, ".2006-01-02-15-04-05")
	golog.SetupLogrus(nil)

	for i := 0; i < 10; i++ {
		logrus.Warnf("这是警告信息 %d", i)
		logrus.Infof("这是普通信息 %d", i)
		logrus.Debugf("这是调试信息 %d", i)

		time.Sleep(1 * time.Second)
	}
}
