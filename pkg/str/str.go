package str

import "strings"

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
