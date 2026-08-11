package mapz

import "sort"

// GetOrDefault returns the corresponding key provided by from or else the supplied defaultValue
func GetOrDefault[k int | string, T any](from map[k]T, key k, defaultSupplier func() T) T {
	if value, ok := from[key]; ok {
		return value
	}

	return defaultSupplier()
}

// Mapped returns a map of the slice hashed by the provided keySupplier function
func Mapped[T any](slice []T, keySupplier func(T) string) map[string]T {
	var result = map[string]T{}

	for _, t := range slice {
		result[keySupplier(t)] = t
	}

	return result
}

// MappedInt64 returns a map of the slice hashed by the provided int64 keySupplier function
func MappedInt64[T any](slice []T, keySupplier func(T) int64) map[int64]T {
	var result = map[int64]T{}

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
