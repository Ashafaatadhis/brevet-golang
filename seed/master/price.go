package master

import (
	"brevet-api/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedPrices is a function that seeds prices to the database
func SeedPrices(db *gorm.DB) error {

	var gundarGroup, mahasiswaLuarGroup, umumGroup models.Group

	if err := db.Where("name = ?", "mahasiswa gunadarma").First(&gundarGroup).Error; err != nil {
		return err
	}
	if err := db.Where("name = ?", "mahasiswa non-gunadarma").First(&mahasiswaLuarGroup).Error; err != nil {
		return err
	}
	if err := db.Where("name = ?", "umum").First(&umumGroup).Error; err != nil {
		return err
	}

	prices := []models.Price{
		{ID: uuid.New(), GroupID: gundarGroup.ID, Price: 750000},
		{ID: uuid.New(), GroupID: mahasiswaLuarGroup.ID, Price: 1000000},
		{ID: uuid.New(), GroupID: umumGroup.ID, Price: 2300000},
	}

	for _, price := range prices {
		if err := db.Where("group_id = ?", price.GroupID).FirstOrCreate(&price).Error; err != nil {
			return err
		}
	}
	return nil
}
