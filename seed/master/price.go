package master

import (
	"brevet-api/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedPrices is a function that seeds prices to the database
func SeedPrices(db *gorm.DB) error {

	prices := []models.Price{
		{ID: uuid.New(), GroupType: models.MahasiswaGunadarma, Price: 750000},
		{ID: uuid.New(), GroupType: models.MahasiswaNonGunadarma, Price: 1000000},
		{ID: uuid.New(), GroupType: models.Umum, Price: 2300000},
	}

	for _, price := range prices {
		if err := db.FirstOrCreate(&price).Error; err != nil {
			return err
		}
	}
	return nil
}
