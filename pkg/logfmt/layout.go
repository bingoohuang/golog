package logfmt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/bingoohuang/golog/pkg/caller"
	"github.com/bingoohuang/golog/pkg/gid"
	"github.com/bingoohuang/golog/pkg/spec"
	"github.com/bingoohuang/golog/pkg/str"
	"github.com/bingoohuang/golog/pkg/timex"
	"github.com/sirupsen/logrus"
)

// Layout describes the parsed layout of expression.
type Layout struct {
	Parts []Part
}

func (l Layout) Append(b *bytes.Buffer, e Entry) {
	for _, p := range l.Parts {
		p.Append(b, e)
	}
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
func NewLayout(lo LogrusOption) (*Layout, error) {
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
			l.addLiteralPart(strings.TrimSpace(layout[:percentPos]))
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

		var p Part

		switch indicator {
		case "t", "time":
			p, err = parseTime(minus, digits, options)
		case "l", "level":
			p, err = lo.parseLevel(minus, digits, options)
		case "pid":
			p, err = parsePid(minus, digits, options)
		case "gid":
			p, err = parseGid(minus, digits, options)
		case "trace":
			p, err = parseTrace(minus, digits, options)
		case "caller":
			p, err = parseCaller(minus, digits, options)
		case "fields":
			p, err = parseFields(minus, digits, options)
		case "message", "msg", "m":
			p, err = parseMessage(minus, digits, options)
		case "n":
			p, err = parseNewLine(minus, digits, options)
		}

		if err != nil {
			return nil, err
		}

		l.addPart(p)
	}

	return l, nil
}

type NewLinePart struct {
}

func (n NewLinePart) Append(b *bytes.Buffer, e Entry) {
	b.WriteString("\n")
}

func parseNewLine(minus bool, digits string, options string) (Part, error) {
	return &NewLinePart{}, nil
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
		b.WriteString(" " + strings.Replace(msg, "\n", `\n`, -1))
	} else {
		b.WriteString(" " + msg)
	}
}

func parseMessage(minus bool, digits string, options string) (Part, error) {
	p := &MessagePart{SingleLine: true}

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
		case "singleline":
			p.SingleLine = str.ParseBool(v, true)
		}

	}
	return p, nil
}

type FieldsPart struct {
}

func (p FieldsPart) Append(b *bytes.Buffer, e Entry) {
	if fields := e.Fields(); len(fields) > 0 {
		if v, err := json.Marshal(fields); err == nil {
			b.Write(v)
		}
	}
}

func parseFields(minus bool, digits string, options string) (Part, error) {
	return &FieldsPart{}, nil
}

type CallerPart struct {
	Digits string
	Level  logrus.Level
	Sep    string
}

func (p CallerPart) Append(b *bytes.Buffer, e Entry) {
	ll, _ := logrus.ParseLevel(e.Level())
	if ll > p.Level {
		return
	}

	fileLine := "-"
	c := e.Caller()
	if c == nil {
		c = caller.GetCaller()
	}

	if c != nil {
		fileLine = fmt.Sprintf("%s%s%d", filepath.Base(c.File), p.Sep, c.Line)
	}

	b.WriteString(fmt.Sprintf(" %"+p.Digits+"s", fileLine))
}

func parseCaller(minus bool, digits string, options string) (Part, error) {
	c := &CallerPart{Digits: compositeDigits(minus, digits, "")}

	fields := strings.FieldsFunc(options, func(c rune) bool {
		return unicode.IsSpace(c) || c == ','
	})

	level := ""

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
		case "level":
			level = v
		case "sep":
			c.Sep = v
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
	b.WriteString(fmt.Sprintf(" %"+t.Digits+"s", e.TraceID()))
}

func parseTrace(minus bool, digits string, options string) (Part, error) {
	p := &TracePart{Digits: compositeDigits(minus, digits, "")}

	return p, nil
}

type GidPart struct {
	Digits string
}

func (p GidPart) Append(b *bytes.Buffer, e Entry) {
	b.WriteString(fmt.Sprintf("%"+p.Digits+"s", gid.CurGoroutineID()))
}

func parseGid(minus bool, digits string, options string) (Part, error) {
	p := &GidPart{Digits: compositeDigits(minus, digits, "")}

	return p, nil
}

type PidPart struct {
	Digits string
}

func (p PidPart) Append(b *bytes.Buffer, e Entry) {
	b.WriteString(fmt.Sprintf(" %d ", Pid))
}

func parsePid(minus bool, digits string, options string) (Part, error) {
	p := &PidPart{Digits: compositeDigits(minus, digits, "")}

	return p, nil
}

type LevelPart struct {
	Digits     string
	PrintColor bool
	LowerCase  bool
	Length     int
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

func (lo LogrusOption) parseLevel(minus bool, digits string, options string) (Part, error) {
	l := &LevelPart{Digits: compositeDigits(minus, digits, "5"), PrintColor: lo.PrintColor}

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
		//case "printcolor":
		//	l.PrintColor = str.ParseBool(v, false)
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
	b.WriteString(timex.OrNow(e.Time()).Format(t.Format) + " ")
}

func parseTime(minus bool, digits string, options string) (Part, error) {
	format := spec.ConvertTimeLayout(str.Or(options, "2006-01-02 15:04:05.000"))
	return &Time{Format: format}, nil
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
