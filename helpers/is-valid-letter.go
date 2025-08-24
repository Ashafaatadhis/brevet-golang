package helpers

// IsValidLetter for checking letter option answer
func IsValidLetter(letter string, optCount int) bool {
	if len(letter) != 1 {
		return false
	}
	r := rune(letter[0])
	return r >= 'A' && r < 'A'+rune(optCount)
}
