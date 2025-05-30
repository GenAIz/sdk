package mapz

import "sort"

// Mapped returns a map of the slice hashed by the provided keySupplier function
func Mapped[T any](slice []T, keySupplier func(T) string) map[string]T {
	var result = map[string]T{}

	for _, t := range slice {
		result[keySupplier(t)] = t
	}

	return result
}

// Sorted invokes the consumer with all the keys of the provided map in order dictated by sort.Strings
func Sorted(sorting map[string]string, consumer func(string)) {
	var keys = make([]string, 0, len(sorting))

	for key := range sorting {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		consumer(key)
	}
}
