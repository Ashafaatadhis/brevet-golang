package master

import (
	"brevet-api/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedRoles is a function that seeds the roles table
func SeedRoles(db *gorm.DB) error {
	roles := []models.Role{
		{ID: uuid.New(), Name: "siswa"},
		{ID: uuid.New(), Name: "admin"},
		{ID: uuid.New(), Name: "guru"},
	}

	for _, role := range roles {
		// FirstOrCreate biar gak duplikat kalau sudah ada
		if err := db.Where(models.Role{Name: role.Name}).FirstOrCreate(&role).Error; err != nil {
			return err
		}
	}
	return nil
}
