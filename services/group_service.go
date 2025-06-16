package services

import (
	"brevet-api/models"
	"brevet-api/utils"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GroupService is a struct that represents a group service
type GroupService struct {
	db *gorm.DB
}

// NewGroupService creates a new group service
func NewGroupService(db *gorm.DB) *GroupService {
	return &GroupService{db: db}
}

// GetAllGroups gets all groups from the database
func (s *GroupService) GetAllGroups(opts utils.QueryOptions) ([]models.Group, int64, error) {
	validSortFields, err := utils.GetValidColumns(s.db, &models.Group{})

	if err != nil {
		return nil, 0, err
	}

	sort := opts.Sort
	if !validSortFields[sort] {

		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := s.db.Model(&models.Group{})

	joinConditions := map[string]string{} // atau bisa nil
	joinedRelations := map[string]bool{}

	db = utils.ApplyFiltersWithJoins(db, "groups", opts.Filters, validSortFields, joinConditions, joinedRelations)

	if opts.Search != "" {
		db = db.Where("name ILIKE ?", "%"+opts.Search+"%")
	}

	var total int64
	db.Count(&total)

	var groups []models.Group
	err = db.Order(fmt.Sprintf("%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&groups).Error

	return groups, total, err
}

// GetGroupByID retrieves a group by its ID
func (s *GroupService) GetGroupByID(id uuid.UUID) (*models.Group, error) {
	var group models.Group
	if err := s.db.First(&group, id).Error; err != nil {
		return nil, err
	}
	return &group, nil
}
