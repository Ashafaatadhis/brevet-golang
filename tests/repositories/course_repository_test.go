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
