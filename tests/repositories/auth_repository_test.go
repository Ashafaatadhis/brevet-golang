package repositories

import (
	"brevet-api/models"
	"brevet-api/repository"
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating sqlmock: %s", err)
	}

	gdb, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("error opening gorm db: %s", err)
	}

	return gdb, mock
}

func TestIsEmailUnique(t *testing.T) {
	ctx := context.Background()

	t.Run("email unik", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewAuthRepository(db)

		email := "test@example.com"

		// GORM akan query `SELECT * FROM "users" WHERE email = $1 ORDER BY "users"."id" LIMIT 1`
		mock.ExpectQuery(`SELECT.*FROM "users"`).
			WithArgs(email, 1).                                      // tambahin argumen ke-2 = LIMIT
			WillReturnRows(sqlmock.NewRows([]string{"id", "email"})) // kosong

		result := repo.IsEmailUnique(ctx, email)
		assert.True(t, result)
	})

	t.Run("email sudah ada", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewAuthRepository(db)

		email := "exist@example.com"

		mock.ExpectQuery(`SELECT.*FROM "users"`).
			WithArgs(email, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).
				AddRow(uuid.New(), email))

		result := repo.IsEmailUnique(ctx, email)
		assert.False(t, result)
	})

	t.Run("db error", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewAuthRepository(db)

		email := "error@example.com"

		mock.ExpectQuery(`SELECT.*FROM "users"`).
			WithArgs(email, 1).
			WillReturnError(errors.New("db error"))

		result := repo.IsEmailUnique(ctx, email)
		assert.False(t, result) // karena error selain ErrRecordNotFound dianggap email tidak unik
	})
}

func TestIsPhoneUnique(t *testing.T) {
	ctx := context.Background()

	t.Run("phone unik", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewAuthRepository(db)

		phone := "08123456789"

		// GORM akan query: SELECT * FROM "users" WHERE phone = $1 ORDER BY "users"."id" LIMIT 1
		mock.ExpectQuery(`SELECT.*FROM "users"`).
			WithArgs(phone, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "phone"})) // kosong → artinya tidak ada user dengan phone itu

		result := repo.IsPhoneUnique(ctx, phone)
		assert.True(t, result)
	})

	t.Run("phone tidak unik", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewAuthRepository(db)

		phone := "08123456789"

		// Query yang sama, tapi kita balikin row dummy (artinya phone sudah dipakai user lain)
		mock.ExpectQuery(`SELECT.*FROM "users"`).
			WithArgs(phone, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "phone"}).
				AddRow(1, phone))

		result := repo.IsPhoneUnique(ctx, phone)
		assert.False(t, result)
	})

	t.Run("db error", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewAuthRepository(db)

		phone := "08123456789"

		mock.ExpectQuery(`SELECT.*FROM "users"`).
			WithArgs(phone, 1).
			WillReturnError(errors.New("db error"))

		result := repo.IsPhoneUnique(ctx, phone)
		assert.False(t, result) // karena error selain ErrRecordNotFound dianggap email tidak unik
	})
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("success - create user", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewAuthRepository(db)

		user := &models.User{
			ID: uuid.New(),
		}

		// GORM biasanya pakai transaction
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "users"`).
			WithArgs(
				sqlmock.AnyArg(), // name
				sqlmock.AnyArg(), // phone
				sqlmock.AnyArg(), // avatar
				sqlmock.AnyArg(), // email
				sqlmock.AnyArg(), // password
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
				sqlmock.AnyArg(), // is_verified
				sqlmock.AnyArg(), // verify_code
				sqlmock.AnyArg(), // code_expiry
				sqlmock.AnyArg(), // role_type
				sqlmock.AnyArg(), // id
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(user.ID))
		mock.ExpectCommit()

		err := repo.CreateUser(ctx, user)
		assert.NoError(t, err)
	})

	t.Run("fail - insert error", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewAuthRepository(db)

		user := &models.User{
			ID: uuid.New(),
		}

		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "users"`).
			WithArgs(
				sqlmock.AnyArg(), // name
				sqlmock.AnyArg(), // phone
				sqlmock.AnyArg(), // avatar
				sqlmock.AnyArg(), // email
				sqlmock.AnyArg(), // password
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
				sqlmock.AnyArg(), // is_verified
				sqlmock.AnyArg(), // verify_code
				sqlmock.AnyArg(), // code_expiry
				sqlmock.AnyArg(), // role_type
				sqlmock.AnyArg(), // id
			).
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.CreateUser(ctx, user)
		assert.Error(t, err)
	})
}

