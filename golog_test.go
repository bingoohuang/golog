package golog_test

import (
	"testing"

	"github.com/bingoohuang/golog"
	"github.com/sirupsen/logrus"
)

func TestSetupLogrus(t *testing.T) {
	_ = golog.Setup(golog.Spec("level=debug,rotate=.yyyy-mm-dd-HH-mm-ss,maxAge=5s,gzipAge=3s"))

	for i := 0; i < 10; i++ {
		logrus.Warnf("这是警告信息 %d", i)
		logrus.Infof("这是普通信息 %d", i)
		logrus.Debugf("这是调试信息 %d", i)
	}
}

func TestSetupLogrusLayout(t *testing.T) {
	layout := `%t{HH:mm:ss.SSS} %5l{length=5} %pid --- [GID=%gid] [%trace] %caller : %fields %msg%n`
	_ = golog.Setup(golog.Spec("level=debug,rotate=.yyyy-mm-dd-HH-mm"), golog.Layout(layout))

	for i := 0; i < 10; i++ {
		logrus.Warnf("这是警告信息 %d", i)
		logrus.Infof("这是普通信息 %d", i)
		logrus.Debugf("这是调试信息 %d", i)
	}
}
