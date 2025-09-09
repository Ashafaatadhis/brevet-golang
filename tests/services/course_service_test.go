package services

import (
	"brevet-api/dto"
	"brevet-api/mocks"
	"brevet-api/models"
	"brevet-api/services"
	"brevet-api/utils"
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
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

func TestCourseService_GetAllFilteredCourses(t *testing.T) {
	ctx := context.Background()

	t.Run("success - get filtered courses", func(t *testing.T) {
		repo := new(mocks.ICourseRepository)
		fileService := new(mocks.IFileService)
		db, _ := setupMockDB(t)
		service := services.NewCourseService(repo, db, fileService)

		opts := utils.QueryOptions{
			Limit:   10,
			Offset:  0,
			Filters: map[string]string{"title": "golang"},
		}

		courses := []models.Course{
			{ID: uuid.New(), Title: "Belajar Golang"},
			{ID: uuid.New(), Title: "Mastering Golang"},
		}
		total := int64(2)

		// Mock repo behavior
		repo.On("GetAllFilteredCourses", ctx, opts).
			Return(courses, total, nil)

		result, count, err := service.GetAllFilteredCourses(ctx, opts)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, total, count)
		assert.Equal(t, "Belajar Golang", result[0].Title)

		repo.AssertExpectations(t)
	})

	t.Run("fail - repository error", func(t *testing.T) {
		repo := new(mocks.ICourseRepository)
		fileService := new(mocks.IFileService)
		db, _ := setupMockDB(t)
		service := services.NewCourseService(repo, db, fileService)

		opts := utils.QueryOptions{Limit: 5, Offset: 0}

		repo.On("GetAllFilteredCourses", ctx, opts).
			Return(nil, int64(0), errors.New("db error"))

		result, count, err := service.GetAllFilteredCourses(ctx, opts)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, int64(0), count)

		repo.AssertExpectations(t)
	})
}

