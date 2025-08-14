package services

import (
	"brevet-api/config"
	"brevet-api/dto"
	"brevet-api/models"
	"brevet-api/repository"
	"context"
	"fmt"
	"strconv"
	"time"

	"brevet-api/utils"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// IAuthService interface
type IAuthService interface {
	Register(ctx context.Context, tx *gorm.DB, req *dto.RegisterRequest) (*dto.RegisterResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest, c *fiber.Ctx) (*dto.LoginResult, error)
	VerifyUserEmail(ctx context.Context, token, code string) error
	ResendVerificationCode(ctx context.Context, token string) error
	RefreshTokens(ctx context.Context, refreshToken string) (*dto.LoginResult, error)
	LogoutUser(ctx context.Context, accessToken, refreshToken string) error
	RevokeUserSessionByRefreshToken(ctx context.Context, refreshToken string) error
}

// AuthService is a struct that represents the authentication service
type AuthService struct {
	repo            repository.IAuthRepository
	verificationSvc IVerificationService
	sessionRepo     repository.IUserSessionRepository
}

// NewAuthService creates a new instance of AuthService
func NewAuthService(repo repository.IAuthRepository, verificationSvc IVerificationService,
	sessionRepo repository.IUserSessionRepository) IAuthService {
	return &AuthService{
		repo:            repo,
		verificationSvc: verificationSvc,
		sessionRepo:     sessionRepo,
	}
}

// Register creates a new user and profile
func (s *AuthService) Register(ctx context.Context, tx *gorm.DB, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	// Validasi
	if !s.repo.WithTx(tx).IsEmailUnique(ctx, req.Email) || !s.repo.WithTx(tx).IsPhoneUnique(ctx, req.Phone) {
		return nil, errors.New("Email atau nomor telephone sudah digunakan")
	}

	// Mapping & hash
	var user models.User
	copier.Copy(&user, req)
	user.RoleType = models.RoleTypeSiswa
	user.Password, _ = utils.HashPassword(req.Password)

	groupVerified := req.GroupType == models.Umum

	if err := s.repo.WithTx(tx).CreateUser(ctx, &user); err != nil {
		return nil, err
	}

	// Generate kode verifikasi
	code, err := s.verificationSvc.GenerateVerificationCode(ctx, tx, user.ID)
	if err != nil {
		return nil, err
	}

	// Kirim email async
	token, _ := utils.GenerateVerificationToken(user.ID, user.Email)
	go utils.SendVerificationEmail(user.Email, code, token)

	// Create Profile
	var profile models.Profile
	copier.Copy(&profile, req)
	profile.UserID = user.ID
	profile.GroupVerified = groupVerified

	if err := s.repo.WithTx(tx).CreateProfile(ctx, &profile); err != nil {
		return nil, err
	}

	// Buat response
	fullUser, err := s.repo.WithTx(tx).GetUserByID(ctx, user.ID)
	if err != nil {
		fmt.Println("Error fetching full user:", err)
		return nil, err
	}

	fmt.Println("User registered successfully:", user)
	fmt.Println("User registered successfully 2:", fullUser)
	var resp dto.RegisterResponse
	copier.Copy(&resp, &fullUser)
	copier.Copy(&resp.Profile, &profile)

	return &resp, nil
}

// Login authenticates a user and returns access and refresh tokens
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest, c *fiber.Ctx) (*dto.LoginResult, error) {

	user, err := s.repo.GetUserByEmailWithProfile(ctx, req.Email)
	if err != nil {
		return nil, errors.New("Email tidak ditemukan")
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return nil, errors.New("Password salah")
	}

	if !user.IsVerified {
		return nil, errors.New("Email belum diverifikasi")
	}

	accessTokenSecret := config.GetEnv("ACCESS_TOKEN_SECRET", "default-key")
	accessTokenExpiryStr := config.GetEnv("ACCESS_TOKEN_EXPIRY_HOURS", "24")
	accessTokenExpiryHours, err := strconv.Atoi(accessTokenExpiryStr)
	if err != nil {
		accessTokenExpiryHours = 24
	}
	accessToken, err := utils.GenerateJWT(*user, accessTokenSecret, accessTokenExpiryHours)
	if err != nil {
		return nil, fmt.Errorf("gagal generate access token: %w", err)
	}

	refreshTokenSecret := config.GetEnv("REFRESH_TOKEN_SECRET", "default-key")
	refreshTokenExpiryStr := config.GetEnv("REFRESH_TOKEN_EXPIRY_HOURS", "24")
	refreshTokenExpiryHours, err := strconv.Atoi(refreshTokenExpiryStr)
	if err != nil {
		refreshTokenExpiryHours = 24
	}
	refreshToken, err := utils.GenerateJWT(*user, refreshTokenSecret, refreshTokenExpiryHours)
	if err != nil {
		return nil, fmt.Errorf("gagal generate refresh token: %w", err)
	}

	if err := s.repo.CreateUserSession(ctx, user.ID, refreshToken, c); err != nil {
		return nil, fmt.Errorf("gagal menyimpan sesi user: %w", err)
	}

	// Mapping ke DTO response
	var userResponse dto.UserResponse
	copier.Copy(&userResponse, &user)
	copier.Copy(&userResponse.Profile, &user.Profile)

	return &dto.LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userResponse,
	}, nil
}

