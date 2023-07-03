package logfmt

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestParseLimitConf(t *testing.T) {
	conf, msg := ParseLimitConf([]byte(`[L:100,15s:ignore.sync] to limit 1 message every 15 seconds or every 100 messages with "ignore.sync" as key`))
	assert.Equal(t, LimitConf{EveryNum: 100, EveryTime: 15 * time.Second, Key: "ignore.sync", Level: logrus.InfoLevel}, *conf)
	assert.Equal(t, `to limit 1 message every 15 seconds or every 100 messages with "ignore.sync" as key`, string(msg))

	conf, msg = ParseLimitConf([]byte(`[L:15s:ignore.sync]      to limit 1 message every 15 seconds with "ignore.sync" as key`))
	assert.Equal(t, LimitConf{EveryNum: 0, EveryTime: 15 * time.Second, Key: "ignore.sync", Level: logrus.InfoLevel}, *conf)
	assert.Equal(t, `to limit 1 message every 15 seconds with "ignore.sync" as key`, string(msg))

	conf, msg = ParseLimitConf([]byte(`[L:100,15s]  to limit 1 message every 15 seconds or every 100 messages with the first two words in the message as key`))
	assert.Equal(t, LimitConf{EveryNum: 100, EveryTime: 15 * time.Second, Key: "to limit", Level: logrus.InfoLevel}, *conf)
	assert.Equal(t, `to limit 1 message every 15 seconds or every 100 messages with the first two words in the message as key`, string(msg))

	conf, msg = ParseLimitConf([]byte(`[L:100,0s]  to limit 1 message every 100 messages with the first two words in the message as key`))
	assert.Equal(t, LimitConf{EveryNum: 100, EveryTime: 0, Key: "to limit", Level: logrus.InfoLevel}, *conf)
	assert.Equal(t, `to limit 1 message every 100 messages with the first two words in the message as key`, string(msg))

	conf, msg = ParseLimitConf([]byte(`[L:15s]      to limit 1 message every 15 seconds with the first two words in the message as key`))
	assert.Equal(t, LimitConf{EveryNum: 0, EveryTime: 15 * time.Second, Key: "to limit", Level: logrus.InfoLevel}, *conf)
	assert.Equal(t, `to limit 1 message every 15 seconds with the first two words in the message as key`, string(msg))

	RegisterLimitConf(LimitConf{
		EveryNum:  0,
		EveryTime: 15 * time.Second,
		Key:       "LimitConf1",
		Level:     logrus.InfoLevel,
	})
	conf, msg = ParseLimitConf([]byte(`[L:LimitConf1]      to limit using configuration whose name is LimitConf1`))
	assert.Equal(t, LimitConf{EveryNum: 0, EveryTime: 15 * time.Second, Key: "LimitConf1", Level: logrus.InfoLevel}, *conf)
	assert.Equal(t, `to limit using configuration whose name is LimitConf1`, string(msg))
}
