package utils

import "github.com/gofiber/fiber/v2"

// BuildPaginationMeta is a function to build pagination meta
func BuildPaginationMeta(total int64, limit, page int) fiber.Map {
	return fiber.Map{
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	}
}
