package services

import (
	"brevet-api/config"
	"brevet-api/dto"
	"brevet-api/mocks"
	"brevet-api/models"
	"brevet-api/services"
	"brevet-api/utils"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthService_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("success - register user", func(t *testing.T) {
		// Mock repository dan service
		repo := new(mocks.IAuthRepository)
		verificationSvc := new(mocks.IVerificationService)
		sessionRepo := new(mocks.IUserSessionRepository)

		service := services.NewAuthService(repo, verificationSvc, sessionRepo, nil, nil)

		req := &dto.RegisterRequest{
			Name:      "Test User",
			Email:     "test@example.com",
			Phone:     "081234567890",
			Password:  "password123",
			GroupType: models.Umum,
		}

		userID := uuid.New()

		profile := models.Profile{GroupVerified: true}
		fullUser := models.User{
			ID:       userID,
			Name:     req.Name,
			Email:    req.Email,
			Phone:    req.Phone,
			RoleType: models.RoleTypeSiswa,
		}

		// Mock: validasi unik
		repo.On("WithTx", mock.Anything).Return(repo)
		repo.On("IsEmailUnique", ctx, req.Email).Return(true)
		repo.On("IsPhoneUnique", ctx, req.Phone).Return(true)

		// Mock: CreateUser
		repo.On("CreateUser", ctx, mock.AnythingOfType("*models.User")).Return(nil).Run(func(args mock.Arguments) {
			u := args.Get(1).(*models.User)
			u.ID = userID
		})

		// Mock: GenerateVerificationCode
		verificationSvc.On("GenerateVerificationCode", ctx, mock.Anything, userID).Return("123456", nil)

		// Mock: CreateProfile
		repo.On("CreateProfile", ctx, mock.AnythingOfType("*models.Profile")).Return(nil)

		// Mock: GetUserByID
		repo.On("GetUserByID", ctx, userID).Return(&fullUser, nil)

		// Panggil service
		resp, err := service.Register(ctx, nil, req)
		assert.NoError(t, err)
		assert.Equal(t, req.Name, resp.Name)
		assert.Equal(t, profile.GroupVerified, resp.Profile.GroupVerified)

		// Assert semua ekspektasi terpenuhi
		repo.AssertExpectations(t)
		verificationSvc.AssertExpectations(t)
	})

	t.Run("fail - email or phone not unique", func(t *testing.T) {
		repo := new(mocks.IAuthRepository)
		verificationSvc := new(mocks.IVerificationService)
		sessionRepo := new(mocks.IUserSessionRepository)

		service := services.NewAuthService(repo, verificationSvc, sessionRepo, nil, nil)

		req := &dto.RegisterRequest{
			Email: "exists@example.com",
			Phone: "081234567890",
		}

		// Mock: validasi unik gagal
		repo.On("WithTx", mock.Anything).Return(repo)
		repo.On("IsEmailUnique", ctx, req.Email).Return(false)
		repo.On("IsPhoneUnique", ctx, req.Phone).Return(true)

		resp, err := service.Register(ctx, nil, req)
		assert.Nil(t, resp)
		assert.EqualError(t, err, "Email atau nomor telephone sudah digunakan")
	})
}

