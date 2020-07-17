package spec

import (
	"strconv"
	"strings"

	"github.com/bingoohuang/golog/pkg/rotate"
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

	value, _ := strconv.Atoi(s[0:i])
	unit := s[i:]
	switch strings.ToUpper(unit) {
	case "K":
		*size = Size(value * rotate.KiB)
	case "M":
		*size = Size(value * rotate.MiB)
	case "G":
		*size = Size(value * rotate.GiB)
	default:
		*size = Size(value)
		return errors.Errorf("unknown unit %s", s)
	}

	return nil
}
