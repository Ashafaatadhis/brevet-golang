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
	var blog models.Blog

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		copier.Copy(&blog, body)
		slug := utils.GenerateUniqueSlug(body.Title, s.repo)
		blog.Slug = slug

		if err := s.repo.WithTx(tx).Create(&blog); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &blog, nil

}

// UpdateBlog updates an existing blog with the provided details
func (s *BlogService) UpdateBlog(id uuid.UUID, body *dto.UpdateBlogRequest) (*models.Blog, error) {
	var blog models.Blog
	var oldImage string
	var shouldDelete bool

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		blogPtr, err := s.repo.WithTx(tx).FindByID(id)
		if err != nil {
			return err
		}

		blog = utils.Safe(blogPtr, models.Blog{})
		oldImage = blog.Image

		if err := copier.CopyWithOption(&blog, body, copier.Option{
			IgnoreEmpty: true,
			DeepCopy:    true,
		}); err != nil {
			return err
		}

		if err := s.repo.WithTx(tx).Update(&blog); err != nil {
			return err
		}

		// ✅ Tandai kalau gambar berubah, tapi jangan hapus dulu
		if oldImage != "" && oldImage != blog.Image {
			shouldDelete = true
		}

		return nil
	})

	// Setelah TX berhasil hapus file
	if err == nil && shouldDelete {
		if delErr := s.fileService.DeleteFile(oldImage); delErr != nil {
			log.Errorf("Gagal hapus file %s: %v", oldImage, delErr)
		}
	}

	if err != nil {
		return nil, err
	}
	return &blog, nil
}

// DeleteBlog deletes a blog by its ID
func (s *BlogService) DeleteBlog(id uuid.UUID) error {
	var blog models.Blog
	var shouldDelete bool
	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		var err error
		blogPtr, err := s.repo.WithTx(tx).FindByID(id)
		if err != nil {
			return err
		}

		blog = utils.Safe(blogPtr, models.Blog{})

		// Hapus blog (images akan ikut terhapus karena cascade)
		if err := s.repo.WithTx(tx).DeleteByID(id); err != nil {
			return err
		}

		if blog.Image != "" {
			shouldDelete = true
		}

		return nil
	})

	if err == nil && shouldDelete {
		// Hapus file gambar
		if delErr := s.fileService.DeleteFile(blog.Image); delErr != nil {
			log.Errorf("Gagal hapus file %s: %v", blog.Image, delErr)
		}
	}

	if err != nil {
		return err
	}
	return nil

}