func TestAuthService_Login(t *testing.T) {
	ctx := context.Background()

	t.Run("success - login user", func(t *testing.T) {
		// Mock repos
		repo := new(mocks.IAuthRepository)
		verificationSvc := new(mocks.IVerificationService)
		sessionRepo := new(mocks.IUserSessionRepository)

		service := services.NewAuthService(repo, verificationSvc, sessionRepo, nil, nil)

		userID := uuid.New()
		hashedPassword, _ := utils.HashPassword("password123")

		user := &models.User{
			ID:         userID,
			Email:      "test@example.com",
			Password:   hashedPassword,
			IsVerified: true,
			Profile: &models.Profile{
				UserID:        userID,
				GroupVerified: true,
			},
		}

		// Mock GetUserByEmailWithProfile
		repo.On("GetUserByEmailWithProfile", ctx, user.Email).Return(user, nil)

		// Mock CreateUserSession
		repo.On("CreateUserSession", ctx, user.ID, mock.Anything, mock.Anything).Return(nil)

		// Gunakan fiber ctx kosong
		c := &fiber.Ctx{}

		result, err := service.Login(ctx, &dto.LoginRequest{
			Email:    user.Email,
			Password: "password123",
		}, c)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.AccessToken)
		assert.NotEmpty(t, result.RefreshToken)
		assert.Equal(t, user.Email, result.User.Email)

		repo.AssertExpectations(t)
		sessionRepo.AssertExpectations(t)
	})

	t.Run("fail - email not found", func(t *testing.T) {
		repo := new(mocks.IAuthRepository)
		service := services.NewAuthService(repo, nil, nil, nil, nil)

		repo.On("GetUserByEmailWithProfile", ctx, "notfound@example.com").
			Return(nil, assert.AnError)

		c := &fiber.Ctx{}

		result, err := service.Login(ctx, &dto.LoginRequest{
			Email:    "notfound@example.com",
			Password: "any",
		}, c)

		assert.Nil(t, result)
		assert.EqualError(t, err, "Email tidak ditemukan")
	})

	t.Run("fail - wrong password", func(t *testing.T) {
		repo := new(mocks.IAuthRepository)
		service := services.NewAuthService(repo, nil, nil, nil, nil)

		hashedPassword, _ := utils.HashPassword("correctpass")
		user := &models.User{
			Email:    "test@example.com",
			Password: hashedPassword,
		}

		repo.On("GetUserByEmailWithProfile", ctx, user.Email).Return(user, nil)

		c := &fiber.Ctx{}

		result, err := service.Login(ctx, &dto.LoginRequest{
			Email:    user.Email,
			Password: "wrongpass",
		}, c)

		assert.Nil(t, result)
		assert.EqualError(t, err, "Password salah")
	})

	t.Run("fail - email not verified", func(t *testing.T) {
		repo := new(mocks.IAuthRepository)
		service := services.NewAuthService(repo, nil, nil, nil, nil)

		hashedPassword, _ := utils.HashPassword("password123")
		user := &models.User{
			Email:      "test@example.com",
			Password:   hashedPassword,
			IsVerified: false,
		}

		repo.On("GetUserByEmailWithProfile", ctx, user.Email).Return(user, nil)

		c := &fiber.Ctx{}

		result, err := service.Login(ctx, &dto.LoginRequest{
			Email:    user.Email,
			Password: "password123",
		}, c)

		assert.Nil(t, result)
		assert.EqualError(t, err, "Email belum diverifikasi")
	})
}

