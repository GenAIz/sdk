package stringz

import "strings"

// FirstNonEmpty returns the first non-empty provided as argument: YAY!
func FirstNonEmpty(s1 string, s2 ...string) string {
	if s1 == "" {
		return FirstNonEmpty(s2[0], s2[1:]...)
	}

	return s1
}

func SingleTagMaybe(label string, delimiter string, tag string) string {
	if tag != "" {
		if delimiter != "" && !strings.Contains(label, delimiter) {
			return label + delimiter + tag
		}
	}

	return label
}
