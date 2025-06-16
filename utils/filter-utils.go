package utils

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// ApplyFiltersWithJoins applies filters to a GORM query with support for joins
func ApplyFiltersWithJoins(
	db *gorm.DB,
	baseTable string,
	filters map[string]string,
	validSortFields map[string]bool,
	joinConditions map[string]string,
	joinedRelations map[string]bool,
) *gorm.DB {
	for key, val := range filters {
		if strings.Contains(key, ".") {
			if !validSortFields[key] {
				continue
			}
			parts := strings.SplitN(key, ".", 2)
			relation, column := parts[0], parts[1]
			alias := relation + "s"
			if !joinedRelations[relation] {
				if cond, ok := joinConditions[relation]; ok {
					db = db.Joins(cond)
				} else {
					db = db.Joins(fmt.Sprintf("LEFT JOIN %ss AS %s ON %s.id = %s.%s_id", relation, alias, alias, baseTable, relation))
				}
				joinedRelations[relation] = true
			}
			db = db.Where(fmt.Sprintf("%s.%s = ?", alias, column), val)
		} else {
			if validSortFields[key] {
				db = db.Where(fmt.Sprintf("%s = ?", key), val)
			}
		}
	}
	return db
}