func TestAuthService_VerifyUserEmail(t *testing.T) {
	ctx := context.Background()
	code := "123456"

	t.Run("success - email verified", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)

		service := services.NewAuthService(mockRepo, mockVerificationSvc, nil, mockTokenService, nil)

		// Test data
		userID := uuid.New()
		user := &models.User{ID: userID, IsVerified: false}
		token := "valid-token"
		claims := &utils.VerificationClaims{UserID: userID}

		// Setup mock expectations
		mockTokenService.On("ExtractUserIDFromToken", token, mock.AnythingOfType("string")).
			Return(claims, nil)
		mockRepo.On("GetUserByID", ctx, userID).Return(user, nil)
		mockVerificationSvc.On("VerifyCode", ctx, userID, code).Return(true)

		// Execute
		err := service.VerifyUserEmail(ctx, token, code)

		// Assert
		assert.NoError(t, err)

		// Verify all mocks were called as expected
		mockTokenService.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
		mockVerificationSvc.AssertExpectations(t)
	})

	t.Run("fail - invalid token", func(t *testing.T) {
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)

		service := services.NewAuthService(mockRepo, mockVerificationSvc, nil, mockTokenService, nil)

		// Setup mock untuk return error
		mockTokenService.On("ExtractUserIDFromToken", "invalid-token", mock.AnythingOfType("string")).
			Return(nil, errors.New("invalid token"))

		err := service.VerifyUserEmail(ctx, "invalid-token", code)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token tidak valid")
		mockTokenService.AssertExpectations(t)
	})

	t.Run("fail - user already verified", func(t *testing.T) {
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)

		service := services.NewAuthService(mockRepo, mockVerificationSvc, nil, mockTokenService, nil)

		userID := uuid.New()
		user := &models.User{ID: userID, IsVerified: true} // Already verified
		token := "valid-token"
		claims := &utils.VerificationClaims{UserID: userID}

		mockTokenService.On("ExtractUserIDFromToken", token, mock.AnythingOfType("string")).
			Return(claims, nil)
		mockRepo.On("GetUserByID", ctx, userID).Return(user, nil)

		err := service.VerifyUserEmail(ctx, token, code)

		assert.Error(t, err)
		assert.Equal(t, "email sudah diverifikasi", err.Error())
		mockTokenService.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("fail - invalid verification code", func(t *testing.T) {
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)

		service := services.NewAuthService(mockRepo, mockVerificationSvc, nil, mockTokenService, nil)

		userID := uuid.New()
		user := &models.User{ID: userID, IsVerified: false}
		token := "valid-token"
		wrongCode := "wrong-code"
		claims := &utils.VerificationClaims{UserID: userID}

		mockTokenService.On("ExtractUserIDFromToken", token, mock.AnythingOfType("string")).
			Return(claims, nil)
		mockRepo.On("GetUserByID", ctx, userID).Return(user, nil)
		mockVerificationSvc.On("VerifyCode", ctx, userID, wrongCode).Return(false)

		err := service.VerifyUserEmail(ctx, token, wrongCode)

		assert.Error(t, err)
		assert.Equal(t, "kode verifikasi salah atau kadaluarsa", err.Error())
		mockTokenService.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
		mockVerificationSvc.AssertExpectations(t)
	})

	t.Run("fail - user not found", func(t *testing.T) {
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)

		service := services.NewAuthService(mockRepo, mockVerificationSvc, nil, mockTokenService, nil)

		userID := uuid.New()
		token := "valid-token"
		claims := &utils.VerificationClaims{UserID: userID}

		mockTokenService.On("ExtractUserIDFromToken", token, mock.AnythingOfType("string")).
			Return(claims, nil)
		mockRepo.On("GetUserByID", ctx, userID).Return(nil, errors.New("user not found"))

		err := service.VerifyUserEmail(ctx, token, code)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user tidak ditemukan")
		mockTokenService.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}

