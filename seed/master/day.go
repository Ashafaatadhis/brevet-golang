package master

import (
	"brevet-api/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedDays seeds the days table with weekday names
func SeedDays(db *gorm.DB) error {
	days := []models.Day{
		{ID: uuid.New(), Name: "Senin"},
		{ID: uuid.New(), Name: "Selasa"},
		{ID: uuid.New(), Name: "Rabu"},
		{ID: uuid.New(), Name: "Kamis"},
		{ID: uuid.New(), Name: "Jumat"},
		{ID: uuid.New(), Name: "Sabtu"},
		{ID: uuid.New(), Name: "Minggu"},
	}

	for _, day := range days {
		if err := db.Where(models.Day{Name: day.Name}).FirstOrCreate(&day).Error; err != nil {
			return err
		}
	}

	return nil
}
