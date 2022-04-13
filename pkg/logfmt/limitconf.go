package logfmt

import (
	"bytes"
	"regexp"
	"strconv"
	"sync"
	"time"
	"unicode"

	"github.com/bingoohuang/golog/pkg/caller"
	"github.com/bingoohuang/golog/pkg/gid"
	"github.com/bingoohuang/golog/pkg/stack"
	"github.com/sirupsen/logrus"
)

var (
	// reglimitTip parses limit tip, e.g.
	// [L:100,15s:ignore.sync]  to limit 1 message every 15 seconds  or every 100 messages with "ignore.sync" as key
	// [L:15s:ignore.sync]      to limit 1 message every 15 seconds with "ignore.sync" as key
	// [L:100,15s]  to limit 1 message every 15 seconds or every 100 messages with the first two words in the message as key
	// [L:100,0s]  to limit 1 message every 100 messages with the first two words in the message as key
	// [L:15s]      to limit 1 message every 15 seconds with the first two words in the message as key
	// [L:LimitConf1]      to limit using configuration whose name is LimitConf1
	reglimitTip = regexp.MustCompile(`\[L:[\w\d-.:,\s]+]`)               // https://regex101.com/r/LTjduP/4
	reglimitSep = regexp.MustCompile(`(\d+,)?(\d+\w{0,2})(:[\w\d-.]+)?`) // https://regex101.com/r/kfQJrN/3
)

type LimitConf struct {
	EveryNum  int
	EveryTime time.Duration
	Key       string
	Level     logrus.Level
}

type limitRuntime struct {
	conf *LimitConf
	num  int
	msg  []byte
	sync.Mutex
	level       logrus.Level
	ll          *logrus.Logger
	goroutineID gid.GoroutineID
	call        *stack.Call
}

func (r *limitRuntime) run() {
	if r.conf.EveryTime == 0 {
		return
	}

	ticker := time.NewTicker(r.conf.EveryTime)
	for {
		r.sendMsg()
		<-ticker.C
	}
}

func (r *limitRuntime) SendNewMsg(ll *logrus.Logger, level logrus.Level, msg []byte, formatter *LogrusFormatter) {
	r.Lock()
	defer r.Unlock()

	if r.conf.EveryNum > 0 {
		if r.num == r.conf.EveryNum {
			r.num = 0
		}

		if r.num == 0 {
			ll.WithField(caller.Skip, 13).Log(level, string(msg))
			r.msg = nil
			r.num++
			return
		}
	}

	r.num++
	r.msg = msg
	r.level = level
	r.ll = ll
	r.goroutineID = gid.CurGoroutineID()

	if formatter.PrintCaller {
		call := stack.Caller(5)
		r.call = &call
	}
}

func (r *limitRuntime) sendMsg() {
	r.Lock()
	defer r.Unlock()

	if len(r.msg) > 0 {
		r.ll.
			WithField(caller.Skip, -1).
			WithField(caller.GidKey, r.goroutineID).
			WithField(caller.CallerKey, r.call).
			Log(r.level, string(r.msg))
		r.msg = nil
	}
}

var (
	limiter         = map[string]*limitRuntime{}
	limiterLock     sync.Mutex
	limiterRegistry = map[string]*LimitConf{}
)

func Limit(ll *logrus.Logger, level logrus.Level, msg []byte, formatter *LogrusFormatter) (filteredMsg []byte, limited bool) {
	conf, s := ParseLimitConf(msg)
	if conf == nil {
		return msg, false
	}

	if level < conf.Level {
		return s, false
	}

	limiterLock.Lock()
	defer limiterLock.Unlock()

	rt := limiter[conf.Key]
	if rt == nil {
		rt = &limitRuntime{conf: conf}
		limiter[conf.Key] = rt

		go rt.run()
	}

	rt.SendNewMsg(ll, level, s, formatter)
	return nil, true
}

func getLimitConf(key string) *LimitConf {
	limiterLock.Lock()
	defer limiterLock.Unlock()

	return limiterRegistry[key]
}

func RegisterLimitConf(limitConf LimitConf) {
	if limitConf.Level == logrus.PanicLevel {
		limitConf.Level = logrus.InfoLevel
	}

	limiterLock.Lock()
	defer limiterLock.Unlock()

	limiterRegistry[limitConf.Key] = &limitConf
}

func ParseLimitConf(msg []byte) (*LimitConf, []byte) {
	catches := reglimitTip.FindIndex(msg)
	if len(catches) == 0 {
		return nil, msg
	}

	x, y := catches[0], catches[1]
	confValue := msg[x:y]
	confValue = bytes.TrimSpace(confValue[3 : len(confValue)-1]) // get xyz from [L:xyz]
	confStr := string(confValue)

	newMsg := clearMsg(msg, x, y)
	conf := getLimitConf(confStr)
	if conf != nil {
		return conf, newMsg
	}

	subs := reglimitSep.FindStringSubmatch(confStr)
	if len(subs) == 0 {
		return nil, newMsg
	}

	everyNumVal, everyTimeVal, keyVal := subs[1], subs[2], subs[3]
	everyNum := 0
	if everyNumVal != "" {
		everyNumVal = everyNumVal[:len(everyNumVal)-1]
		everyNum, _ = strconv.Atoi(everyNumVal)
	}

	everyTime, err := time.ParseDuration(everyTimeVal)
	if err != nil {
		return nil, newMsg
	}

	if keyVal != "" {
		keyVal = keyVal[1:]
	} else {
		spaceCount := 0
		if idx := bytes.IndexFunc(newMsg, func(r rune) bool {
			if unicode.IsSpace(r) {
				spaceCount++
			}
			return spaceCount >= 2
		}); idx < 0 {
			keyVal = string(newMsg)
		} else {
			keyVal = string(newMsg[:idx])
		}
	}

	return &LimitConf{
		EveryNum:  everyNum,
		EveryTime: everyTime,
		Key:       keyVal,
		Level:     logrus.InfoLevel,
	}, newMsg
}