func TestAuthService_ResendVerificationCode(t *testing.T) {
	ctx := context.Background()

	t.Run("success - resend verification code", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)
		mockEmailService := new(mocks.IEmailService)

		service := services.NewAuthService(mockRepo, mockVerificationSvc, nil, mockTokenService, mockEmailService)

		// Test data
		userID := uuid.New()
		user := &models.User{
			ID:         userID,
			Email:      "test@example.com",
			IsVerified: false,
		}
		inputToken := "valid-token"
		claims := &utils.VerificationClaims{UserID: userID}
		generatedCode := "123456"
		newToken := "new-verification-token"

		// Setup mock expectations
		mockTokenService.On("ExtractUserIDFromToken", inputToken, mock.AnythingOfType("string")).
			Return(claims, nil)
		mockRepo.On("GetUserByID", ctx, userID).Return(user, nil)
		mockVerificationSvc.On("GetCooldownRemaining", ctx, userID).
			Return(time.Duration(0), nil) // No cooldown
		mockVerificationSvc.On("GenerateVerificationCode", ctx, mock.Anything, userID).
			Return(generatedCode, nil)

		mockTokenService.On("GenerateVerificationToken", userID, user.Email).
			Return(newToken, nil)
		mockEmailService.On("SendVerificationEmail", user.Email, generatedCode, newToken).
			Return(nil)

		// Execute
		err := service.ResendVerificationCode(ctx, inputToken)

		// Assert
		assert.NoError(t, err)

		// Verify all mocks were called
		mockTokenService.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
		mockVerificationSvc.AssertExpectations(t)
		mockEmailService.AssertExpectations(t)
	})

	t.Run("fail - invalid token", func(t *testing.T) {
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)
		mockEmailService := new(mocks.IEmailService)

		service := services.NewAuthService(mockRepo, mockVerificationSvc, nil, mockTokenService, mockEmailService)

		mockTokenService.On("ExtractUserIDFromToken", "invalid-token", mock.AnythingOfType("string")).
			Return(nil, errors.New("invalid token"))

		err := service.ResendVerificationCode(ctx, "invalid-token")

		assert.Error(t, err)
		assert.Equal(t, "token tidak valid", err.Error())
		mockTokenService.AssertExpectations(t)
	})

	t.Run("fail - user not found", func(t *testing.T) {
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)
		mockEmailService := new(mocks.IEmailService)

		service := services.NewAuthService(mockRepo, mockVerificationSvc, nil, mockTokenService, mockEmailService)

		userID := uuid.New()
		claims := &utils.VerificationClaims{UserID: userID}

		mockTokenService.On("ExtractUserIDFromToken", "valid-token", mock.AnythingOfType("string")).
			Return(claims, nil)
		mockRepo.On("GetUserByID", ctx, userID).Return(nil, errors.New("user not found"))

		err := service.ResendVerificationCode(ctx, "valid-token")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user tidak ditemukan")
		mockTokenService.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("fail - user already verified", func(t *testing.T) {
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)
		mockEmailService := new(mocks.IEmailService)

		service := services.NewAuthService(mockRepo, mockVerificationSvc, nil, mockTokenService, mockEmailService)

		userID := uuid.New()
		user := &models.User{
			ID:         userID,
			Email:      "test@example.com",
			IsVerified: true, // Already verified
		}
		claims := &utils.VerificationClaims{UserID: userID}

		mockTokenService.On("ExtractUserIDFromToken", "valid-token", mock.AnythingOfType("string")).
			Return(claims, nil)
		mockRepo.On("GetUserByID", ctx, userID).Return(user, nil)

		err := service.ResendVerificationCode(ctx, "valid-token")

		assert.Error(t, err)
		assert.Equal(t, "email sudah diverifikasi", err.Error())
		mockTokenService.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("fail - cooldown active", func(t *testing.T) {
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)
		mockEmailService := new(mocks.IEmailService)

		service := services.NewAuthService(mockRepo, mockVerificationSvc, nil, mockTokenService, mockEmailService)

		userID := uuid.New()
		user := &models.User{
			ID:         userID,
			Email:      "test@example.com",
			IsVerified: false,
		}
		claims := &utils.VerificationClaims{UserID: userID}
		cooldownRemaining := 30 * time.Second

		mockTokenService.On("ExtractUserIDFromToken", "valid-token", mock.AnythingOfType("string")).
			Return(claims, nil)
		mockRepo.On("GetUserByID", ctx, userID).Return(user, nil)
		mockVerificationSvc.On("GetCooldownRemaining", ctx, userID).
			Return(cooldownRemaining, nil)

		err := service.ResendVerificationCode(ctx, "valid-token")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tunggu 30 detik")
		mockTokenService.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
		mockVerificationSvc.AssertExpectations(t)
	})

	t.Run("fail - generate code error", func(t *testing.T) {
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)
		mockEmailService := new(mocks.IEmailService)

		service := services.NewAuthService(mockRepo, mockVerificationSvc, nil, mockTokenService, mockEmailService)

		userID := uuid.New()
		user := &models.User{
			ID:         userID,
			Email:      "test@example.com",
			IsVerified: false,
		}
		claims := &utils.VerificationClaims{UserID: userID}

		mockTokenService.On("ExtractUserIDFromToken", "valid-token", mock.AnythingOfType("string")).
			Return(claims, nil)
		mockRepo.On("GetUserByID", ctx, userID).Return(user, nil)
		mockVerificationSvc.On("GetCooldownRemaining", ctx, userID).
			Return(time.Duration(0), nil)
		mockVerificationSvc.On("GenerateVerificationCode", ctx, mock.Anything, userID).
			Return("", errors.New("failed to generate code"))

		err := service.ResendVerificationCode(ctx, "valid-token")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "gagal generate kode")
		mockTokenService.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
		mockVerificationSvc.AssertExpectations(t)
	})

	t.Run("fail - generate token error", func(t *testing.T) {
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)
		mockEmailService := new(mocks.IEmailService)

		service := services.NewAuthService(mockRepo, mockVerificationSvc, nil, mockTokenService, mockEmailService)

		userID := uuid.New()
		user := &models.User{
			ID:         userID,
			Email:      "test@example.com",
			IsVerified: false,
		}
		claims := &utils.VerificationClaims{UserID: userID}
		generatedCode := "123456"

		mockTokenService.On("ExtractUserIDFromToken", "valid-token", mock.AnythingOfType("string")).
			Return(claims, nil)
		mockRepo.On("GetUserByID", ctx, userID).Return(user, nil)
		mockVerificationSvc.On("GetCooldownRemaining", ctx, userID).
			Return(time.Duration(0), nil)
		mockVerificationSvc.On("GenerateVerificationCode", ctx, mock.Anything, userID).
			Return(generatedCode, nil)
		mockTokenService.On("GenerateVerificationToken", userID, user.Email).
			Return("", errors.New("failed to generate token"))

		err := service.ResendVerificationCode(ctx, "valid-token")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "gagal buat token verifikasi")
		mockTokenService.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
		mockVerificationSvc.AssertExpectations(t)
	})

	t.Run("fail - send email error", func(t *testing.T) {
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)
		mockEmailService := new(mocks.IEmailService)

		service := services.NewAuthService(mockRepo, mockVerificationSvc, nil, mockTokenService, mockEmailService)

		userID := uuid.New()
		user := &models.User{
			ID:         userID,
			Email:      "test@example.com",
			IsVerified: false,
		}
		claims := &utils.VerificationClaims{UserID: userID}
		generatedCode := "123456"
		newToken := "new-verification-token"

		mockTokenService.On("ExtractUserIDFromToken", "valid-token", mock.AnythingOfType("string")).
			Return(claims, nil)
		mockRepo.On("GetUserByID", ctx, userID).Return(user, nil)
		mockVerificationSvc.On("GetCooldownRemaining", ctx, userID).
			Return(time.Duration(0), nil)
		mockVerificationSvc.On("GenerateVerificationCode", ctx, mock.Anything, userID).
			Return(generatedCode, nil)
		mockTokenService.On("GenerateVerificationToken", userID, user.Email).
			Return(newToken, nil)
		mockEmailService.On("SendVerificationEmail", user.Email, generatedCode, newToken).
			Return(errors.New("failed to send email"))

		err := service.ResendVerificationCode(ctx, "valid-token")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "gagal kirim email")
		mockTokenService.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
		mockVerificationSvc.AssertExpectations(t)
		mockEmailService.AssertExpectations(t)
	})
}

