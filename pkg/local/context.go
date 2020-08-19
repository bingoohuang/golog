package local

import (
	"reflect"

	"github.com/bingoohuang/golog/pkg/gid"
)

func loadUint64(val reflect.Value) uint64 {
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return uint64(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint()
	}

	return 0
}

// Uint64 convert any number value to uint64.
func Uint64(key interface{}) uint64 {
	return loadUint64(reflect.ValueOf(Value(key)))
}

func Int64(key interface{}) int64 {
	return int64(Uint64(key))
}

// String return the local value as string.
func String(key interface{}) string {
	if v, ok := Value(key).(string); ok {
		return v
	}

	return ""
}

func Goid() uint64 {
	return gid.CurGoroutineID().Uint64()
}
