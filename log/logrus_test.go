package log_test

import (
	"testing"

	"github.com/bingoohuang/golog/log"
	"github.com/sirupsen/logrus"
)

func TestSetupLogrus(t *testing.T) {
	log.LogrusSetup()

	logrus.Warnf("这是警告信息")
	logrus.Infof("这是普通信息")
	logrus.Debugf("这是调试信息")
}