func TestCreateProfile(t *testing.T) {
	ctx := context.Background()

	t.Run("success - create profile", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewAuthRepository(db)

		profile := &models.Profile{
			ID:     uuid.New(),
			UserID: uuid.New(),
		}

		// GORM biasanya pakai transaction
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "profiles"`).
			WithArgs(
				sqlmock.AnyArg(), // id
				sqlmock.AnyArg(), // user_id
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(profile.ID))
		mock.ExpectCommit()

		err := repo.CreateProfile(ctx, profile)
		assert.NoError(t, err)
	})

	t.Run("fail - insert error", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewAuthRepository(db)

		profile := &models.Profile{
			ID:     uuid.New(),
			UserID: uuid.New(),
		}

		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "profiles"`).
			WithArgs(
				sqlmock.AnyArg(), // id
				sqlmock.AnyArg(), // user_id
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
			).
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.CreateProfile(ctx, profile)
		assert.Error(t, err)
	})
}

func TestGetUsers(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := repository.NewAuthRepository(db)

	t.Run("success - found user", func(t *testing.T) {
		ctx := context.Background()
		userID := uuid.New()

		expectedUser := &models.User{
			ID:    userID,
			Email: "test@example.com",
			Phone: "08123456789",
		}

		rows := sqlmock.NewRows([]string{"id", "email", "phone"}).
			AddRow(expectedUser.ID, expectedUser.Email, expectedUser.Phone)

		mock.ExpectQuery(
			regexp.QuoteMeta(`SELECT * FROM "users" WHERE id = $1 ORDER BY "users"."id" LIMIT $2`),
		).WithArgs(userID, 1).
			WillReturnRows(rows)

		user, err := repo.GetUsers(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.Equal(t, expectedUser.Email, user.Email)
		assert.Equal(t, expectedUser.Phone, user.Phone)
	})

	t.Run("error - not found", func(t *testing.T) {
		ctx := context.Background()
		userID := uuid.New()

		mock.ExpectQuery(
			regexp.QuoteMeta(`SELECT * FROM "users" WHERE id = $1 ORDER BY "users"."id" LIMIT $2`),
		).WithArgs(userID, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		user, err := repo.GetUsers(ctx, userID)
		require.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestGetUserByEmail(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := repository.NewAuthRepository(db)

	t.Run("success - found user", func(t *testing.T) {
		ctx := context.Background()
		email := "test@example.com"
		userID := uuid.New()

		expectedUser := &models.User{
			ID:    userID,
			Email: email,
		}

		// Query user
		userRows := sqlmock.NewRows([]string{"id", "email"}).
			AddRow(expectedUser.ID, expectedUser.Email)

		mock.ExpectQuery(
			regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 ORDER BY "users"."id" LIMIT $2`),
		).WithArgs(email, 1).
			WillReturnRows(userRows)

		user, err := repo.GetUserByEmail(ctx, email)
		require.NoError(t, err)
		assert.Equal(t, expectedUser.Email, user.Email)
	})

	t.Run("error - user not found", func(t *testing.T) {
		ctx := context.Background()
		email := "notfound@example.com"

		mock.ExpectQuery(
			regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 ORDER BY "users"."id" LIMIT $2`),
		).WithArgs(email, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		user, err := repo.GetUserByEmail(ctx, email)
		require.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestGetUserByEmailWithProfile(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := repository.NewAuthRepository(db)

	t.Run("success - found user with profile", func(t *testing.T) {
		ctx := context.Background()
		email := "test@example.com"
		userID := uuid.New()

		expectedUser := &models.User{
			ID:    userID,
			Email: email,
		}
		expectedProfile := &models.Profile{
			ID:          uuid.New(),
			UserID:      userID,
			Institution: "Test Institution",
		}

		// Query user
		userRows := sqlmock.NewRows([]string{"id", "email"}).
			AddRow(expectedUser.ID, expectedUser.Email)

		mock.ExpectQuery(
			regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 ORDER BY "users"."id" LIMIT $2`),
		).WithArgs(email, 1).
			WillReturnRows(userRows)

		// Query preload (LEFT JOIN kalau pointer)
		profileRows := sqlmock.NewRows([]string{"id", "user_id", "institution"}).
			AddRow(expectedProfile.ID, expectedProfile.UserID, expectedProfile.Institution)

		mock.ExpectQuery(
			regexp.QuoteMeta(`SELECT * FROM "profiles" WHERE "profiles"."user_id" = $1`),
		).WithArgs(userID).
			WillReturnRows(profileRows)

		user, err := repo.GetUserByEmailWithProfile(ctx, email)
		require.NoError(t, err)
		assert.Equal(t, expectedUser.Email, user.Email)
		require.NotNil(t, user.Profile)
		assert.Equal(t, expectedProfile.Institution, user.Profile.Institution)
	})

	t.Run("error - user not found", func(t *testing.T) {
		ctx := context.Background()
		email := "notfound@example.com"

		mock.ExpectQuery(
			regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 ORDER BY "users"."id" LIMIT $2`),
		).WithArgs(email, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		user, err := repo.GetUserByEmailWithProfile(ctx, email)
		require.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestGetUserByIDWithProfile(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := repository.NewAuthRepository(db)

	t.Run("success - found user with profile", func(t *testing.T) {
		ctx := context.Background()
		email := "test@example.com"
		userID := uuid.New()

		expectedUser := &models.User{
			ID:    userID,
			Email: email,
		}
		expectedProfile := &models.Profile{
			ID:          uuid.New(),
			UserID:      userID,
			Institution: "Test Institution",
		}

		// Query user
		userRows := sqlmock.NewRows([]string{"id", "email"}).
			AddRow(expectedUser.ID, expectedUser.Email)

		mock.ExpectQuery(
			regexp.QuoteMeta(`SELECT * FROM "users" WHERE id = $1 ORDER BY "users"."id" LIMIT $2`),
		).WithArgs(userID, 1).
			WillReturnRows(userRows)

		// Query preload (LEFT JOIN kalau pointer)
		profileRows := sqlmock.NewRows([]string{"id", "user_id", "institution"}).
			AddRow(expectedProfile.ID, expectedProfile.UserID, expectedProfile.Institution)

		mock.ExpectQuery(
			regexp.QuoteMeta(`SELECT * FROM "profiles" WHERE "profiles"."user_id" = $1`),
		).WithArgs(userID).
			WillReturnRows(profileRows)

		user, err := repo.GetUserByIDWithProfile(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedUser.Email, user.Email)
		require.NotNil(t, user.Profile)
		assert.Equal(t, expectedProfile.Institution, user.Profile.Institution)
	})

	t.Run("error - user not found", func(t *testing.T) {
		ctx := context.Background()

		userID := uuid.New()

		mock.ExpectQuery(
			regexp.QuoteMeta(`SELECT * FROM "users" WHERE id = $1 ORDER BY "users"."id" LIMIT $2`),
		).WithArgs(userID, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		user, err := repo.GetUserByIDWithProfile(ctx, userID)
		require.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestGetUserByID(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := repository.NewAuthRepository(db)

	t.Run("success - found user with profile", func(t *testing.T) {
		ctx := context.Background()
		email := "test@example.com"
		userID := uuid.New()

		expectedUser := &models.User{
			ID:    userID,
			Email: email,
		}
		expectedProfile := &models.Profile{
			ID:          uuid.New(),
			UserID:      userID,
			Institution: "Test Institution",
		}

		// Query user
		userRows := sqlmock.NewRows([]string{"id", "email"}).
			AddRow(expectedUser.ID, expectedUser.Email)

		mock.ExpectQuery(
			regexp.QuoteMeta(`SELECT * FROM "users" WHERE id = $1 ORDER BY "users"."id" LIMIT $2`),
		).WithArgs(userID, 1).
			WillReturnRows(userRows)

		// Query preload (LEFT JOIN kalau pointer)
		profileRows := sqlmock.NewRows([]string{"id", "user_id", "institution"}).
			AddRow(expectedProfile.ID, expectedProfile.UserID, expectedProfile.Institution)

		mock.ExpectQuery(
			regexp.QuoteMeta(`SELECT * FROM "profiles" WHERE "profiles"."user_id" = $1`),
		).WithArgs(userID).
			WillReturnRows(profileRows)

		user, err := repo.GetUserByID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedUser.Email, user.Email)
		require.NotNil(t, user.Profile)
		assert.Equal(t, expectedProfile.Institution, user.Profile.Institution)
	})

	t.Run("error - user not found", func(t *testing.T) {
		ctx := context.Background()

		userID := uuid.New()

		mock.ExpectQuery(
			regexp.QuoteMeta(`SELECT * FROM "users" WHERE id = $1 ORDER BY "users"."id" LIMIT $2`),
		).WithArgs(userID, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		user, err := repo.GetUserByID(ctx, userID)
		require.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestCreateUserSession(t *testing.T) {
	ctx := context.Background()

	t.Run("success - create session", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewAuthRepository(db)

		userID := uuid.New()
		refreshToken := "dummy-refresh-token"

		app := fiber.New()
		reqCtx := &fasthttp.RequestCtx{}

		// Set User-Agent
		reqCtx.Request.Header.Set("User-Agent", "UnitTest")

		// Set RemoteAddr (manual isi)
		reqCtx.RemoteAddr()
		reqCtx.SetUserValue("remoteAddr", "127.0.0.1:5000")

		// Bungkus ke Fiber ctx
		c := app.AcquireCtx(reqCtx)
		defer app.ReleaseCtx(c)

		// Ekspektasi insert
		mock.ExpectBegin()
		sessionID := uuid.New() // buat UUID baru
		mock.ExpectQuery(`INSERT INTO "user_sessions"`).
			WithArgs(
				userID,           // user_id
				refreshToken,     // refresh_token
				"UnitTest",       // user_agent
				sqlmock.AnyArg(), // ip_address
				false,            // is_revoked
				sqlmock.AnyArg(), // expires_at
				sqlmock.AnyArg(), // created_at
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(sessionID))

		mock.ExpectCommit()

		err := repo.CreateUserSession(ctx, userID, refreshToken, c)
		assert.NoError(t, err)
	})

	t.Run("fail - insert error", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewAuthRepository(db)

		userID := uuid.New()
		refreshToken := "dummy-refresh-token"

		app := fiber.New()
		reqCtx := &fasthttp.RequestCtx{}

		// Set User-Agent
		reqCtx.Request.Header.Set("User-Agent", "UnitTest")

		// Set RemoteAddr (manual isi)
		reqCtx.RemoteAddr()
		reqCtx.SetUserValue("remoteAddr", "127.0.0.1:5000")

		// Bungkus ke Fiber ctx
		c := app.AcquireCtx(reqCtx)
		defer app.ReleaseCtx(c)

		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "user_sessions"`).
			WithArgs(
				userID,           // user_id
				refreshToken,     // refresh_token
				"UnitTest",       // user_agent
				sqlmock.AnyArg(), // ip_address, biar cocok walau 0.0.0.0
				false,            // is_revoked
				sqlmock.AnyArg(), // expires_at
				sqlmock.AnyArg(), // created_at
			).
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.CreateUserSession(ctx, userID, refreshToken, c)
		assert.Error(t, err)
	})
}

func TestRevokeUserSessionByRefreshToken(t *testing.T) {
	ctx := context.Background()

	t.Run("success - revoke session", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewAuthRepository(db)

		refreshToken := "dummy-refresh-token"
		sessionID := uuid.New()
		userID := uuid.New()

		// Mock SELECT session by refresh token (arg kedua = 1 karena LIMIT 1)
		rows := sqlmock.NewRows([]string{"id", "user_id", "refresh_token", "is_revoked"}).
			AddRow(sessionID, userID, refreshToken, false)

		mock.ExpectQuery(`SELECT \* FROM "user_sessions" WHERE refresh_token = .* ORDER BY .* LIMIT .*`).
			WithArgs(refreshToken, 1).
			WillReturnRows(rows)

		// Mock UPDATE is_revoked
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "user_sessions" SET .* WHERE "id" = .*`).
			WithArgs(
				sqlmock.AnyArg(), // user_id
				sqlmock.AnyArg(), // refresh_token
				sqlmock.AnyArg(), // user_agent
				sqlmock.AnyArg(), // ip_address
				sqlmock.AnyArg(), // is_revoked
				sqlmock.AnyArg(), // expires_at
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // id
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.RevokeUserSessionByRefreshToken(ctx, refreshToken)
		assert.NoError(t, err)
	})

	t.Run("fail - token not found", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewAuthRepository(db)

		refreshToken := "notfound-token"

		mock.ExpectQuery(`SELECT \* FROM "user_sessions" WHERE refresh_token = .* ORDER BY .* LIMIT .*`).
			WithArgs(refreshToken, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		err := repo.RevokeUserSessionByRefreshToken(ctx, refreshToken)
		assert.Error(t, err)
		assert.Equal(t, "refresh token session not found", err.Error())
	})

	t.Run("fail - update error", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewAuthRepository(db)

		refreshToken := "dummy-refresh-token"
		sessionID := uuid.New()
		userID := uuid.New()

		// SELECT session found
		rows := sqlmock.NewRows([]string{"id", "user_id", "refresh_token", "is_revoked"}).
			AddRow(sessionID, userID, refreshToken, false)

		mock.ExpectQuery(`SELECT \* FROM "user_sessions" WHERE refresh_token = .* ORDER BY .* LIMIT .*`).
			WithArgs(refreshToken, 1). // <=== Perhatikan arg kedua
			WillReturnRows(rows)

		// UPDATE gagal
		mock.ExpectBegin()

		mock.ExpectExec(`UPDATE "user_sessions" SET .* WHERE "id" = .*`).
			WithArgs(
				sqlmock.AnyArg(), // user_id
				sqlmock.AnyArg(), // refresh_token
				sqlmock.AnyArg(), // user_agent
				sqlmock.AnyArg(), // ip_address
				sqlmock.AnyArg(), // is_revoked
				sqlmock.AnyArg(), // expires_at
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // id
			).
			WillReturnError(errors.New("update error"))
		mock.ExpectRollback()

		err := repo.RevokeUserSessionByRefreshToken(ctx, refreshToken)
		assert.Error(t, err)
		assert.Equal(t, "update error", err.Error())
	})
}
