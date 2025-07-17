package utils

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"gorm.io/gorm"
)

// GetValidColumns returns a map of valid columns for each model, grouped as model.column.
// Deprecated: Use GetValidColumnsFromStruct for better performance without DB introspection.
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

// GetValidColumnsFromStruct returns a map of valid column names from struct fields.
// It uses the `gorm:"column:..."` tag if present, otherwise converts field name to snake_case.
func GetValidColumnsFromStruct(models ...any) map[string]bool {
	validColumns := make(map[string]bool)

	for i, model := range models {
		t := reflect.TypeOf(model)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		modelName := strings.ToLower(t.Name())

		for j := range t.NumField() {

			field := t.Field(j)

			colName := toSnakeCase(field.Name)
			if colName == "" {
				continue
			}

			if i == 0 {
				validColumns[colName] = true
			} else {
				validColumns[modelName+"."+colName] = true
			}
		}
	}

	return validColumns
}

// getColumnName extracts column name from gorm tag, or converts field name to snake_case
func getColumnName(field reflect.StructField) string {
	gormTag := field.Tag.Get("gorm")
	for _, tag := range strings.Split(gormTag, ";") {
		if strings.HasPrefix(tag, "column:") {
			return strings.TrimPrefix(tag, "column:")
		}
	}
	return toSnakeCase(field.Name)
}

// toSnakeCase converts CamelCase to snake_case
func toSnakeCase(str string) string {
	var result []rune
	for i, r := range str {
		if i > 0 && unicode.IsUpper(r) && (i+1 < len(str) && unicode.IsLower(rune(str[i+1]))) {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}
