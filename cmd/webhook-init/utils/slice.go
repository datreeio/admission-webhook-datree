package utils

func FindIndex[T interface{}, K []T](findFn func(element T) bool, array K) (idx int) {
	for i, v := range array {
		if findFn(v) {
			return i
		}
	}
	return -1
}
