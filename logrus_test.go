package golog_test

import (
	"testing"
	"time"

	"github.com/bingoohuang/golog"
	"github.com/bingoohuang/golog/rotate"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSetupLogrus(t *testing.T) {
	golog.SetupLogrus(nil, "level=debug,rotate=.yyyy-mm-dd-HH-mm-ss")

	for i := 0; i < 10; i++ {
		logrus.Warnf("这是警告信息 %d", i)
		logrus.Infof("这是普通信息 %d", i)
		logrus.Debugf("这是调试信息 %d", i)

		time.Sleep(1 * time.Second)
	}
}

func TestParseSpec(t *testing.T) {
	spec := "level=info,file=a.log,rotate=.yyyy-MM-dd,gzipDays=3,maxSize=100M,printColor,stdout=true,printCaller"
	logSpec := golog.LogSpec{}

	assert.Nil(t, golog.ParseSpec(spec, "spec", &logSpec))
	assert.Equal(t, golog.LogSpec{
		Level:       "info",
		File:        "a.log",
		Rotate:      ".yyyy-MM-dd",
		KeepDays:    30,
		GzipDays:    3,
		MaxSize:     "100M",
		PrintColor:  true,
		Stdout:      true,
		PrintCaller: true,
	}, logSpec)

	assert.Equal(t, 100*rotate.MiB, logSpec.GetMaxSize())
	assert.Equal(t, ".2006-01-02", logSpec.GetRotate())
}
