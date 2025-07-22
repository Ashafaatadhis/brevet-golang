package utils

// Safe is a utility function that safely dereferences a pointer.
func Safe[T any](ptr *T, zero T) T {
	if ptr == nil {
		return zero
	}
	return *ptr
}
