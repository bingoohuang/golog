package golog_test

import (
	"testing"
	"time"

	"github.com/bingoohuang/golog/pkg/spec"

	"github.com/bingoohuang/golog/pkg/timex"

	"github.com/bingoohuang/golog"
	"github.com/bingoohuang/golog/pkg/rotate"
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
	s := "level=info,file=a.log,rotate=.yyyy-MM-dd,gzipAge=3d,maxSize=100M,printColor,stdout=true,printCaller"
	logSpec := golog.LogSpec{}

	assert.Nil(t, spec.ParseSpec(s, "spec", &logSpec))
	assert.Equal(t, golog.LogSpec{
		Level:       "info",
		File:        "a.log",
		Rotate:      ".2006-01-02",
		KeepAge:     30 * timex.Day,
		GzipAge:     3 * timex.Day,
		MaxSize:     100 * rotate.MiB,
		PrintColor:  true,
		Stdout:      true,
		PrintCaller: true,
	}, logSpec)
}
