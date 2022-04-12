package hlog

import (
	"log"
	"os"

	"github.com/bingoohuang/golog/pkg/spec"
)

func EnvSize(envName string, defaultValue int) int {
	if s := os.Getenv(envName); s != "" {
		var size spec.Size
		if err := size.Parse(s); err != nil {
			log.Printf("parse env %s=%s failed: %+v", envName, s, err)
		} else if size > 0 {
			return int(size)
		}
	}
	return defaultValue
}
