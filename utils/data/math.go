package data

import "strconv"

func StringToInteger(s string) int64 {
	i, _ := strconv.Atoi(s)
	// return int64(i)
	return int64(i)
}
