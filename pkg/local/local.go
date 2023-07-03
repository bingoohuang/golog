package local

import (
	"context"
	"sync"
)

// nolint gochecknoglobals
var locals = &struct {
	ctx map[uint64]context.Context
	sync.RWMutex
}{
	ctx: make(map[uint64]context.Context),
}

func get(gid uint64) context.Context {
	locals.RLock()
	ctx := locals.ctx[gid]
	locals.RUnlock()

	if ctx == nil {
		ctx = context.Background()
	}

	return ctx
}

// nolint golint
func set(ctx context.Context, gid uint64) {
	locals.Lock()
	locals.ctx[gid] = ctx
	locals.Unlock()
}

func temp(gid uint64, key, val interface{}) context.Context {
	ctx := context.WithValue(get(gid), key, val)
	set(ctx, gid)

	return ctx
}

func clear(gid uint64) context.Context {
	locals.Lock()
	ctx := locals.ctx[gid]
	delete(locals.ctx, gid)
	locals.Unlock()

	return ctx
}

// Get ...
func Get() context.Context {
	return get(Goid())
}

// Set ...
func Set(ctx context.Context) {
	set(ctx, Goid())
}

// ...
func Clear() context.Context {
	return clear(Goid())
}

// Value ...
func Temp(key, val interface{}) context.Context {
	return temp(Goid(), key, val)
}

// Value ...
func Value(key interface{}) interface{} {
	ctx := Get()
	if ctx == nil {
		return nil
	}

	return ctx.Value(key)
}

// Go ...
func Go(fn func()) {
	GoContext(Get(), fn)
}

// GoContext ...
func GoContext(ctx context.Context, fn func()) {
	go func() {
		Set(ctx)

		defer Clear()

		fn()
	}()
}

type Key int

const (
	TraceId Key = iota
)
