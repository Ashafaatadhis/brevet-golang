package master

import (
	"brevet-api/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedGroups is a function that seeds the groups data
func SeedGroups(db *gorm.DB) error {
	groups := []models.Group{
		{ID: uuid.New(), Name: "mahasiswa gunadarma"},
		{ID: uuid.New(), Name: "mahasiswa non-gunadarma"},
		{ID: uuid.New(), Name: "umum"},
	}

	for _, group := range groups {
		if err := db.Where("name = ?", group.Name).FirstOrCreate(&group).Error; err != nil {
			return err
		}
	}
	return nil
}