func TestAuthService_RefreshTokens(t *testing.T) {
	ctx := context.Background()

	t.Run("success - refresh tokens", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)
		mockEmailService := new(mocks.IEmailService)

		service := services.NewAuthService(mockRepo, mockVerificationSvc, nil, mockTokenService, mockEmailService)

		// Test data
		userID := uuid.New()
		user := &models.User{
			ID:    userID,
			Email: "test@example.com",
			Name:  "testuser",
			Profile: &models.Profile{
				ID:          uuid.New(),
				UserID:      userID,
				Institution: "TEST INSTITUTION",
			},
		}
		refreshToken := "valid-refresh-token"
		claims := &utils.VerificationClaims{UserID: userID}
		newAccessToken := "new-access-token"

		// Setup mock expectations
		mockTokenService.On("ExtractUserIDFromToken", refreshToken, mock.AnythingOfType("string")).
			Return(claims, nil)
		mockRepo.On("GetUserByID", ctx, userID).Return(user, nil)
		mockTokenService.On("GenerateJWT", *user, mock.AnythingOfType("string"), mock.AnythingOfType("int")).
			Return(newAccessToken, nil)

		// Execute
		result, err := service.RefreshTokens(ctx, refreshToken)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newAccessToken, result.AccessToken)
		assert.Equal(t, user.Email, result.User.Email)
		assert.Equal(t, user.Name, result.User.Name)
		assert.Equal(t, user.Profile.Institution, result.User.Profile.Institution)

		// Verify all mocks were called
		mockTokenService.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("fail - invalid refresh token", func(t *testing.T) {
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)
		mockEmailService := new(mocks.IEmailService)

		service := services.NewAuthService(mockRepo, mockVerificationSvc, nil, mockTokenService, mockEmailService)

		mockTokenService.On("ExtractUserIDFromToken", "invalid-token", mock.AnythingOfType("string")).
			Return(nil, errors.New("invalid token"))

		result, err := service.RefreshTokens(ctx, "invalid-token")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "refresh token tidak valid")
		mockTokenService.AssertExpectations(t)
	})

	t.Run("fail - user not found", func(t *testing.T) {
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)
		mockEmailService := new(mocks.IEmailService)

		service := services.NewAuthService(mockRepo, mockVerificationSvc, nil, mockTokenService, mockEmailService)

		userID := uuid.New()
		claims := &utils.VerificationClaims{UserID: userID}

		mockTokenService.On("ExtractUserIDFromToken", "valid-token", mock.AnythingOfType("string")).
			Return(claims, nil)
		mockRepo.On("GetUserByID", ctx, userID).Return(nil, errors.New("user not found"))

		result, err := service.RefreshTokens(ctx, "valid-token")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "user tidak ditemukan")
		mockTokenService.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("fail - generate JWT error", func(t *testing.T) {
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)
		mockEmailService := new(mocks.IEmailService)

		service := services.NewAuthService(mockRepo, mockVerificationSvc, nil, mockTokenService, mockEmailService)

		userID := uuid.New()
		user := &models.User{
			ID:    userID,
			Email: "test@example.com",
			Name:  "testuser",
			Profile: &models.Profile{
				ID:          uuid.New(),
				UserID:      userID,
				Institution: "TEST INSTITUTION",
			},
		}
		claims := &utils.VerificationClaims{UserID: userID}

		mockTokenService.On("ExtractUserIDFromToken", "valid-token", mock.AnythingOfType("string")).
			Return(claims, nil)
		mockRepo.On("GetUserByID", ctx, userID).Return(user, nil)
		mockTokenService.On("GenerateJWT", *user, mock.AnythingOfType("string"), mock.AnythingOfType("int")).
			Return("", errors.New("failed to generate JWT"))

		result, err := service.RefreshTokens(ctx, "valid-token")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "gagal generate token baru")
		mockTokenService.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("fail - expired refresh token", func(t *testing.T) {
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)
		mockEmailService := new(mocks.IEmailService)

		service := services.NewAuthService(mockRepo, mockVerificationSvc, nil, mockTokenService, mockEmailService)

		mockTokenService.On("ExtractUserIDFromToken", "expired-token", mock.AnythingOfType("string")).
			Return(nil, errors.New("token expired"))

		result, err := service.RefreshTokens(ctx, "expired-token")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "refresh token tidak valid")
		mockTokenService.AssertExpectations(t)
	})

	t.Run("fail - malformed refresh token", func(t *testing.T) {
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)
		mockEmailService := new(mocks.IEmailService)

		service := services.NewAuthService(mockRepo, mockVerificationSvc, nil, mockTokenService, mockEmailService)

		mockTokenService.On("ExtractUserIDFromToken", "malformed.token", mock.AnythingOfType("string")).
			Return(nil, errors.New("malformed token"))

		result, err := service.RefreshTokens(ctx, "malformed.token")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "refresh token tidak valid")
		mockTokenService.AssertExpectations(t)
	})

	t.Run("fail - empty refresh token", func(t *testing.T) {
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)
		mockEmailService := new(mocks.IEmailService)

		service := services.NewAuthService(mockRepo, mockVerificationSvc, nil, mockTokenService, mockEmailService)

		mockTokenService.On("ExtractUserIDFromToken", "", mock.AnythingOfType("string")).
			Return(nil, errors.New("empty token"))

		result, err := service.RefreshTokens(ctx, "")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "refresh token tidak valid")
		mockTokenService.AssertExpectations(t)
	})

	t.Run("fail - database connection error", func(t *testing.T) {
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)
		mockEmailService := new(mocks.IEmailService)

		service := services.NewAuthService(mockRepo, mockVerificationSvc, nil, mockTokenService, mockEmailService)

		userID := uuid.New()
		claims := &utils.VerificationClaims{UserID: userID}

		mockTokenService.On("ExtractUserIDFromToken", "valid-token", mock.AnythingOfType("string")).
			Return(claims, nil)
		mockRepo.On("GetUserByID", ctx, userID).Return(nil, errors.New("database connection failed"))

		result, err := service.RefreshTokens(ctx, "valid-token")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "user tidak ditemukan")
		mockTokenService.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("fail - user with nil profile", func(t *testing.T) {
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)
		mockEmailService := new(mocks.IEmailService)

		service := services.NewAuthService(mockRepo, mockVerificationSvc, nil, mockTokenService, mockEmailService)

		userID := uuid.New()
		user := &models.User{
			ID:    userID,
			Email: "test@example.com",
			Name:  "testuser",
			// Profile is nil/empty - should still work
		}
		claims := &utils.VerificationClaims{UserID: userID}
		newAccessToken := "new-access-token"

		mockTokenService.On("ExtractUserIDFromToken", "valid-token", mock.AnythingOfType("string")).
			Return(claims, nil)
		mockRepo.On("GetUserByID", ctx, userID).Return(user, nil)
		mockTokenService.On("GenerateJWT", *user, mock.AnythingOfType("string"), mock.AnythingOfType("int")).
			Return(newAccessToken, nil)

		result, err := service.RefreshTokens(ctx, "valid-token")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newAccessToken, result.AccessToken)
		assert.Equal(t, user.Email, result.User.Email)
		assert.Equal(t, user.Name, result.User.Name)
		assert.Empty(t, result.User.Profile)

		mockTokenService.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}

