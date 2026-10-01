package slicez

// Filter returns a slice of the elements for which filter evaluated to true, typical because slices.DeleteFunc is confusing and an extremely bug prone construct, also because it just iterates twice over the collection evaluating filter
// twice for no good reasons.
func Filter[S ~[]E, E any](s S, filter func(E) bool) S {
	var result []E

	for _, element := range s {
		if filter(element) {
			result = append(result, element)
		}
	}

	return result
}

// Transform returns a slice of the elements transformed by the provided function
func Transform[S ~[]E, E any, P any](s S, transform func(E) P) []P {
	var result []P

	for _, element := range s {
		result = append(result, transform(element))
	}

	return result
}
