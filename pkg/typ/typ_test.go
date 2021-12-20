package typ_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/bingoohuang/golog/pkg/typ"
	"github.com/stretchr/testify/assert"
)

type Ger interface {
	Do()
}

type Mer struct{}

func (m *Mer) Do() {
}

func TestType(t *testing.T) {
	assert.False(t, typ.Implements(reflect.TypeOf(Mer{}), func(Ger) {}))
	assert.True(t, typ.PtrImplements(reflect.TypeOf(Mer{}), func(Ger) {}))

	assert.True(t, typ.Implements(reflect.TypeOf(time.Now()), func(fmt.Stringer) {}))
	assert.True(t, typ.PtrImplements(reflect.TypeOf(time.Now()), func(fmt.Stringer) {}))

	assert.True(t, typ.IsType(reflect.TypeOf(time.Second), func(time.Duration) {}))
}