func TestAuthService_LogoutUser(t *testing.T) {
	ctx := context.Background()

	t.Run("success - logout user", func(t *testing.T) {
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockSessionRepo := new(mocks.IUserSessionRepository)
		mockTokenService := new(mocks.ITokenService)
		mockEmailService := new(mocks.IEmailService)

		db, mockRedis := redismock.NewClientMock()
		config.RedisClient = db

		service := services.NewAuthService(mockRepo, mockVerificationSvc, mockSessionRepo, mockTokenService, mockEmailService)

		accessToken := "valid-access-token"
		refreshToken := "valid-refresh-token"

		// Mock session repo
		mockSessionRepo.On("GetByRefreshToken", ctx, refreshToken).
			Return(&models.UserSession{RefreshToken: refreshToken}, nil)

		mockSessionRepo.On("Update", ctx, mock.AnythingOfType("*models.UserSession")).Return(nil)

		// Redis expectation
		mockRedis.ExpectSet(accessToken, "blacklisted", 24*time.Hour).SetVal("OK")

		// Execute
		err := service.LogoutUser(ctx, accessToken, refreshToken)

		// Assert
		assert.NoError(t, err)
		assert.NoError(t, mockRedis.ExpectationsWereMet())
		mockSessionRepo.AssertExpectations(t)
	})
	t.Run("fail - revoke session error", func(t *testing.T) {
		mockSessionRepo := new(mocks.IUserSessionRepository)
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)
		mockEmailService := new(mocks.IEmailService)

		db, _ := redismock.NewClientMock()
		config.RedisClient = db

		service := services.NewAuthService(
			mockRepo,
			mockVerificationSvc,
			mockSessionRepo, // ✅ inject mock sessionRepo
			mockTokenService,
			mockEmailService,
		)

		accessToken := "valid-access-token"
		refreshToken := "invalid-refresh-token"

		// Bikin GetByRefreshToken return error biar revoke gagal
		mockSessionRepo.
			On("GetByRefreshToken", ctx, refreshToken).
			Return(nil, errors.New("db error"))

		// Execute
		err := service.LogoutUser(ctx, accessToken, refreshToken)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "gagal revoke session")
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("fail - redis set error", func(t *testing.T) {
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)
		mockEmailService := new(mocks.IEmailService)
		mockSessionRepo := new(mocks.IUserSessionRepository) // ✅ ini yang dipakai LogoutUser

		db, mockRedis := redismock.NewClientMock()
		config.RedisClient = db

		service := services.NewAuthService(
			mockRepo,
			mockVerificationSvc,
			mockSessionRepo, // ✅ inject sessionRepo
			mockTokenService,
			mockEmailService,
		)

		accessToken := "valid-access-token"
		refreshToken := "valid-refresh-token"

		// Mock revoke session sukses → harus pakai sessionRepo
		mockSessionRepo.
			On("GetByRefreshToken", ctx, refreshToken).
			Return(&models.UserSession{RefreshToken: refreshToken}, nil)

		mockSessionRepo.
			On("Update", ctx, mock.AnythingOfType("*models.UserSession")).
			Return(nil)

		// Redis gagal → SetErr dipanggil di expectation
		mockRedis.ExpectSet(accessToken, "blacklisted", 24*time.Hour).
			SetErr(errors.New("redis error"))

		// Execute
		err := service.LogoutUser(ctx, accessToken, refreshToken)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "gagal blacklist token")

		// Pastikan expectation redis terpenuhi
		assert.NoError(t, mockRedis.ExpectationsWereMet())
		mockSessionRepo.AssertExpectations(t)
	})

}

