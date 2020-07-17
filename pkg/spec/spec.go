package spec

import (
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/bingoohuang/golog/pkg/typ"

	"github.com/bingoohuang/golog/pkg/str"
	"github.com/bingoohuang/golog/pkg/timex"
	"github.com/pkg/errors"
)

type Parser interface {
	Parse(string) error
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

		if err := setFieldSpec(vv.Field(i), specMap, specName, defaultValue); err != nil {
			return err
		}
	}

	return nil
}

func setFieldSpec(fv reflect.Value, specMap map[string]string, name, defaultValue string) error {
	specValue, ok := specMap[name]
	if specValue == "" {
		specValue = defaultValue
	}

	ftt := fv.Type()

	if typ.Implements(ftt, func(Parser) {}) { // 指针类型
		rv := reflect.New(ftt.Elem())
		if err := rv.Interface().(Parser).Parse(specValue); err != nil {
			return err
		}

		fv.Set(rv.Convert(ftt))
		return nil
	}

	if typ.PtrImplements(ftt, func(Parser) {}) { // 非指针类型
		rv := reflect.New(ftt)
		if err := rv.Interface().(Parser).Parse(specValue); err != nil {
			return err
		}

		fv.Set(rv.Elem().Convert(ftt))
		return nil
	}

	if typ.IsType(ftt, func(time.Duration) {}) {
		d, err := timex.ParseDuration(specValue)
		if err != nil {
			return err
		}

		fv.Set(reflect.ValueOf(d).Convert(ftt))
		return nil
	}

	switch ftt.Kind() {
	case reflect.String:
		fv.SetString(specValue)
	case reflect.Int:
		v, err := strconv.ParseInt(specValue, 10, 64)
		if err != nil {
			return err
		}

		fv.SetInt(v)
	case reflect.Bool:
		if specValue == "" && ok {
			specValue = "true"
		}

		fv.SetBool(str.AnyOf(strings.ToLower(specValue), "true", "yes", "on", "1", "t"))
	default:
		return errors.Errorf("unsupported field type %v", ftt)
	}

	return nil
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