func TestCourseService_GetCourseBySlug(t *testing.T) {
	ctx := context.Background()

	t.Run("success - get course by slug", func(t *testing.T) {
		repo := new(mocks.ICourseRepository)
		fileService := new(mocks.IFileService)
		db, _ := setupMockDB(t)
		service := services.NewCourseService(repo, db, fileService)

		slug := "belajar-golang"
		course := &models.Course{
			ID:    uuid.New(),
			Slug:  slug,
			Title: "Belajar Golang",
		}

		// Mock repo behavior
		repo.On("GetCourseBySlug", ctx, slug).
			Return(course, nil)

		result, err := service.GetCourseBySlug(ctx, slug)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Belajar Golang", result.Title)

		repo.AssertExpectations(t)
	})

	t.Run("fail - repository error", func(t *testing.T) {
		repo := new(mocks.ICourseRepository)
		fileService := new(mocks.IFileService)
		db, _ := setupMockDB(t)
		service := services.NewCourseService(repo, db, fileService)

		slug := "unknown-course"

		repo.On("GetCourseBySlug", ctx, slug).
			Return(nil, errors.New("db error"))

		result, err := service.GetCourseBySlug(ctx, slug)

		assert.Error(t, err)
		assert.Nil(t, result)

		repo.AssertExpectations(t)
	})
}
func TestCourseService_CreateCourse(t *testing.T) {
	ctx := context.Background()

	t.Run("success - create course", func(t *testing.T) {
		repo := new(mocks.ICourseRepository)
		fileService := new(mocks.IFileService)
		db, mock := setupMockDB(t)
		service := services.NewCourseService(repo, db, fileService)

		body := &dto.CreateCourseRequest{
			Title:            "Belajar Golang",
			ShortDescription: "short desc",
			Description:      "desc",
			LearningOutcomes: "outcomes",
			Achievements:     "achievements",
			CourseImages: []dto.CourseImageRequest{
				{ImageURL: "http://example.com/img1.png"},
				{ImageURL: "http://example.com/img2.png"},
			},
		}

		courseID := uuid.New()
		expectedCourse := &models.Course{
			ID:               courseID,
			Slug:             "belajar-golang",
			Title:            body.Title,
			ShortDescription: body.ShortDescription,
			Description:      body.Description,
			LearningOutcomes: body.LearningOutcomes,
			Achievements:     body.Achievements,
		}

		// Mock transaction
		mock.ExpectBegin()
		mock.ExpectCommit()

		// Mock repository calls
		repo.On("WithTx", testifymock.Anything).Return(repo)
		repo.On("IsSlugExists", ctx, "belajar-golang").Return(false)
		repo.On("Create", ctx, testifymock.AnythingOfType("*models.Course")).
			Run(func(args testifymock.Arguments) {
				c := args.Get(1).(*models.Course)
				c.ID = courseID // assign ID after create
			}).
			Return(nil)

		repo.On("CreateCourseImagesBulk", ctx, testifymock.AnythingOfType("[]models.CourseImage")).
			Return(nil)

		repo.On("FindByIDWithImages", ctx, courseID).
			Return(expectedCourse, nil)

		result, err := service.CreateCourse(ctx, body)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, body.Title, result.Title)
		assert.Equal(t, "belajar-golang", result.Slug)

		repo.AssertExpectations(t)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("fail - repo.Create error", func(t *testing.T) {
		repo := new(mocks.ICourseRepository)
		fileService := new(mocks.IFileService)
		db, mock := setupMockDB(t)
		service := services.NewCourseService(repo, db, fileService)

		body := &dto.CreateCourseRequest{Title: "Fail Course"}

		// Mock transaction - expect begin and rollback due to error
		mock.ExpectBegin()
		mock.ExpectRollback()

		repo.On("WithTx", testifymock.Anything).Return(repo)
		repo.On("IsSlugExists", ctx, "fail-course").Return(false) // Fix: correct slug
		repo.On("Create", ctx, testifymock.AnythingOfType("*models.Course")).
			Return(errors.New("insert failed"))

		result, err := service.CreateCourse(ctx, body)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "insert failed")

		repo.AssertExpectations(t)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("fail - CreateCourseImagesBulk error", func(t *testing.T) {
		repo := new(mocks.ICourseRepository)
		fileService := new(mocks.IFileService)
		db, mock := setupMockDB(t)
		service := services.NewCourseService(repo, db, fileService)

		body := &dto.CreateCourseRequest{
			Title: "Belajar Golang",
			CourseImages: []dto.CourseImageRequest{
				{ImageURL: "http://example.com/img.png"},
			},
		}

		courseID := uuid.New()

		// Mock transaction - expect begin and rollback due to error
		mock.ExpectBegin()
		mock.ExpectRollback()

		repo.On("WithTx", testifymock.Anything).Return(repo)
		repo.On("IsSlugExists", ctx, "belajar-golang").Return(false)
		repo.On("Create", ctx, testifymock.AnythingOfType("*models.Course")).
			Run(func(args testifymock.Arguments) {
				c := args.Get(1).(*models.Course)
				c.ID = courseID
			}).
			Return(nil)

		repo.On("CreateCourseImagesBulk", ctx, testifymock.AnythingOfType("[]models.CourseImage")).
			Return(errors.New("bulk insert failed"))

		result, err := service.CreateCourse(ctx, body)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "bulk insert failed")

		repo.AssertExpectations(t)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("fail - FindByIDWithImages error", func(t *testing.T) {
		repo := new(mocks.ICourseRepository)
		fileService := new(mocks.IFileService)
		db, mock := setupMockDB(t)
		service := services.NewCourseService(repo, db, fileService)

		body := &dto.CreateCourseRequest{
			Title: "Belajar Golang",
			CourseImages: []dto.CourseImageRequest{
				{ImageURL: "http://example.com/img.png"},
			},
		}

		courseID := uuid.New()

		// Mock transaction - expect begin and rollback due to error
		mock.ExpectBegin()
		mock.ExpectRollback()

		repo.On("WithTx", testifymock.Anything).Return(repo)
		repo.On("IsSlugExists", ctx, "belajar-golang").Return(false)
		repo.On("Create", ctx, testifymock.AnythingOfType("*models.Course")).
			Run(func(args testifymock.Arguments) {
				c := args.Get(1).(*models.Course)
				c.ID = courseID
			}).
			Return(nil)

		repo.On("CreateCourseImagesBulk", ctx, testifymock.AnythingOfType("[]models.CourseImage")).
			Return(nil)

		repo.On("FindByIDWithImages", ctx, courseID).
			Return(nil, errors.New("find failed"))

		result, err := service.CreateCourse(ctx, body)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "find failed")

		repo.AssertExpectations(t)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCourseService_UpdateCourse(t *testing.T) {
	ctx := context.Background()

	t.Run("success - update course", func(t *testing.T) {
		repo := new(mocks.ICourseRepository)
		fileService := new(mocks.IFileService)
		db, mock := setupMockDB(t)
		service := services.NewCourseService(repo, db, fileService)

		courseID := uuid.New()
		body := &dto.UpdateCourseRequest{
			Title: ptrToString("Updated Title"),
			CourseImages: &[]dto.CourseImageRequest{
				{ImageURL: "http://example.com/newimg.png"},
			},
		}

		existingCourse := &models.Course{
			ID:    courseID,
			Title: "Old Title",
		}
		updatedCourse := &models.Course{
			ID:    courseID,
			Title: "Updated Title",
		}

		// Mock transaction
		mock.ExpectBegin()
		mock.ExpectCommit()

		// Mock repository
		repo.On("WithTx", testifymock.Anything).Return(repo)
		repo.On("FindByID", ctx, courseID).Return(existingCourse, nil)
		repo.On("Update", ctx, existingCourse).Return(nil)
		repo.On("DeleteCourseImagesByCourseID", ctx, courseID).Return(nil)
		repo.On("CreateCourseImagesBulk", ctx, testifymock.AnythingOfType("[]models.CourseImage")).Return(nil)
		repo.On("FindByIDWithImages", ctx, courseID).Return(updatedCourse, nil)

		result, err := service.UpdateCourse(ctx, courseID, body)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Updated Title", result.Title)

		repo.AssertExpectations(t)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("fail - FindByID error", func(t *testing.T) {
		repo := new(mocks.ICourseRepository)
		fileService := new(mocks.IFileService)
		db, mock := setupMockDB(t)
		service := services.NewCourseService(repo, db, fileService)

		courseID := uuid.New()
		body := &dto.UpdateCourseRequest{Title: ptrToString("Updated Title")}

		mock.ExpectBegin()
		mock.ExpectRollback()

		repo.On("WithTx", testifymock.Anything).Return(repo)
		repo.On("FindByID", ctx, courseID).Return(nil, errors.New("not found"))

		result, err := service.UpdateCourse(ctx, courseID, body)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not found")

		repo.AssertExpectations(t)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Tambahkan skenario fail lain sesuai kebutuhan
}

func ptrToString(s string) *string {
	return &s
}

func TestCourseService_DeleteCourse(t *testing.T) {
	ctx := context.Background()

	t.Run("success - delete course", func(t *testing.T) {
		repo := new(mocks.ICourseRepository)
		fileService := new(mocks.IFileService)
		db, mock := setupMockDB(t)
		service := services.NewCourseService(repo, db, fileService)

		courseID := uuid.New()
		course := &models.Course{
			ID: courseID,
			CourseImages: []models.CourseImage{
				{ImageURL: "http://example.com/img1.png"},
				{ImageURL: "http://example.com/img2.png"},
			},
		}

		// Mock transaction
		mock.ExpectBegin()
		mock.ExpectCommit()

		// Mock repository
		repo.On("WithTx", testifymock.Anything).Return(repo)
		repo.On("FindByIDWithImages", ctx, courseID).Return(course, nil)
		repo.On("DeleteByID", ctx, courseID).Return(nil)

		// Mock file service
		fileService.On("DeleteFile", "http://example.com/img1.png").Return(nil)
		fileService.On("DeleteFile", "http://example.com/img2.png").Return(nil)

		err := service.DeleteCourse(ctx, courseID)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
		fileService.AssertExpectations(t)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("fail - FindByIDWithImages error", func(t *testing.T) {
		repo := new(mocks.ICourseRepository)
		fileService := new(mocks.IFileService)
		db, mock := setupMockDB(t)
		service := services.NewCourseService(repo, db, fileService)

		courseID := uuid.New()

		mock.ExpectBegin()
		mock.ExpectRollback()

		repo.On("WithTx", testifymock.Anything).Return(repo)
		repo.On("FindByIDWithImages", ctx, courseID).Return(nil, errors.New("not found"))

		err := service.DeleteCourse(ctx, courseID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		repo.AssertExpectations(t)
		fileService.AssertExpectations(t)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("fail - DeleteByID error", func(t *testing.T) {
		repo := new(mocks.ICourseRepository)
		fileService := new(mocks.IFileService)
		db, mock := setupMockDB(t)
		service := services.NewCourseService(repo, db, fileService)

		courseID := uuid.New()
		course := &models.Course{ID: courseID}

		mock.ExpectBegin()
		mock.ExpectRollback()

		repo.On("WithTx", testifymock.Anything).Return(repo)
		repo.On("FindByIDWithImages", ctx, courseID).Return(course, nil)
		repo.On("DeleteByID", ctx, courseID).Return(errors.New("delete failed"))

		err := service.DeleteCourse(ctx, courseID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete failed")
		repo.AssertExpectations(t)
		fileService.AssertExpectations(t)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success - DeleteFile error logs only", func(t *testing.T) {
		repo := new(mocks.ICourseRepository)
		fileService := new(mocks.IFileService)
		db, mock := setupMockDB(t)
		service := services.NewCourseService(repo, db, fileService)

		courseID := uuid.New()
		course := &models.Course{
			ID: courseID,
			CourseImages: []models.CourseImage{
				{ImageURL: "http://example.com/img1.png"},
			},
		}

		mock.ExpectBegin()
		mock.ExpectCommit()

		repo.On("WithTx", testifymock.Anything).Return(repo)
		repo.On("FindByIDWithImages", ctx, courseID).Return(course, nil)
		repo.On("DeleteByID", ctx, courseID).Return(nil)

		fileService.On("DeleteFile", "http://example.com/img1.png").Return(errors.New("failed delete"))

		err := service.DeleteCourse(ctx, courseID)

		assert.NoError(t, err) // tetap success karena transaction committed
		repo.AssertExpectations(t)
		fileService.AssertExpectations(t)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
