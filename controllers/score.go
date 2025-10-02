package controllers

import (
	"brevet-api/services"
	"brevet-api/utils"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ScoreController is struct
type ScoreController struct {
	scoreService services.IScoreService
	db           *gorm.DB
}

// NewScoreController creates a new instance of ScoreController
func NewScoreController(scoreService services.IScoreService, db *gorm.DB) *ScoreController {
	return &ScoreController{
		scoreService: scoreService,
		db:           db,
	}
}

// GetScores retrieves assignment & quiz scores of authenticated student in a batch
func (ctrl *ScoreController) GetScores(c *fiber.Ctx) error {
	ctx := c.UserContext()

	batchIDParam := c.Params("batchID")
	if batchIDParam == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "batchID is required", "")
	}

	batchID, err := uuid.Parse(batchIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid batchID", err.Error())
	}

	// Ambil user dari context
	userClaims := c.Locals("user").(*utils.Claims)

	// Panggil service
	scores, err := ctrl.scoreService.GetScoresByBatchUser(ctx, batchID, userClaims)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Scores not found", "")
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve scores", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Scores retrieved successfully", scores)
}

// GetStudentScores retrieves scores of a student in a batch (admin/guru)
func (ctrl *ScoreController) GetStudentScores(c *fiber.Ctx) error {
	ctx := c.UserContext()

	batchSlug := c.Params("batchSlug")
	studentIDParam := c.Params("studentID")

	// Panggil service
	scores, err := ctrl.scoreService.GetScoresByBatchStudentSlug(ctx, batchSlug, studentIDParam)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Scores not found", "")
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve scores", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Scores retrieved successfully", scores)
}
