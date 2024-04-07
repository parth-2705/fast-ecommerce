package helpers

import "time"

func MakeTimeStringfromTimestamp(time time.Time) string {
	return time.Format("January 02, 2006 at 15:04 MST")
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a < b {
		return b
	}
	return a
}
