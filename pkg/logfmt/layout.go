package logfmt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/bingoohuang/golog/pkg/caller"
	"github.com/bingoohuang/golog/pkg/gid"
	"github.com/bingoohuang/golog/pkg/logctx"
	"github.com/bingoohuang/golog/pkg/spec"
	"github.com/bingoohuang/golog/pkg/str"
	"github.com/bingoohuang/golog/pkg/timex"
	"github.com/sirupsen/logrus"
)

// Layout describes the parsed layout of expression.
type Layout struct {
	Parts []Part
}

func (l *Layout) ResetForLogFile() *Layout {
	ps := make([]Part, len(l.Parts))
	for i, p := range l.Parts {
		if v, ok := p.(LogFileReset); ok {
			ps[i] = v.ResetForLogFile()
		} else {
			ps[i] = p
		}
	}

	return &Layout{Parts: ps}
}

func (l Layout) Append(b *bytes.Buffer, e Entry) {
	for _, p := range l.Parts {
		p.Append(b, e)
	}
}

type LogFileReset interface {
	ResetForLogFile() Part
}

type Part interface {
	Append(*bytes.Buffer, Entry)
}

type LiteralPart string

func (l LiteralPart) Append(b *bytes.Buffer, _ Entry) {
	b.WriteString(string(l))
}

func (l *Layout) addLiteralPart(part string) {
	l.addPart(LiteralPart(part))
}

func (l *Layout) addPart(p Part) {
	l.Parts = append(l.Parts, p)
}

// NewLayout creates a new layout from string expression.
func NewLayout(lo Option) (*Layout, error) {
	l := &Layout{}
	percentPos := 0
	layout := lo.Layout

	for layout != "" && percentPos >= 0 {
		percentPos = strings.Index(layout, "%")
		if percentPos < 0 {
			l.addLiteralPart(layout)
			continue
		}

		if percentPos > 0 {
			p := layout[:percentPos]
			if len(p) > 0 {
				l.addLiteralPart(layout[:percentPos])
			}
		}

		layout = layout[percentPos+1:]

		if strings.HasPrefix(layout, "%") {
			l.addLiteralPart("%")
			layout = layout[1:]
			continue
		}

		minus := false
		digits := ""
		indicator := ""
		options := ""
		var err error

		layout, minus = parseMinus(layout)
		layout, digits = parseDigits(layout)
		layout, indicator = parseIndicator(layout)
		layout, options, err = parseOptions(layout)
		if err != nil {
			return nil, err
		}

		p, err := lo.createPart(indicator, minus, digits, options)
		if err != nil {
			return nil, err
		}

		l.addPart(p)
	}

	return l, nil
}

func (lo Option) createPart(indicator string, minus bool, digits, options string) (Part, error) {
	switch indicator {
	case "t", "time":
		return parseTime(minus, digits, options)
	case "l", "level":
		return lo.parseLevel(minus, digits, options)
	case "pid":
		return parsePid(minus, digits, options)
	case "context":
		return parseContext(minus, digits, options)
	case "gid":
		return parseGid(minus, digits, options)
	case "trace":
		return parseTrace(minus, digits, options)
	case "caller":
		return parseCaller(minus, digits, options)
	case "fields":
		return parseFields(minus, digits, options)
	case "message", "msg", "m":
		return parseMessage(minus, digits, options)
	case "n":
		return parseNewLine(minus, digits, options)
	}

	return nil, fmt.Errorf("unknown indicator %s", indicator)
}

type NewLinePart struct{}

func (n NewLinePart) Append(b *bytes.Buffer, e Entry) {
	b.WriteString("\n")
}

func parseNewLine(minus bool, digits string, options string) (Part, error) {
	return NewLinePart{}, nil
}

type MessagePart struct {
	SingleLine bool
}

func (p MessagePart) Append(b *bytes.Buffer, e Entry) {
	// indent multiple lines log
	msg := e.Message()
	msg = strings.TrimRight(msg, "\r\n")

	if p.SingleLine {
		// indent multiple lines log
		b.WriteString(strings.ReplaceAll(msg, "\n", `\n`))
	} else {
		b.WriteString(msg)
	}
}

