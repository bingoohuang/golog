package golog

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/bingoohuang/golog/log"
	"github.com/bingoohuang/golog/rotate"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// SetupLogrus setup the logrus logger with specific configuration like guava CacheBuilderSpec.
// eg: "level=info,file=a.log,rotate=yyyy-MM-dd,keepDays=30,gzipDays=3,maxSize=100M,printColor,stdout,printCaller"
func SetupLogrus(ll *logrus.Logger, spec string) io.Writer {
	logSpec := &LogSpec{}

	if err := ParseSpec(spec, "spec", logSpec); err != nil {
		panic(err)
	}

	logfile := logSpec.File
	if logfile == "" {
		logfile = "~/logs/" + filepath.Base(os.Args[0]) + ".log"
	}

	logrusOption := log.LogrusOption{
		Level:               logSpec.Level,
		PrintColors:         logSpec.PrintColor,
		PrintCaller:         logSpec.PrintCaller,
		Stdout:              logSpec.Stdout,
		LogPath:             logfile,
		RotatePostfixLayout: logSpec.GetRotate(),
	}

	fmt.Println("log file created:", logrusOption.LogPath)

	return logrusOption.Setup(ll)
}

// LogSpec defines the spec structure to be mapped to the log specification.
type LogSpec struct {
	Level       string `spec:"level,info"`
	File        string `spec:"file"`
	Rotate      string `spec:"rotate,yyyy-MM-dd"`
	KeepDays    int    `spec:"keepDays,30"`
	GzipDays    int    `spec:"gzipDays,3"`
	MaxSize     string `spec:"maxSize,100M"`
	PrintColor  bool   `spec:"printColor,true"`
	Stdout      bool   `spec:"stdout,true"`
	PrintCaller bool   `spec:"printCall,true"`
}

// GetMaxSize returns the maximum size.
func (l LogSpec) GetMaxSize() int {
	i := strings.IndexFunc(l.MaxSize, func(r rune) bool {
		return r < '0' || r > '9'
	})

	if i < 0 {
		size, _ := strconv.Atoi(l.MaxSize)
		return size
	}

	value, _ := strconv.Atoi(l.MaxSize[0:i])
	unit := l.MaxSize[i:]
	switch strings.ToUpper(unit) {
	case "K":
		return value * rotate.KiB
	case "M":
		return value * rotate.MiB
	case "G":
		return value * rotate.GiB
	}

	return value
}

// GetRotate returns the go-style rotate.
func (l LogSpec) GetRotate() string {
	s := l.Rotate

	s = strings.Replace(s, "yyyy", "2006", -1)
	s = strings.Replace(s, "yy", "06", -1)
	s = strings.Replace(s, "MM", "01", -1)
	s = strings.Replace(s, "dd", "02", -1)
	s = strings.Replace(s, "hh", "03", -1)
	s = strings.Replace(s, "HH", "15", -1)
	s = strings.Replace(s, "mm", "04", -1)
	s = strings.Replace(s, "ss", "05", -1)
	s = strings.Replace(s, "SSS", "000", -1)

	return s
}

// ParseSpec parses a specification to a structure.
func ParseSpec(spec, tagName string, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return errors.Errorf("v must be a pointer to a struct")
	}

	specMap := ParseSpecMap(spec)

	vv := rv.Elem()
	vt := vv.Type()

	for i := 0; i < vv.NumField(); i++ {
		ft := vt.Field(i)
		if ft.PkgPath != "" /*not exportable*/ || ft.Anonymous {
			continue
		}

		tag := ft.Tag.Get(tagName)
		if tag == "" {
			continue
		}

		specName, defaultValue := parseSpecTag(tag)

		setFieldSpec(vv.Field(i), specMap, specName, defaultValue)
	}

	return nil
}

func setFieldSpec(fv reflect.Value, specMap map[string]string, name, defaultValue string) {
	specValue, ok := specMap[name]
	if specValue == "" {
		specValue = defaultValue
	}

	switch ft := fv.Type(); ft.Kind() {
	case reflect.String:
		fv.SetString(specValue)
	case reflect.Int:
		v, _ := strconv.ParseInt(specValue, 10, 64)
		fv.SetInt(v)
	case reflect.Bool:
		if specValue == "" && ok {
			specValue = "true"
		}

		fv.SetBool(log.AnyOf(strings.ToLower(specValue), "true", "yes", "on", "1", "t"))
	default:

	}
}

func parseSpecTag(tag string) (string, string) {
	if i := strings.Index(tag, ","); i > 0 {
		return tag[:i], tag[i+1:]
	}

	return tag, ""
}

// ParseSpecMap parses the guava cache specification like string and returns
// a map listing the values specified for each key.
// ParseSpecMap always returns a non-nil map containing all the
// valid query parameters found; err describes the first decoding error
// encountered, if any.
//
// Query is expected to be a list of key=value settings separated by
// ampersands or semicolons or comma. A setting without an equals sign is
// interpreted as a key set to an empty value.
func ParseSpecMap(query string) map[string]string {
	m := make(map[string]string)

	for query != "" {
		key := query
		if i := strings.IndexAny(key, "&;,"); i >= 0 {
			key, query = key[:i], key[i+1:]
		} else {
			query = ""
		}

		if key == "" {
			continue
		}

		value := ""
		if i := strings.Index(key, "="); i >= 0 {
			key, value = key[:i], key[i+1:]
		}

		m[key] = value
	}

	return m
}
