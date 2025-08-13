package services

import (
	"brevet-api/dto"
	"brevet-api/models"
	"brevet-api/repository"
	"brevet-api/utils"
	"errors"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// UserService is a struct that represents a user service
type UserService struct {
	userRepo repository.IUserTXRepository
	db       *gorm.DB
	authRepo *repository.AuthRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.IUserTXRepository, db *gorm.DB, authRepo *repository.AuthRepository) *UserService {
	return &UserService{userRepo: userRepo, db: db, authRepo: authRepo}
}

// GetAllFilteredUsers is a method that returns all users
func (s *UserService) GetAllFilteredUsers(opts utils.QueryOptions) ([]models.User, int64, error) {
	return s.userRepo.FindAllFiltered(opts)
}

// GetUserByID retrieves a user by their ID
func (s *UserService) GetUserByID(id uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetProfileResponseByID retrieves a user's profile response by their ID
func (s *UserService) GetProfileResponseByID(userID uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	var userResp dto.UserResponse
	if err := copier.Copy(&userResp, user); err != nil {
		return nil, err
	}
	if err := copier.Copy(&userResp.Profile, user.Profile); err != nil {
		return nil, err
	}

	return &userResp, nil
}

// CreateUserWithProfile creates a new user with an associated profile
func (s *UserService) CreateUserWithProfile(body *dto.CreateUserWithProfileRequest) (*dto.UserResponse, error) {
	var userResp dto.UserResponse

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		// Cek duplikasi
		if !s.authRepo.WithTx(tx).IsEmailUnique(body.Email) || !s.authRepo.WithTx(tx).IsPhoneUnique(body.Phone) {
			return errors.New("email atau nomor telepon sudah digunakan")
		}

		// Hash password
		hashedPassword, err := utils.HashPassword(body.Password)
		if err != nil {
			return err
		}

		// Mapping ke model
		user := &models.User{
			RoleType:   models.RoleTypeSiswa,
			IsVerified: true,
			Password:   hashedPassword,
		}
		if err := copier.CopyWithOption(user, body, copier.Option{IgnoreEmpty: true}); err != nil {
			return err
		}

		profile := &models.Profile{}
		if err := copier.CopyWithOption(profile, body, copier.Option{IgnoreEmpty: true}); err != nil {
			return err
		}

		// Simpan user
		if err := s.userRepo.WithTx(tx).Create(user); err != nil {
			return err
		}

		profile.UserID = user.ID

		// Create profile setelah user
		if err := s.userRepo.WithTx(tx).CreateProfile(profile); err != nil {
			return err
		}

		// Fetch ulang user lengkap
		fullUser, err := s.userRepo.WithTx(tx).FindByID(user.ID)
		if err != nil {
			return err
		}

		// Mapping ke response
		if err := copier.Copy(&userResp, fullUser); err != nil {
			return err
		}
		if fullUser.Profile != nil {
			if err := copier.Copy(&userResp.Profile, fullUser.Profile); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &userResp, nil
}

// UpdateUserWithProfile updates an existing user and their profile
func (s *UserService) UpdateUserWithProfile(userID uuid.UUID, body *dto.UpdateUserWithProfileRequest) (*dto.UserResponse, error) {

	var userResp dto.UserResponse
	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {

		user, err := s.userRepo.WithTx(tx).FindByID(userID)
		if err != nil {
			return err
		}

		// Copy ke user
		if err := copier.CopyWithOption(user, body, copier.Option{IgnoreEmpty: true}); err != nil {
			return err
		}

		if user.Profile == nil {
			user.Profile = &models.Profile{UserID: user.ID}
		}

		if err := copier.CopyWithOption(user.Profile, body, copier.Option{IgnoreEmpty: true}); err != nil {
			return err
		}

		// Save user
		if err := s.userRepo.WithTx(tx).Save(user); err != nil {
			return err
		}

		if err := s.userRepo.WithTx(tx).SaveProfile(user.Profile); err != nil {
			return err
		}

		// Mapping response
		if err := copier.Copy(&userResp, user); err != nil {
			return err
		}
		if user.Profile != nil {
			if err := copier.Copy(&userResp.Profile, user.Profile); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &userResp, nil

}

// DeleteUserByID deletes a user by their ID
func (s *UserService) DeleteUserByID(userID uuid.UUID) error {
	// Validasi: apakah user ada
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	// Hapus user
	return s.userRepo.DeleteUser(userID)
}

// UpdateMyProfile updates the profile of the authenticated user
func (s *UserService) UpdateMyProfile(userID uuid.UUID, body *dto.UpdateMyProfile) (*dto.UserResponse, error) {
	var userResp dto.UserResponse

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		user, err := s.userRepo.WithTx(tx).FindByID(userID)
		if err != nil {
			return err
		}

		// Inisialisasi profile jika belum ada
		if user.Profile == nil {
			user.Profile = &models.Profile{UserID: user.ID}
		}

		// Copy data
		if err := copier.CopyWithOption(user, body, copier.Option{IgnoreEmpty: true}); err != nil {
			return err
		}
		if err := copier.CopyWithOption(user.Profile, body, copier.Option{IgnoreEmpty: true}); err != nil {
			return err
		}

		// Simpan user dan profile
		if err := s.userRepo.WithTx(tx).Save(user); err != nil {
			return err
		}
		if err := s.userRepo.WithTx(tx).SaveProfile(user.Profile); err != nil {
			return err
		}

		// Copy ke response
		if err := copier.Copy(&userResp, user); err != nil {
			return err
		}
		if user.Profile != nil {
			if err := copier.Copy(&userResp.Profile, user.Profile); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &userResp, nil
}
