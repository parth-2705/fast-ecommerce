package utils

func Includes[T comparable](array []T, member T) bool {
	for _, item := range array {
		if item == member {
			return true
		}
	}
	return false
}
