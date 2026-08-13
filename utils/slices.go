package utils

func SliceToSet[T comparable](s []T) map[T]EmptyType {
	m := make(map[T]EmptyType, len(s))
	for _, i := range s {
		m[i] = Empty
	}
	return m
}

func SliceDedup[T comparable](s []T) []T {
	sets := make(map[T]EmptyType, len(s))
	v := make([]T, len(s))
	l := 0
	for _, i := range s {
		if _, dup := sets[i]; !dup {
			v[l] = i
			sets[i] = Empty
			l++
		}
	}
	return v[:l]
}
