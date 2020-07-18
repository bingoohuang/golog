package spec

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Size int

func (size *Size) Parse(s string) error {
	i := strings.IndexFunc(s, func(r rune) bool {
		return r < '0' || r > '9'
	})
	if i < 0 {
		v, err := strconv.Atoi(s)
		*size = Size(v)
		return err
	}

	switch v, _ := strconv.Atoi(s[0:i]); strings.ToUpper(s[i:]) {
	case "K", "KIB":
		*size = Size(v * KiB)
	case "M", "MIB":
		*size = Size(v * MiB)
	case "G", "GIB":
		*size = Size(v * GiB)
	case "KB":
		*size = Size(v * KB)
	case "MB":
		*size = Size(v * MB)
	case "GB":
		*size = Size(v * GB)
	default:
		*size = Size(v)
		return errors.Errorf("unknown unit %s", s)
	}

	return nil
}

const (
	KiB = 1024
	MiB = 1024 * KiB
	GiB = 1024 * MiB
	KB  = 1000
	MB  = 1000 * KB
	GB  = 1000 * MB
)
