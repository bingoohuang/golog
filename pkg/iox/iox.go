package iox

import (
	"fmt"
	"os"
	"time"
)

func InfoReport(a ...interface{}) {
	t := time.Now().Format("2006-01-02 15:04:05.000")
	fmt.Println(append([]interface{}{t}, a...)...)
}

func ErrorReport(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}
