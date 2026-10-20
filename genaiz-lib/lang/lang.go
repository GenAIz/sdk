package lang

// FirstNonNil returns the first reference non nil in the provided arguments.
//
// Note: This will only work with declared references to non-interface fields. If you declare a non-pointer type and reference it, it automatically becomes non-nil, but empty. The same behavior can be observed with interfaces as they automatically become non-nil on declaration.
func FirstNonNil[T any](values ...*T) *T {
	for _, i := range values {
		if i != nil {
			return i
		}
	}

	return nil
}
