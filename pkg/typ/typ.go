package typ

import "reflect"

func Implements(t reflect.Type, i interface{}) bool {
	return t.Implements(reflect.TypeOf(i).In(0))
}

func PtrImplements(t reflect.Type, i interface{}) bool {
	return reflect.PtrTo(t).Implements(reflect.TypeOf(i).In(0))
}

func IsType(t reflect.Type, i interface{}) bool {
	return t == reflect.TypeOf(i).In(0)
}
