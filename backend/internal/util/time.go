package util

import "time"

func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func PtrTime(t time.Time) *time.Time {
	return &t
}
