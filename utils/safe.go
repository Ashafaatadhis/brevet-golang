package utils

// Safe is a utility function that safely dereferences a pointer.
func Safe[T any](ptr *T, zero T) T {
	if ptr == nil {
		return zero
	}
	return *ptr
}

// SafeNil is a utility function that safely dereferences a pointer and returns nil if the pointer is nil.
func SafeNil[T any](ptr *T) *T {
	if ptr == nil {
		return nil
	}
	val := *ptr
	return &val
}
