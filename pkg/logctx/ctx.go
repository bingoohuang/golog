package logctx

import "sync"

var ctxMap map[string]string
var ctxMapLock sync.Mutex

func Get(key string) (string, bool) {
	ctxMapLock.Lock()
	defer ctxMapLock.Unlock()

	v, ok := ctxMap[key]
	return v, ok
}

func Set(key, val string, kvs ...string) {
	ctxMapLock.Lock()
	defer ctxMapLock.Unlock()

	ctxMap = map[string]string{}
	ctxMap[key] = val

	for i := 0; i+1 < len(kvs); i += 2 {
		ctxMap[kvs[i]] = kvs[i+1]
	}
}

func GetVars() map[string]string {
	ctxMapLock.Lock()
	defer ctxMapLock.Unlock()

	m := map[string]string{}
	for k, v := range ctxMap {
		m[k] = v
	}

	return m
}
