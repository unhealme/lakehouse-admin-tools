package utils

func SliceToSet[T comparable](s []T) map[T]EmptyType {
	m := make(map[T]EmptyType)
	for _, i := range s {
		m[i] = Empty
	}
	return m
}
