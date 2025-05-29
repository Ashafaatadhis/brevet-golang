package utils

import (
	"github.com/gofiber/fiber/v2"
)

// QueryOptions represents the query options for pagination and sorting.
type QueryOptions struct {
	Search       string
	Sort         string
	Order        string
	SelectFields string
	Limit        int
	Page         int
	Offset       int
	Filters      map[string]string
}

// ParseQueryOptions parses the query parameters and returns a QueryOptions struct.
func ParseQueryOptions(c *fiber.Ctx) QueryOptions {
	search := c.Query("q", "")
	sort := c.Query("sort", "id")
	order := c.Query("order", "asc")
	selectFields := c.Query("select", "")
	limit := c.QueryInt("limit", 10)
	page := c.QueryInt("page", 1)
	offset := (page - 1) * limit

	// Ambil semua query parameter sebagai filters (kecuali yg sudah dipakai)
	filters := make(map[string]string)
	for key, values := range c.Queries() {
		if key != "q" && key != "sort" && key != "order" && key != "select" && key != "limit" && key != "page" {
			filters[key] = values
		}
	}

	return QueryOptions{
		Search:       search,
		Sort:         sort,
		Order:        order,
		SelectFields: selectFields,
		Limit:        limit,
		Page:         page,
		Offset:       offset,
		Filters:      filters,
	}
}
