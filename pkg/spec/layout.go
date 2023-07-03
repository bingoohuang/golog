package spec

import "strings"

type Layout string

func (l *Layout) Parse(s string) error {
	*l = Layout(ConvertTimeLayout(s))
	return nil
}

func ConvertTimeLayout(s string) string {
	s = strings.ReplaceAll(s, "yyyy", "2006")
	s = strings.ReplaceAll(s, "yy", "06")
	s = strings.ReplaceAll(s, "MM", "01")
	s = strings.ReplaceAll(s, "dd", "02")
	s = strings.ReplaceAll(s, "hh", "03")
	s = strings.ReplaceAll(s, "HH", "15")
	s = strings.ReplaceAll(s, "mm", "04")
	s = strings.ReplaceAll(s, "ss", "05")
	s = strings.ReplaceAll(s, "SSS", "000")
	return s
}
