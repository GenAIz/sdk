package stringz

import (
	"strings"
)

// AllNonEmpty filters empty strings out of the provided strings, returning a new array of strings
func AllNonEmpty(ss ...string) []string {
	var result []string

	for _, s := range ss {
		if s != "" {
			result = append(result, s)
		}
	}

	return result
}

// CharAt returns a string of one character if the index is contained within the string
func CharAt(s string, i int) string {
	if i >= 0 && len(s) > i {
		return s[i : i+1]
	}

	return ""
}

// FirstNonEmpty returns the first non-empty provided as argument: YAY!
func FirstNonEmpty(s1 string, s2 ...string) string {
	if s1 == "" && len(s2) > 0 {
		return FirstNonEmpty(s2[0], s2[1:]...)
	}

	return s1
}

// MultiTagLabel appends a tag to a label with the provided delimiter if both the tag and delimiter are not empty
func MultiTagLabel(label string, delimiter string, tag string) string {
	if tag != "" && delimiter != "" {
		return label + delimiter + tag
	}

	return label
}

// SingleTagLabel appends a tag to its label with the provided delimiter if both the tag and delimiter are not empty and if the delimiter is not already present in the label
func SingleTagLabel(label string, delimiter string, tag string) string {
	if tag != "" && delimiter != "" && !strings.Contains(label, delimiter) {
		return label + delimiter + tag
	}

	return label
}

// SubstrFrom extracts the part of the string after the specified, including the indexed character or empty string if the index overflows the string length.
func SubstrFrom(label string, index int) string {
	if index < len(label) {
		return label[index:]
	}

	return ""
}
