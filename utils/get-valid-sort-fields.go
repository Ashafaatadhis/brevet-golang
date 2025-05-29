package utils

import "gorm.io/gorm"

// GetValidSortFields mengambil nama kolom yang bisa digunakan untuk sorting
func GetValidSortFields(db *gorm.DB, models ...interface{}) (map[string]bool, error) {
	validSortFields := make(map[string]bool)

	for _, model := range models {
		columns, err := db.Migrator().ColumnTypes(model)
		if err != nil {
			return nil, err
		}

		for _, column := range columns {
			validSortFields[column.Name()] = true
		}
	}

	return validSortFields, nil
}