func TestAuthService_RevokeUserSessionByRefreshToken(t *testing.T) {
	ctx := context.Background()

	t.Run("success - revoke session", func(t *testing.T) {
		mockSessionRepo := new(mocks.IUserSessionRepository)
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)
		mockEmailService := new(mocks.IEmailService)

		service := services.NewAuthService(
			mockRepo,
			mockVerificationSvc,
			mockSessionRepo,
			mockTokenService,
			mockEmailService,
		)

		refreshToken := "valid-refresh-token"
		session := &models.UserSession{RefreshToken: refreshToken, IsRevoked: false}

		// Mocking
		mockSessionRepo.
			On("GetByRefreshToken", ctx, refreshToken).
			Return(session, nil)
		mockSessionRepo.
			On("Update", ctx, mock.AnythingOfType("*models.UserSession")).
			Return(nil)

		// Execute
		err := service.RevokeUserSessionByRefreshToken(ctx, refreshToken)

		// Assert
		assert.NoError(t, err)
		assert.True(t, session.IsRevoked)
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("fail - session not found", func(t *testing.T) {
		mockSessionRepo := new(mocks.IUserSessionRepository)
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)
		mockEmailService := new(mocks.IEmailService)

		service := services.NewAuthService(
			mockRepo,
			mockVerificationSvc,
			mockSessionRepo,
			mockTokenService,
			mockEmailService,
		)

		refreshToken := "invalid-refresh-token"

		// Mocking
		mockSessionRepo.
			On("GetByRefreshToken", ctx, refreshToken).
			Return(nil, errors.New("not found"))

		// Execute
		err := service.RevokeUserSessionByRefreshToken(ctx, refreshToken)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "refresh token session not found")
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("fail - update error", func(t *testing.T) {
		mockSessionRepo := new(mocks.IUserSessionRepository)
		mockRepo := new(mocks.IAuthRepository)
		mockVerificationSvc := new(mocks.IVerificationService)
		mockTokenService := new(mocks.ITokenService)
		mockEmailService := new(mocks.IEmailService)

		service := services.NewAuthService(
			mockRepo,
			mockVerificationSvc,
			mockSessionRepo,
			mockTokenService,
			mockEmailService,
		)

		refreshToken := "valid-refresh-token"
		session := &models.UserSession{RefreshToken: refreshToken, IsRevoked: false}

		// Mocking
		mockSessionRepo.
			On("GetByRefreshToken", ctx, refreshToken).
			Return(session, nil)
		mockSessionRepo.
			On("Update", ctx, mock.AnythingOfType("*models.UserSession")).
			Return(errors.New("db error"))

		// Execute
		err := service.RevokeUserSessionByRefreshToken(ctx, refreshToken)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "gagal revoke session")
		mockSessionRepo.AssertExpectations(t)
	})
}
