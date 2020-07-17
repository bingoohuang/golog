package spec

import "strings"

type Layout string

func (l *Layout) Parse(s string) error {
	s = strings.Replace(s, "yyyy", "2006", -1)
	s = strings.Replace(s, "yy", "06", -1)
	s = strings.Replace(s, "MM", "01", -1)
	s = strings.Replace(s, "dd", "02", -1)
	s = strings.Replace(s, "hh", "03", -1)
	s = strings.Replace(s, "HH", "15", -1)
	s = strings.Replace(s, "mm", "04", -1)
	s = strings.Replace(s, "ss", "05", -1)
	s = strings.Replace(s, "SSS", "000", -1)

	*l = Layout(s)
	return nil
}
