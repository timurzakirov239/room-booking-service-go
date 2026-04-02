package timeutil

import "time"

func NowUTC() time.Time {
	return time.Now().UTC()
}

func NormalizeUTC(value time.Time) time.Time {
	return value.UTC()
}
