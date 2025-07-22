package utils

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GenerateUniqueCode generates a unique 3-digit code
func GenerateUniqueCode() int {
	return rand.Intn(900) + 100 // hasil antara 100 dan 999
}
