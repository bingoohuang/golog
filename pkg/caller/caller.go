package caller

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
)

const (
	maximumCallerDepth = 25
	// Skip is the key to set/get call skip value.
	Skip      = "_CallerSkip"
	GidKey    = "_CallerGid"
	CallerKey = "_CallerCaller"
)

var (
	minimumCallerDepth = 12
	// Used for caller information initialisation
	callerInitOnce sync.Once
)

// GetCaller retrieves the name of the first non-logrus calling function
func GetCaller(skip int, terminalPkg string) *runtime.Frame {
	// cache this package's fully-qualified name
	callerInitOnce.Do(func() {
		pcs := make([]uintptr, maximumCallerDepth)
		_ = runtime.Callers(0, pcs)

		// dynamic get the package name and the minimum caller depth
		for i := 0; i < maximumCallerDepth; i++ {
			funcName := runtime.FuncForPC(pcs[i]).Name()
			if strings.HasPrefix(funcName, terminalPkg) {
				minimumCallerDepth = i - 1
				break
			}
		}
	})

	// Restrict the lookback frames to avoid runaway lookups
	pcs := make([]uintptr, maximumCallerDepth)
	depth := runtime.Callers(minimumCallerDepth, pcs)
	frames := runtime.CallersFrames(pcs[:depth])

	for f, again := frames.Next(); again; f, again = frames.Next() {
		pkg := GetPackageName(f.Function)
		// If the caller isn't part of this package, we're done
		if strings.HasPrefix(pkg, terminalPkg) {
			continue
		}
		if skip != 0 {
			skip--
			continue
		}

		if f.Function != "github.com/bingoohuang/golog.NewLimitLog.func1" {
			return &f
		}
	}

	// if we got here, we failed to find the caller's context
	return nil
}

// PrintStack prints stack information.
func PrintStack(max int) {
	for c := 0; c < max; c++ {
		pc, file, line, ok := runtime.Caller(c)
		if !ok {
			break
		}
		fmt.Println(c, pc, file, line)
	}
}

// GetPackageName reduces a fully qualified function name to the package name
// There really ought to be to be a better way...
func GetPackageName(f string) string {
	for {
		lastPeriod := strings.LastIndex(f, ".")
		lastSlash := strings.LastIndex(f, "/")

		if lastPeriod > lastSlash {
			f = f[:lastPeriod]
		} else {
			break
		}
	}

	return f
}
