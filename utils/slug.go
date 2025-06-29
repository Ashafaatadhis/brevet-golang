package utils

import (
	"fmt"
	"strings"

	"github.com/gosimple/slug"
)

// SlugChecker is an interface that defines a method to check if a slug already exists
type SlugChecker interface {
	IsSlugExists(slug string) bool
}

// GenerateUniqueSlug membuat slug unik berdasarkan title
func GenerateUniqueSlug(title string, checker SlugChecker) string {
	baseSlug := slug.Make(strings.ToLower(title))
	slug := baseSlug
	i := 1
	for checker.IsSlugExists(slug) {
		slug = fmt.Sprintf("%s-%d", baseSlug, i)
		i++
	}
	return slug
}
