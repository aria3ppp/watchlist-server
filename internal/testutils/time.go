package testutils

import "time"

// Deprecated: stop using this function and use `t = t.In(loc)` explicitly
func SetTimeLocation(t *time.Time, loc *time.Location) {
	*t = t.In(loc)
}

func Date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