func parseMessage(minus bool, digits string, options string) (Part, error) {
	p := MessagePart{SingleLine: true}

	fields := strings.FieldsFunc(options, func(c rune) bool {
		return unicode.IsSpace(c) || c == ','
	})

	for _, f := range fields {
		parts := strings.SplitN(f, "=", 2)
		k := ""
		v := ""
		if len(parts) > 0 {
			k = strings.ToLower(parts[0])
		}

		if len(parts) > 1 {
			v = parts[1]
		}

		if k == "singleline" {
			p.SingleLine = str.ParseBool(v, true)
		}

	}
	return p, nil
}

type FieldsPart struct{}

func (p FieldsPart) Append(b *bytes.Buffer, e Entry) {
	if fields := e.Fields(); len(fields) > 0 {
		if v, err := json.Marshal(fields); err == nil {
			b.Write(v)
		}
	}
}

func parseFields(minus bool, digits string, options string) (Part, error) {
	return FieldsPart{}, nil
}

type ContextPart struct {
	Name   string
	Digits string
}

func (p ContextPart) Append(b *bytes.Buffer, e Entry) {
	v, _ := logctx.Get(p.Name)
	b.WriteString(fmt.Sprintf("%"+p.Digits+"s", v))
}

func parseContext(minus bool, digits string, options string) (Part, error) {
	c := ContextPart{Digits: compositeDigits(minus, digits, "")}

	fields := strings.FieldsFunc(options, func(c rune) bool {
		return unicode.IsSpace(c) || c == ','
	})

	name := ""

	for _, f := range fields {
		parts := strings.SplitN(f, "=", 2)
		k, v := "", ""
		if len(parts) > 0 {
			k = strings.ToLower(parts[0])
		}
		if len(parts) > 1 {
			v = parts[1]
		}

		if k == "name" {
			name = v
		}
	}

	if name == "" {
		return nil, errors.New("name required for %var")
	}

	c.Name = name
	return c, nil
}

type CallerPart struct {
	Digits string
	Sep    string
	skip   int
	Level  logrus.Level
}

func (p CallerPart) Append(b *bytes.Buffer, e Entry) {
	ll, _ := logrus.ParseLevel(e.Level())
	if ll > p.Level {
		return
	}

	fileLine := "-"
	callSkip := p.skip
	if v, ok := e.Fields()[caller.Skip]; ok {
		callSkip = v.(int)
		delete(e.Fields(), caller.Skip)
	}

	for i := 0; i < callSkip; i++ {
		if c := caller.GetCaller(i, "github.com/sirupsen/logrus"); c != nil {
			fileLine = fmt.Sprintf("\n%d%s%s %s%s%d ", i+1, p.Sep, filepath.Base(c.Function), filepath.Base(c.File), p.Sep, c.Line)
			b.WriteString(fmt.Sprintf("%"+p.Digits+"s", fileLine))
		}
	}
}

func parseCaller(minus bool, digits string, options string) (Part, error) {
	c := CallerPart{Digits: compositeDigits(minus, digits, "")}

	fields := strings.FieldsFunc(options, func(c rune) bool {
		return unicode.IsSpace(c) || c == ','
	})

	level := ""

	for _, f := range fields {
		parts := strings.SplitN(f, "=", 2)
		k, v := "", ""
		if len(parts) > 0 {
			k = strings.ToLower(parts[0])
		}
		if len(parts) > 1 {
			v = parts[1]
		}

		switch k {
		case "level":
			level = v
		case "sep":
			c.Sep = v
		case "skip":
			c.skip, _ = strconv.Atoi(v)
		}
	}

	c.Level, _ = logrus.ParseLevel(str.Or(level, "warn"))
	c.Sep = str.Or(c.Sep, ":")

	return c, nil
}

type TracePart struct {
	Digits string
}

func (t TracePart) Append(b *bytes.Buffer, e Entry) {
	b.WriteString(fmt.Sprintf("%"+t.Digits+"s", e.TraceID()))
}

func parseTrace(minus bool, digits string, options string) (Part, error) {
	return TracePart{Digits: compositeDigits(minus, digits, "")}, nil
}

type GidPart struct {
	Digits string
}

