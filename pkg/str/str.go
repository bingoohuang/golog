package str

import (
	"strconv"
	"strings"
)

func AnyOf(v string, any ...string) bool {
	for _, s := range any {
		if v == s {
			return true
		}
	}

	return false
}

func Or(a, b string) string {
	if a == "" {
		return b
	}

	return a
}

// HasSuffixes tests that string s has any of suffixes.
func HasSuffixes(s string, suffixes ...string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}

	return false
}

func ParseInt(s string, defaultValue int) int {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return defaultValue
	}

	return int(i)
}

func ParseBool(s string, defaultValue bool) bool {
	switch strings.ToLower(s) {
	case "true", "t", "yes", "y", "on", "1":
		return true
	case "false", "f", "no", "n", "off", "0":
		return false
	}

	return defaultValue
}
