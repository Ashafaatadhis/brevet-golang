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
	tx := s.db.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	defer tx.Rollback()

	course := &models.Course{
		Title:            body.Title,
		ShortDescription: body.ShortDescription,
		Description:      body.Description,
		LearningOutcomes: body.LearningOutcomes,
		Achievements:     body.Achievements,
	}

	slug := utils.GenerateUniqueSlug(body.Title, s.repo)

	course.Slug = slug

	if err := s.repo.CreateTx(tx, course); err != nil {
		return nil, err
	}

	var images []models.CourseImage
	for _, input := range body.CourseImages {
		images = append(images, models.CourseImage{
			CourseID: course.ID,
			ImageURL: input.ImageURL,
		})
	}

	if err := s.repo.CreateCourseImagesBulkTx(tx, images); err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	courseWithImages, err := s.repo.FindByIDWithImages(course.ID)
	if err != nil {
		return nil, err
	}

	return courseWithImages, nil
}

// UpdateCourse is blabla
func (s *CourseService) UpdateCourse(id uuid.UUID, body *dto.UpdateCourseRequest) (*models.Course, error) {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	defer tx.Rollback()

	course, err := s.repo.FindByID(tx, id)
	if err != nil {
		return nil, err
	}

	// Copy field yang tidak nil saja
	if err := copier.CopyWithOption(course, body, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}); err != nil {
		return nil, err
	}

	// Optional: regenerate slug kalau Title berubah
	// if body.Title != nil {
	// 	slug := utils.GenerateUniqueSlug(*body.Title, s.repo)
	// 	course.Slug = slug
	// }

	if err := s.repo.UpdateTx(tx, course); err != nil {
		return nil, err
	}

	// Ganti course_images jika dikirim
	if body.CourseImages != nil {
		if err := s.repo.DeleteCourseImagesByCourseID(tx, course.ID); err != nil {
			return nil, err
		}

		var images []models.CourseImage
		for _, input := range *body.CourseImages {
			images = append(images, models.CourseImage{
				CourseID: course.ID,
				ImageURL: input.ImageURL,
			})
		}

		if err := s.repo.CreateCourseImagesBulkTx(tx, images); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return s.repo.FindByIDWithImages(course.ID)
}

// DeleteCourse deletes a course by its ID
func (s *CourseService) DeleteCourse(courseID uuid.UUID) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	defer tx.Rollback()

	// Optional: cek dulu apakah ada course-nya
	course, err := s.repo.FindByIDWithImagesTx(tx, courseID)
	if err != nil {
		return err
	}

	// Hapus course (images akan ikut terhapus karena cascade)
	if err := s.repo.DeleteByIDTx(tx, courseID); err != nil {
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// Hapus file gambar
	for _, img := range course.CourseImages {
		err := s.fileService.DeleteFile(img.ImageURL)
		if err != nil {
			// Log error tapi lanjut
			log.Errorf("Gagal hapus file %s: %v", img.ImageURL, err)
		}
	}

	return nil
}
