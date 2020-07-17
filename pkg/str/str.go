package str

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