// VerifyUserEmail verifies the user's email using a verification token and code
func (s *AuthService) VerifyUserEmail(ctx context.Context, token, code string) error {
	jwtSecret := config.GetEnv("VERIFICATION_TOKEN_SECRET", "default-key")

	payload, err := utils.ExtractUserIDFromToken(token, jwtSecret)
	if err != nil {
		return fmt.Errorf("token tidak valid: %w", err)
	}

	user, err := s.repo.GetUserByID(ctx, payload.UserID)
	if err != nil {
		return fmt.Errorf("user tidak ditemukan: %w", err)
	}

	if user.IsVerified {
		return fmt.Errorf("email sudah diverifikasi")
	}

	isValid := s.verificationSvc.VerifyCode(ctx, user.ID, code)
	if !isValid {
		return fmt.Errorf("kode verifikasi salah atau kadaluarsa")
	}

	return nil
}

// ResendVerificationCode resends the verification code to the user's email
func (s *AuthService) ResendVerificationCode(ctx context.Context, token string) error {
	jwtSecret := config.GetEnv("VERIFICATION_TOKEN_SECRET", "default-key")

	payload, err := utils.ExtractUserIDFromToken(token, jwtSecret)
	if err != nil {
		return fmt.Errorf("token tidak valid")
	}

	user, err := s.repo.GetUserByID(ctx, payload.UserID)
	if err != nil {
		return fmt.Errorf("user tidak ditemukan: %w", err)
	}

	if user.IsVerified {
		return fmt.Errorf("email sudah diverifikasi")
	}

	// Cek cooldown
	remaining, err := s.verificationSvc.GetCooldownRemaining(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("gagal cek cooldown: %w", err)
	}
	if remaining > 0 {
		return fmt.Errorf("tunggu %d detik untuk mengirim ulang", int(remaining.Seconds()))
	}

	// Generate ulang kode
	code, err := s.verificationSvc.GenerateVerificationCode(ctx, nil, user.ID)
	if err != nil {
		return fmt.Errorf("gagal generate kode: %w", err)
	}

	// Generate token verifikasi baru
	token, err = utils.GenerateVerificationToken(user.ID, user.Email)
	if err != nil {
		return fmt.Errorf("gagal buat token verifikasi: %w", err)
	}

	// Kirim email (async)
	go utils.SendVerificationEmail(user.Email, code, token)

	return nil
}

// RefreshTokens refreshes the access token using the provided refresh token
func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*dto.LoginResult, error) {
	// Verifikasi refresh token
	refreshSecret := config.GetEnv("REFRESH_TOKEN_SECRET", "default-key")
	claims, err := utils.ExtractUserIDFromToken(refreshToken, refreshSecret)
	if err != nil {
		return nil, fmt.Errorf("refresh token tidak valid: %w", err)
	}

	// Ambil user dari DB
	user, err := s.repo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user tidak ditemukan: %w", err)
	}

	// Buat access token baru
	accessSecret := config.GetEnv("ACCESS_TOKEN_SECRET", "default-key")
	accessToken := ""
	accessExpiry := config.GetIntEnv("ACCESS_TOKEN_EXPIRY_HOURS", 24)

	accessToken, err = utils.GenerateJWT(*user, accessSecret, accessExpiry)
	if err != nil {
		return nil, fmt.Errorf("gagal generate token baru: %w", err)
	}

	// Mapping response
	var userResp dto.UserResponse
	_ = copier.Copy(&userResp, &user)
	_ = copier.Copy(&userResp.Profile, &user.Profile)

	return &dto.LoginResult{
		AccessToken: accessToken,
		User:        userResp,
	}, nil
}

// LogoutUser logs out the user by revoking their session and blacklisting the access token
func (s *AuthService) LogoutUser(ctx context.Context, accessToken, refreshToken string) error {
	// Revoke session
	if err := s.RevokeUserSessionByRefreshToken(ctx, refreshToken); err != nil {
		return fmt.Errorf("gagal revoke session: %w", err)
	}

	// Masukkan access token ke Redis blacklist
	ttlStr := config.GetEnv("TOKEN_BLACKLIST_TTL", "86400") // default 24 jam
	ttl, err := strconv.Atoi(ttlStr)
	if err != nil || ttl <= 0 {
		ttl = 86400
	}

	err = config.RedisClient.Set(config.Ctx, accessToken, "blacklisted", time.Duration(ttl)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("gagal blacklist token: %w", err)
	}

	return nil
}

// RevokeUserSessionByRefreshToken revokes a user session by its refresh token
func (s *AuthService) RevokeUserSessionByRefreshToken(ctx context.Context, refreshToken string) error {
	session, err := s.sessionRepo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return fmt.Errorf("refresh token session not found: %w", err)
	}

	session.IsRevoked = true
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return fmt.Errorf("gagal revoke session: %w", err)
	}

	return nil
}
