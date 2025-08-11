package helpers

import "strings"

// FormatGroupType is function
func FormatGroupType(gt string) string {
	// ganti underscore (_) dengan spasi
	s := strings.ReplaceAll(gt, "_", " ")
	// kapitalisasi setiap kata
	return strings.Title(s)
}
