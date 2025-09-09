package repositories

import (
	"brevet-api/models"
	"brevet-api/repository"
	"brevet-api/utils"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestGetAllFilteredCourses(t *testing.T) {
	ctx := context.Background()

	t.Run("success - return courses", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewCourseRepository(db)

		opts := utils.QueryOptions{
			Limit:  10,
			Offset: 0,
			Sort:   "title",
			Order:  "asc",
			Search: "golang",
		}

		// Mock count query
		mock.ExpectQuery(`SELECT count\(\*\) FROM "courses"`).
			WithArgs("%" + opts.Search + "%").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		courseID := uuid.New() // simpan supaya sama dipakai di kedua query

		// Mock select query ke courses
		rows := sqlmock.NewRows([]string{
			"id", "slug", "title", "short_description", "description",
			"learning_outcomes", "achievements", "created_at", "updated_at",
		}).AddRow(courseID, "golang-course", "Belajar Golang", "short desc", "desc", "outcomes", "achievements", time.Now(), time.Now())

		mock.ExpectQuery(`SELECT.*FROM "courses"`).
			WithArgs("%"+opts.Search+"%", opts.Limit).
			WillReturnRows(rows)

		// Mock preload query ke course_images
		mock.ExpectQuery(`SELECT.*FROM "course_images"`).
			WithArgs(courseID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "url", "created_at", "updated_at",
			}))

		courses, total, err := repo.GetAllFilteredCourses(ctx, opts)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, courses, 1)
		assert.Equal(t, "Belajar Golang", courses[0].Title)
	})

	t.Run("db error", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewCourseRepository(db)

		opts := utils.QueryOptions{
			Limit:  10,
			Offset: 0,
		}

		mock.ExpectQuery(`SELECT count\(\*\) FROM "courses"`).
			WillReturnError(errors.New("db error"))

		courses, total, err := repo.GetAllFilteredCourses(ctx, opts)
		assert.Error(t, err)
		assert.Nil(t, courses)
		assert.Equal(t, int64(0), total)
	})
}
func TestGetCourseBySlug(t *testing.T) {
	ctx := context.Background()

	t.Run("success - return courses", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewCourseRepository(db)

		courseID := uuid.New() // simpan supaya sama dipakai di kedua query
		slug := "golang-course"
		// Mock select query ke courses
		rows := sqlmock.NewRows([]string{
			"id", "slug", "title", "short_description", "description",
			"learning_outcomes", "achievements", "created_at", "updated_at",
		}).AddRow(courseID, slug, "Belajar Golang", "short desc", "desc", "outcomes", "achievements", time.Now(), time.Now())

		mock.ExpectQuery(`SELECT .* FROM "courses" WHERE slug = \$1.*LIMIT.*`).
			WithArgs(slug, 1).
			WillReturnRows(rows)

		// Mock preload query ke course_images
		mock.ExpectQuery(`SELECT.*FROM "course_images"`).
			WithArgs(courseID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "course_id", "url", "created_at", "updated_at",
			}))

		courses, err := repo.GetCourseBySlug(ctx, slug)
		assert.NoError(t, err)
		assert.Equal(t, "Belajar Golang", courses.Title)
	})

	t.Run("db error", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewCourseRepository(db)

		slug := "golang-course"

		mock.ExpectQuery(`SELECT .* FROM "courses" WHERE slug = \$1.*LIMIT.*`).
			WithArgs(slug, 1).
			WillReturnError(errors.New("db error"))

		course, err := repo.GetCourseBySlug(ctx, slug)
		assert.Error(t, err)
		assert.Nil(t, course)
	})
}

func TestCreateCourse(t *testing.T) {
	ctx := context.Background()

	t.Run("success - create course", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewCourseRepository(db)

		courseID := uuid.New()
		now := time.Now()

		course := &models.Course{
			ID:               courseID,
			Slug:             "golang-course",
			Title:            "Belajar Golang",
			ShortDescription: "short desc",
			Description:      "desc",
			LearningOutcomes: "outcomes",
			Achievements:     "achievements",
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		// GORM Create akan jalanin transaction kecil
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "courses"`).
			WithArgs(
				course.Slug,
				course.Title,
				course.ShortDescription,
				course.Description,
				course.LearningOutcomes,
				course.Achievements,
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
				course.ID,
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(course.ID))

		mock.ExpectCommit()

		err := repo.Create(ctx, course)
		assert.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewCourseRepository(db)

		course := &models.Course{
			ID:    uuid.New(),
			Slug:  "golang-error",
			Title: "Should Fail",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "courses"`).
			WithArgs(
				course.Slug,
				course.Title,
				course.ShortDescription,
				course.Description,
				course.LearningOutcomes,
				course.Achievements,
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				course.ID,
			).
			WillReturnError(errors.New("insert failed"))
		mock.ExpectRollback()

		err := repo.Create(ctx, course)
		assert.Error(t, err)
	})
}

