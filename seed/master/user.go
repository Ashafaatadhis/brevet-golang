package master

import (
	"brevet-api/models"
	"brevet-api/utils"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userSeed struct {
	Name          string
	Email         string
	Phone         string
	Password      string
	RoleType      models.RoleType
	GroupType     *models.GroupType
	GroupVerified bool
	Institution   string
	Origin        string
	Address       string
}

func groupTypePtr(v models.GroupType) *models.GroupType {
	return &v
}

// SeedUsers seeds default users and profiles in an idempotent way (by email).
func SeedUsers(db *gorm.DB) error {
	seeds := []userSeed{
		{
			Name:          "Admin Brevet",
			Email:         "admin@brevet.local",
			Phone:         "081111111111",
			Password:      "Admin123!",
			RoleType:      models.RoleTypeAdmin,
			GroupType:     nil,
			GroupVerified: true,
			Institution:   "Tax Center Gunadarma",
			Origin:        "Jakarta",
			Address:       "Jakarta",
		},
		{
			Name:          "Guru Brevet",
			Email:         "guru@brevet.local",
			Phone:         "082222222222",
			Password:      "Guru123!",
			RoleType:      models.RoleTypeGuru,
			GroupType:     nil,
			GroupVerified: true,
			Institution:   "Tax Center Gunadarma",
			Origin:        "Depok",
			Address:       "Depok",
		},
		{
			Name:          "Siswa Brevet",
			Email:         "siswa@brevet.local",
			Phone:         "083333333333",
			Password:      "Siswa123!",
			RoleType:      models.RoleTypeSiswa,
			GroupType:     groupTypePtr(models.Umum),
			GroupVerified: true,
			Institution:   "Universitas Gunadarma",
			Origin:        "Bekasi",
			Address:       "Bekasi",
		},
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, s := range seeds {
			var user models.User
			err := tx.Where("email = ?", s.Email).First(&user).Error
			switch {
			case errors.Is(err, gorm.ErrRecordNotFound):
				hashedPassword, hashErr := utils.HashPassword(s.Password)
				if hashErr != nil {
					return fmt.Errorf("failed hashing password for %s: %w", s.Email, hashErr)
				}

				user = models.User{
					ID:         uuid.New(),
					Name:       s.Name,
					Phone:      s.Phone,
					Email:      s.Email,
					Password:   hashedPassword,
					RoleType:   s.RoleType,
					IsVerified: true,
				}
				if createErr := tx.Create(&user).Error; createErr != nil {
					return fmt.Errorf("failed creating user %s: %w", s.Email, createErr)
				}
			case err != nil:
				return fmt.Errorf("failed finding user %s: %w", s.Email, err)
			default:
				if updateErr := tx.Model(&user).Updates(map[string]any{
					"name":        s.Name,
					"phone":       s.Phone,
					"role_type":   s.RoleType,
					"is_verified": true,
				}).Error; updateErr != nil {
					return fmt.Errorf("failed updating user %s: %w", s.Email, updateErr)
				}
			}

			var profile models.Profile
			profileErr := tx.Where("user_id = ?", user.ID).First(&profile).Error
			switch {
			case errors.Is(profileErr, gorm.ErrRecordNotFound):
				profile = models.Profile{
					ID:            uuid.New(),
					UserID:        user.ID,
					GroupType:     s.GroupType,
					GroupVerified: s.GroupVerified,
					Institution:   s.Institution,
					Origin:        s.Origin,
					Address:       s.Address,
				}

				if createProfileErr := tx.Create(&profile).Error; createProfileErr != nil {
					return fmt.Errorf("failed creating profile for %s: %w", s.Email, createProfileErr)
				}
			case profileErr != nil:
				return fmt.Errorf("failed finding profile for %s: %w", s.Email, profileErr)
			default:
				if updateProfileErr := tx.Model(&profile).Updates(map[string]any{
					"group_type":     s.GroupType,
					"group_verified": s.GroupVerified,
					"institution":    s.Institution,
					"origin":         s.Origin,
					"address":        s.Address,
				}).Error; updateProfileErr != nil {
					return fmt.Errorf("failed updating profile for %s: %w", s.Email, updateProfileErr)
				}
			}
		}

		return nil
	})
}
