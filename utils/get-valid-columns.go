package utils

import (
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

// GetValidColumns returns a map of valid columns for each model, grouped as model.column (e.g., role.name)
func GetValidColumns(db *gorm.DB, models ...any) (map[string]bool, error) {
	validColumns := make(map[string]bool)

	for i, model := range models {
		columns, err := db.Migrator().ColumnTypes(model)
		if err != nil {
			return nil, err
		}

		t := reflect.TypeOf(model)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		modelName := t.Name()
		modelNameLower := strings.ToLower(modelName)

		for _, column := range columns {
			if i == 0 {
				// First model is the main table (User), no prefix
				validColumns[column.Name()] = true
			} else {
				// Relations: use prefix (e.g., profile.name)
				key := fmt.Sprintf("%s.%s", modelNameLower, column.Name())
				validColumns[key] = true
			}
		}
	}

	return validColumns, nil
}