func TestCreateCourseImagesBulk(t *testing.T) {
	ctx := context.Background()

	t.Run("success - bulk insert images", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewCourseRepository(db)

		courseID := uuid.New()
		now := time.Now()

		images := []models.CourseImage{
			{
				ID:        uuid.New(),
				CourseID:  courseID,
				ImageURL:  "https://example.com/img1.png",
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				ID:        uuid.New(),
				CourseID:  courseID,
				ImageURL:  "https://example.com/img2.png",
				CreatedAt: now,
				UpdatedAt: now,
			},
		}

		// GORM Create untuk slice akan jadi batch insert
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "course_images"`).
			WithArgs(
				images[0].CourseID,
				images[0].ImageURL,
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
				images[0].ID,
				images[1].CourseID,
				images[1].ImageURL,
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				images[1].ID,
			).
			WillReturnRows(
				sqlmock.NewRows([]string{"id"}).
					AddRow(images[0].ID).
					AddRow(images[1].ID),
			)
		mock.ExpectCommit()

		err := repo.CreateCourseImagesBulk(ctx, images)
		assert.NoError(t, err)
	})

	t.Run("success - empty slice", func(t *testing.T) {
		db, _ := setupMockDB(t)
		repo := repository.NewCourseRepository(db)

		err := repo.CreateCourseImagesBulk(ctx, []models.CourseImage{})
		assert.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewCourseRepository(db)

		images := []models.CourseImage{
			{
				ID:       uuid.New(),
				CourseID: uuid.New(),
				ImageURL: "https://example.com/img.png",
			},
		}

		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "course_images"`).
			WithArgs(
				images[0].CourseID,
				images[0].ImageURL,
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				images[0].ID,
			).
			WillReturnError(errors.New("insert failed"))
		mock.ExpectRollback()

		err := repo.CreateCourseImagesBulk(ctx, images)
		assert.Error(t, err)
	})
}

func TestFindByIDWithImages(t *testing.T) {
	ctx := context.Background()

	t.Run("success - course with images", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewCourseRepository(db)

		courseID := uuid.New()
		now := time.Now()

		// --- mock query untuk course ---
		rows := sqlmock.NewRows([]string{
			"id", "slug", "title", "short_description", "description",
			"learning_outcomes", "achievements", "created_at", "updated_at",
		}).AddRow(
			courseID,
			"golang-course",
			"Belajar Golang",
			"short desc",
			"desc",
			"outcomes",
			"achievements",
			now,
			now,
		)

		mock.ExpectQuery(`SELECT .* FROM "courses"`).
			WithArgs(courseID, 1).
			WillReturnRows(rows)

		// --- mock query untuk preload course_images ---
		imageRows := sqlmock.NewRows([]string{
			"id", "course_id", "image_url", "created_at", "updated_at",
		}).AddRow(
			uuid.New(),
			courseID,
			"https://example.com/img1.png",
			now,
			now,
		).AddRow(
			uuid.New(),
			courseID,
			"https://example.com/img2.png",
			now,
			now,
		)

		mock.ExpectQuery(`SELECT .* FROM "course_images"`).
			WithArgs(courseID).
			WillReturnRows(imageRows)

		course, err := repo.FindByIDWithImages(ctx, courseID)
		assert.NoError(t, err)
		assert.Equal(t, courseID, course.ID)
		assert.Len(t, course.CourseImages, 2)
	})

	t.Run("not found", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewCourseRepository(db)

		courseID := uuid.New()

		mock.ExpectQuery(`SELECT .* FROM "courses"`).
			WithArgs(courseID, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		course, err := repo.FindByIDWithImages(ctx, courseID)
		assert.Error(t, err)
		assert.Nil(t, course)
	})

	t.Run("db error", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewCourseRepository(db)

		courseID := uuid.New()

		mock.ExpectQuery(`SELECT .* FROM "courses"`).
			WithArgs(courseID, 1).
			WillReturnError(errors.New("db connection failed"))

		course, err := repo.FindByIDWithImages(ctx, courseID)
		assert.Error(t, err)
		assert.Nil(t, course)
	})
}

