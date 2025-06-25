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

// BlogService provides methods for managing courses
type BlogService struct {
	repo        *repository.BlogRepository
	db          *gorm.DB
	fileService *FileService
}

// NewBlogService creates a new instance of BlogService
func NewBlogService(repo *repository.BlogRepository, db *gorm.DB, fileService *FileService) *BlogService {
	return &BlogService{repo: repo, db: db, fileService: fileService}
}

// GetAllFilteredBlogs retrieves all blogs with pagination and filtering options
func (s *BlogService) GetAllFilteredBlogs(opts utils.QueryOptions) ([]models.Blog, int64, error) {
	blogs, total, err := s.repo.GetAllFilteredBlogs(opts)
	if err != nil {
		return nil, 0, err
	}
	return blogs, total, nil
}

// GetBlogBySlug retrieves a blog by its slug
func (s *BlogService) GetBlogBySlug(slug string) (*models.Blog, error) {
	blog, err := s.repo.GetBlogBySlug(slug)
	if err != nil {
		return nil, err
	}
	return blog, nil
}

// CreateBlog creates a new blog with the provided details
func (s *BlogService) CreateBlog(body *dto.CreateBlogRequest) (*models.Blog, error) {
	tx := s.db.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	defer tx.Rollback()

	var blog models.Blog
	copier.Copy(&blog, body)

	slug := utils.GenerateUniqueSlug(body.Title, s.repo)

	blog.Slug = slug

	if err := s.repo.CreateTx(tx, &blog); err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &blog, nil
}

// UpdateBlog updates an existing blog with the provided details
func (s *BlogService) UpdateBlog(id uuid.UUID, body *dto.UpdateBlogRequest) (*models.Blog, error) {
	tx := s.db.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	defer tx.Rollback()

	blog, err := s.repo.FindByID(tx, id)
	if err != nil {
		return nil, err
	}

	oldImage := blog.Image

	// Copy field yang tidak nil saja
	if err := copier.CopyWithOption(blog, body, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateTx(tx, blog); err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	if body.Image != nil && *body.Image != "" && oldImage != *body.Image {
		err = s.fileService.DeleteFile(oldImage) // error optional, bisa di-log
		if err != nil {
			// Log error tapi lanjut
			log.Errorf("Gagal hapus file %s: %v", oldImage, err)
		}
	}

	return blog, nil

}

// DeleteBlog deletes a blog by its ID
func (s *BlogService) DeleteBlog(id uuid.UUID) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	defer tx.Rollback()

	// Optional: cek dulu apakah ada blog-nya
	blog, err := s.repo.FindByID(tx, id)
	if err != nil {
		return err
	}

	// Hapus blog (images akan ikut terhapus karena cascade)
	if err := s.repo.DeleteByIDTx(tx, id); err != nil {
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// Hapus file gambar
	err = s.fileService.DeleteFile(blog.Image)
	if err != nil {
		// Log error tapi lanjut
		log.Errorf("Gagal hapus file %s: %v", blog.Image, err)
	}

	return nil
}
