package timex

import "time"

func OrNow(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now()
	}

	return t
}
