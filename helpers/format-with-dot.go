package helpers

import "fmt"

// FormatWithDot is function
func FormatWithDot(n int) string {
	s := fmt.Sprintf("%d", n)
	var result []byte
	count := 0
	for i := len(s) - 1; i >= 0; i-- {
		if count == 3 {
			result = append([]byte{'.'}, result...)
			count = 0
		}
		result = append([]byte{s[i]}, result...)
		count++
	}
	return string(result)
}
