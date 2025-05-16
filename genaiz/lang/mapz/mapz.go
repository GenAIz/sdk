package mapz

import "sort"

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
