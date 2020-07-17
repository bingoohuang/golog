package logfmt_test

import (
	"fmt"
	"testing"

	"github.com/bingoohuang/golog/pkg/logfmt"
)

func TestFormat(t *testing.T) {
	f := logfmt.Formatter{
		PrintColor: true,
	}
	v := f.Format(&logfmt.EntryItem{
		EntryMessage: "hello world",
	})

	fmt.Println(string(v))
}

func TestColor(t *testing.T) {
	fmt.Println("\x1B[30m文字颜色\x1B[0m重置文字颜色")
	fmt.Println("\x1B[31m文字颜色\x1B[0m重置文字颜色")
	fmt.Println("\x1B[32m文字颜色\x1B[0m重置文字颜色")
	fmt.Println("\x1B[33m文字颜色\x1B[0m重置文字颜色")
	fmt.Println("\x1B[34m文字颜色\x1B[0m重置文字颜色")
	fmt.Println("\x1B[35m文字颜色\x1B[0m重置文字颜色")
	fmt.Println("\x1B[36m文字颜色\x1B[0m重置文字颜色")
	fmt.Println("\x1B[37m文字颜色\x1B[0m重置文字颜色")

	fmt.Println("\x1B[40m背景颜色\x1B[0m重置背景颜色")
	fmt.Println("\x1B[41m背景颜色\x1B[0m重置背景颜色")
	fmt.Println("\x1B[42m背景颜色\x1B[0m重置背景颜色")
	fmt.Println("\x1B[43m背景颜色\x1B[0m重置背景颜色")
	fmt.Println("\x1B[44m背景颜色\x1B[0m重置背景颜色")
	fmt.Println("\x1B[45m背景颜色\x1B[0m重置背景颜色")
	fmt.Println("\x1B[46m背景颜色\x1B[0m重置背景颜色")
	fmt.Println("\x1B[47m背景颜色\x1B[0m重置背景颜色")

	fmt.Println("\x1B[30;1m加粗效果\x1B[0m重置加粗效果")
	fmt.Println("\x1B[31;1m加粗效果\x1B[0m重置加粗效果")
	fmt.Println("\x1B[32;1m加粗效果\x1B[0m重置加粗效果")
	fmt.Println("\x1B[33;1m加粗效果\x1B[0m重置加粗效果")
	fmt.Println("\x1B[34;1m加粗效果\x1B[0m重置加粗效果")
	fmt.Println("\x1B[35;1m加粗效果\x1B[0m重置加粗效果")
	fmt.Println("\x1B[36;1m加粗效果\x1B[0m重置加粗效果")
	fmt.Println("\x1B[37;1m加粗效果\x1B[0m重置加粗效果")
}