func (p GidPart) Append(b *bytes.Buffer, e Entry) {
	b.WriteString(fmt.Sprintf("%"+p.Digits+"s", gid.CurGoroutineID()))
}

func parseGid(minus bool, digits string, options string) (Part, error) {
	return GidPart{Digits: compositeDigits(minus, digits, "")}, nil
}

type PidPart struct {
	Digits string
}

func (p PidPart) Append(b *bytes.Buffer, e Entry) {
	b.WriteString(fmt.Sprintf("%d", Pid))
}

func parsePid(minus bool, digits string, options string) (Part, error) {
	return PidPart{Digits: compositeDigits(minus, digits, "")}, nil
}

type LevelPart struct {
	Digits     string
	PrintColor bool
	LowerCase  bool
	Length     int
}

func (l LevelPart) ResetForLogFile() Part {
	l.PrintColor = false
	return l
}

func (l LevelPart) Append(b *bytes.Buffer, e Entry) {
	lvl := strings.ToUpper(str.Or(e.Level(), "info"))

	if l.PrintColor {
		_, _ = fmt.Fprintf(b, "\x1b[%dm", ColorByLevel(lvl))
	}

	if strings.ToLower(lvl) == "warning" {
		lvl = lvl[:4]
	}

	if l.Length > 0 && len(lvl) > l.Length {
		lvl = lvl[:l.Length]
	}

	if l.LowerCase {
		lvl = strings.ToLower(lvl)
	}

	b.WriteString(fmt.Sprintf("%"+l.Digits+"s", lvl))

	if l.PrintColor { // reset
		b.WriteString("\x1b[0m")
	}
}

func (lo Option) parseLevel(minus bool, digits string, options string) (Part, error) {
	l := LevelPart{Digits: compositeDigits(minus, digits, "5"), PrintColor: lo.PrintColor}

	fields := strings.FieldsFunc(options, func(c rune) bool {
		return unicode.IsSpace(c) || c == ','
	})

	for _, f := range fields {
		parts := strings.SplitN(f, "=", 2)
		k := ""
		v := ""
		if len(parts) > 0 {
			k = strings.ToLower(parts[0])
		}

		if len(parts) > 1 {
			v = parts[1]
		}

		switch k {
		case "printcolor":
			l.PrintColor = str.ParseBool(v, false)
		case "lowercase":
			l.LowerCase = str.ParseBool(v, false)
		case "length":
			l.Length = str.ParseInt(v, 0)
		}
	}

	return l, nil
}

func compositeDigits(minus bool, digits, defaultValue string) string {
	if digits == "" && defaultValue != "" {
		digits = defaultValue
	}

	if minus {
		digits = "-" + digits
	}

	return digits
}

type Time struct {
	Format string
}

func (t Time) Append(b *bytes.Buffer, e Entry) {
	b.WriteString(timex.OrNow(e.Time()).Format(t.Format))
}

func parseTime(minus bool, digits string, options string) (Part, error) {
	return Time{Format: spec.ConvertTimeLayout(str.Or(options, "2006-01-02 15:04:05.000"))}, nil
}

func parseMinus(layout string) (string, bool) {
	if strings.HasPrefix(layout, "-") {
		return layout[1:], true
	}

	return layout, false
}

func parseDigits(layout string) (string, string) {
	digits := ""
	j := 0

	for i, r := range layout {
		j = i
		if unicode.IsDigit(r) || r == '.' {
			digits += string(r)
			j++
		} else {
			break
		}
	}

	return layout[j:], digits
}

func parseOptions(layout string) (string, string, error) {
	if !strings.HasPrefix(layout, "{") {
		return layout, "", nil
	}

	rightPos := strings.Index(layout, "}")
	if rightPos < 0 {
		return "", "", fmt.Errorf("bad layout, unclosed brace")
	}

	return layout[rightPos+1:], layout[1:rightPos], nil
}

func parseIndicator(layout string) (string, string) {
	indicator := ""
	j := 0

	for i, r := range layout {
		j = i
		if unicode.IsLetter(r) {
			indicator += string(r)
			j++
		} else {
			break
		}
	}

	return layout[j:], indicator
}
