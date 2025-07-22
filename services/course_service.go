package services

import (
	"brevet-api/dto"
	"brevet-api/models"
	"brevet-api/repository"
	"brevet-api/utils"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// CourseService provides methods for managing courses
type CourseService struct {
	repo        *repository.CourseRepository
	db          *gorm.DB
	fileService *FileService
}

// NewCourseService creates a new instance of CourseService
func NewCourseService(repo *repository.CourseRepository, db *gorm.DB, fileService *FileService) *CourseService {
	return &CourseService{repo: repo, db: db, fileService: fileService}
}

// GetAllFilteredCourses retrieves all courses with pagination and filtering options
func (s *CourseService) GetAllFilteredCourses(opts utils.QueryOptions) ([]models.Course, int64, error) {
	courses, total, err := s.repo.GetAllFilteredCourses(opts)
	if err != nil {
		return nil, 0, err
	}
	return courses, total, nil
}

// GetCourseBySlug retrieves a course by its slug
func (s *CourseService) GetCourseBySlug(slug string) (*models.Course, error) {
	course, err := s.repo.GetCourseBySlug(slug)
	if err != nil {
		return nil, err
	}
	return course, nil
}

// CreateCourse creates a new course with the provided details
func (s *CourseService) CreateCourse(body *dto.CreateCourseRequest) (*models.Course, error) {
	var courseResponse models.Course
	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		course := &models.Course{
			Title:            body.Title,
			ShortDescription: body.ShortDescription,
			Description:      body.Description,
			LearningOutcomes: body.LearningOutcomes,
			Achievements:     body.Achievements,
		}

		slug := utils.GenerateUniqueSlug(body.Title, s.repo)

		course.Slug = slug

		if err := s.repo.WithTx(tx).Create(course); err != nil {
			return err
		}

		var images []models.CourseImage
		for _, input := range body.CourseImages {
			images = append(images, models.CourseImage{
				CourseID: course.ID,
				ImageURL: input.ImageURL,
			})
		}

		if err := s.repo.WithTx(tx).CreateCourseImagesBulk(images); err != nil {
			return err
		}

		courseWithImages, err := s.repo.WithTx(tx).FindByIDWithImages(course.ID)
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
func (s *CourseService) UpdateCourse(id uuid.UUID, body *dto.UpdateCourseRequest) (*models.Course, error) {
	var courseResponse models.Course
	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {

		course, err := s.repo.WithTx(tx).FindByID(id)
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

		if err := s.repo.WithTx(tx).Update(course); err != nil {
			return err
		}

		// Ganti course_images jika dikirim
		if body.CourseImages != nil {
			if err := s.repo.WithTx(tx).DeleteCourseImagesByCourseID(course.ID); err != nil {
				return err
			}

			var images []models.CourseImage
			for _, input := range *body.CourseImages {
				images = append(images, models.CourseImage{
					CourseID: course.ID,
					ImageURL: input.ImageURL,
				})
			}

			if err := s.repo.WithTx(tx).CreateCourseImagesBulk(images); err != nil {
				return err
			}
		}

		response, err := s.repo.FindByIDWithImages(course.ID)
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
func (s *CourseService) DeleteCourse(courseID uuid.UUID) error {
	var imagePaths []string

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		course, err := s.repo.WithTx(tx).FindByIDWithImages(courseID)
		if err != nil {
			return err
		}

		// Simpan path image
		for _, img := range course.CourseImages {
			imagePaths = append(imagePaths, img.ImageURL)
		}

		// Hapus course dari DB
		if err := s.repo.WithTx(tx).DeleteByID(courseID); err != nil {
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
