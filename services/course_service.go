package services

import (
	"brevet-api/dto"
	"brevet-api/models"
	"brevet-api/repository"
	"brevet-api/utils"
	"context"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// ICourseService interface
type ICourseService interface {
	GetAllFilteredCourses(ctx context.Context, opts utils.QueryOptions) ([]models.Course, int64, error)
	GetCourseBySlug(ctx context.Context, slug string) (*models.Course, error)
	CreateCourse(ctx context.Context, body *dto.CreateCourseRequest) (*models.Course, error)
	UpdateCourse(ctx context.Context, id uuid.UUID, body *dto.UpdateCourseRequest) (*models.Course, error)
	DeleteCourse(ctx context.Context, courseID uuid.UUID) error
}

// CourseService provides methods for managing courses
type CourseService struct {
	repo        repository.ICourseRepository
	db          *gorm.DB
	fileService IFileService
}

// NewCourseService creates a new instance of CourseService
func NewCourseService(repo repository.ICourseRepository, db *gorm.DB, fileService IFileService) ICourseService {
	return &CourseService{repo: repo, db: db, fileService: fileService}
}

// GetAllFilteredCourses retrieves all courses with pagination and filtering options
func (s *CourseService) GetAllFilteredCourses(ctx context.Context, opts utils.QueryOptions) ([]models.Course, int64, error) {
	courses, total, err := s.repo.GetAllFilteredCourses(ctx, opts)
	if err != nil {
		return nil, 0, err
	}
	return courses, total, nil
}

// GetCourseBySlug retrieves a course by its slug
func (s *CourseService) GetCourseBySlug(ctx context.Context, slug string) (*models.Course, error) {
	course, err := s.repo.GetCourseBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return course, nil
}

// CreateCourse creates a new course with the provided details
func (s *CourseService) CreateCourse(ctx context.Context, body *dto.CreateCourseRequest) (*models.Course, error) {
	var courseResponse models.Course
	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		course := &models.Course{
			Title:            body.Title,
			ShortDescription: body.ShortDescription,
			Description:      body.Description,
			LearningOutcomes: body.LearningOutcomes,
			Achievements:     body.Achievements,
		}

		slug := utils.GenerateUniqueSlug(ctx, body.Title, s.repo)

		course.Slug = slug

		if err := s.repo.WithTx(tx).Create(ctx, course); err != nil {
			return err
		}

		var images []models.CourseImage
		for _, input := range body.CourseImages {
			images = append(images, models.CourseImage{
				CourseID: course.ID,
				ImageURL: input.ImageURL,
			})
		}

		if err := s.repo.WithTx(tx).CreateCourseImagesBulk(ctx, images); err != nil {
			return err
		}

		courseWithImages, err := s.repo.WithTx(tx).FindByIDWithImages(ctx, course.ID)
		if err != nil {
			return err
		}
		courseResponse = utils.Safe(courseWithImages, models.Course{})
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &courseResponse, nil

}

// UpdateCourse is blabla
func (s *CourseService) UpdateCourse(ctx context.Context, id uuid.UUID, body *dto.UpdateCourseRequest) (*models.Course, error) {
	var courseResponse models.Course
	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {

		course, err := s.repo.WithTx(tx).FindByID(ctx, id)
		if err != nil {
			return err
		}

		// Copy field yang tidak nil saja
		if err := copier.CopyWithOption(course, body, copier.Option{
			IgnoreEmpty: true,
			DeepCopy:    true,
		}); err != nil {
			return err
		}

		// Optional: regenerate slug kalau Title berubah
		// if body.Title != nil {
		// 	slug := utils.GenerateUniqueSlug(*body.Title, s.repo)
		// 	course.Slug = slug
		// }

		if err := s.repo.WithTx(tx).Update(ctx, course); err != nil {
			return err
		}

		// Ganti course_images jika dikirim
		if body.CourseImages != nil {
			if err := s.repo.WithTx(tx).DeleteCourseImagesByCourseID(ctx, course.ID); err != nil {
				return err
			}

			var images []models.CourseImage
			for _, input := range *body.CourseImages {
				images = append(images, models.CourseImage{
					CourseID: course.ID,
					ImageURL: input.ImageURL,
				})
			}

			if err := s.repo.WithTx(tx).CreateCourseImagesBulk(ctx, images); err != nil {
				return err
			}
		}

		response, err := s.repo.FindByIDWithImages(ctx, course.ID)
		if err != nil {
			return err
		}

		courseResponse = utils.Safe(response, models.Course{})

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &courseResponse, nil
}

// DeleteCourse deletes a course by its ID
func (s *CourseService) DeleteCourse(ctx context.Context, courseID uuid.UUID) error {
	var imagePaths []string

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		course, err := s.repo.WithTx(tx).FindByIDWithImages(ctx, courseID)
		if err != nil {
			return err
		}

		// Simpan path image
		for _, img := range course.CourseImages {
			imagePaths = append(imagePaths, img.ImageURL)
		}

		// Hapus course dari DB
		if err := s.repo.WithTx(tx).DeleteByID(ctx, courseID); err != nil {
			return err
		}

		return nil
	})

	// ✅ Hapus file di luar transaction (hanya jika tx berhasil)
	if err != nil {
		return err
	}

	for _, path := range imagePaths {
		if delErr := s.fileService.DeleteFile(path); delErr != nil {
			log.Errorf("Gagal hapus file %s: %v", path, delErr)
		}
	}

	return nil

}