func TestIsSlugExists(t *testing.T) {
	ctx := context.Background()

	t.Run("slug exists", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewCourseRepository(db)

		slug := "golang-course"

		// mock query count
		rows := sqlmock.NewRows([]string{"count"}).AddRow(1)

		mock.ExpectQuery(`SELECT count\(\*\) FROM "courses"`).
			WithArgs(slug).
			WillReturnRows(rows)

		exists := repo.IsSlugExists(ctx, slug)
		assert.True(t, exists)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("slug does not exist", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewCourseRepository(db)

		slug := "non-existent"

		// count = 0
		rows := sqlmock.NewRows([]string{"count"}).AddRow(0)

		mock.ExpectQuery(`SELECT count\(\*\) FROM "courses"`).
			WithArgs(slug).
			WillReturnRows(rows)

		exists := repo.IsSlugExists(ctx, slug)
		assert.False(t, exists)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewCourseRepository(db)

		slug := "error-slug"

		mock.ExpectQuery(`SELECT count\(\*\) FROM "courses"`).
			WithArgs(slug).
			WillReturnError(errors.New("db error"))

		// karena ada error, count default = 0 → return false
		exists := repo.IsSlugExists(ctx, slug)
		assert.False(t, exists)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUpdateCourse(t *testing.T) {
	ctx := context.Background()

	t.Run("success - update course", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewCourseRepository(db)

		courseID := uuid.New()
		now := time.Now()

		course := &models.Course{
			ID:               courseID,
			Slug:             "golang-course",
			Title:            "Belajar Golang",
			ShortDescription: "short desc",
			Description:      "desc",
			LearningOutcomes: "outcomes",
			Achievements:     "achievements",
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		// GORM Save akan jadi UPDATE
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "courses" SET`).
			WithArgs(
				course.Slug,
				course.Title,
				course.ShortDescription,
				course.Description,
				course.LearningOutcomes,
				course.Achievements,
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
				course.ID,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(ctx, course)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewCourseRepository(db)

		course := &models.Course{
			ID:    uuid.New(),
			Slug:  "fail-course",
			Title: "Will Fail",
		}

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "courses" SET`).
			WithArgs(
				course.Slug,
				course.Title,
				course.ShortDescription,
				course.Description,
				course.LearningOutcomes,
				course.Achievements,
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
				course.ID,
			).
			WillReturnError(errors.New("update failed"))
		mock.ExpectRollback()

		err := repo.Update(ctx, course)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDeleteCourseImagesByCourseID(t *testing.T) {
	ctx := context.Background()

	t.Run("success - delete course images", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewCourseRepository(db)

		courseID := uuid.New()

		// GORM Delete akan jadi DELETE
		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "course_images" WHERE course_id =`).
			WithArgs(courseID).
			WillReturnResult(sqlmock.NewResult(1, 2)) // 2 rows affected
		mock.ExpectCommit()

		err := repo.DeleteCourseImagesByCourseID(ctx, courseID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewCourseRepository(db)

		courseID := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "course_images" WHERE course_id =`).
			WithArgs(courseID).
			WillReturnError(errors.New("delete failed"))
		mock.ExpectRollback()

		err := repo.DeleteCourseImagesByCourseID(ctx, courseID)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestFindByID(t *testing.T) {
	ctx := context.Background()

	t.Run("success - course", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewCourseRepository(db)

		courseID := uuid.New()
		now := time.Now()

		// --- mock query untuk course ---
		rows := sqlmock.NewRows([]string{
			"id", "slug", "title", "short_description", "description",
			"learning_outcomes", "achievements", "created_at", "updated_at",
		}).AddRow(
			courseID,
			"golang-course",
			"Belajar Golang",
			"short desc",
			"desc",
			"outcomes",
			"achievements",
			now,
			now,
		)

		mock.ExpectQuery(`SELECT .* FROM "courses"`).
			WithArgs(courseID, 1).
			WillReturnRows(rows)

		course, err := repo.FindByID(ctx, courseID)
		assert.NoError(t, err)
		assert.Equal(t, courseID, course.ID)
		assert.NoError(t, mock.ExpectationsWereMet())

	})

	t.Run("not found", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewCourseRepository(db)

		courseID := uuid.New()

		mock.ExpectQuery(`SELECT .* FROM "courses"`).
			WithArgs(courseID, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		course, err := repo.FindByID(ctx, courseID)
		assert.Error(t, err)
		assert.Nil(t, course)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewCourseRepository(db)

		courseID := uuid.New()

		mock.ExpectQuery(`SELECT .* FROM "courses"`).
			WithArgs(courseID, 1).
			WillReturnError(errors.New("db connection failed"))

		course, err := repo.FindByID(ctx, courseID)
		assert.Error(t, err)
		assert.Nil(t, course)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDeleteByID(t *testing.T) {
	ctx := context.Background()

	t.Run("success - delete course", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewCourseRepository(db)

		courseID := uuid.New()

		// GORM Delete akan jadi DELETE
		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "courses" WHERE id =`).
			WithArgs(courseID).
			WillReturnResult(sqlmock.NewResult(1, 2)) // 2 rows affected
		mock.ExpectCommit()

		err := repo.DeleteByID(ctx, courseID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		db, mock := setupMockDB(t)
		repo := repository.NewCourseRepository(db)

		courseID := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "courses" WHERE id =`).
			WithArgs(courseID).
			WillReturnError(errors.New("delete failed"))
		mock.ExpectRollback()

		err := repo.DeleteByID(ctx, courseID)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
