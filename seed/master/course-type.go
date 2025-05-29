package master

import (
	"brevet-api/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedCourseTypes seeds the course_types table with "Online" and "Offline"
func SeedCourseTypes(db *gorm.DB) error {
	types := []models.CourseType{
		{ID: uuid.New(), Name: "Online"},
		{ID: uuid.New(), Name: "Offline"},
	}

	for _, t := range types {
		if err := db.Where(models.CourseType{Name: t.Name}).FirstOrCreate(&t).Error; err != nil {
			return err
		}
	}

	return nil
}
