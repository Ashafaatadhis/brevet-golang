package controllers

import (
	"brevet-api/dto"
	"brevet-api/services"
	"brevet-api/utils"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// QuizController is struct
type QuizController struct {
	quizService services.IQuizService
	db          *gorm.DB
}

// NewQuizController creates a new instance of QuizController
func NewQuizController(quizService services.IQuizService, db *gorm.DB) *QuizController {
	return &QuizController{
		quizService: quizService,
		db:          db,
	}
}

// SaveTempSubmission saves a temporary submission for a quiz
func (ctrl *QuizController) SaveTempSubmission(c *fiber.Ctx) error {
	ctx := c.UserContext()
	user := c.Locals("user").(*utils.Claims)
	body := c.Locals("body").(*dto.SaveTempSubmissionRequest)
	attemptID, err := uuid.Parse(c.Params("attemptID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid attempt ID", err.Error())
	}

	if err := ctrl.quizService.SaveTempSubmission(ctx, user, attemptID, body); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed save temp submission", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Questions saved successfully", fiber.Map{"status": "saved"})

}

// ImportQuestionsFromExcel excel
func (ctrl *QuizController) ImportQuestionsFromExcel(c *fiber.Ctx) error {
	ctx := c.UserContext()
	user := c.Locals("user").(*utils.Claims)

	quizID, err := uuid.Parse(c.Params("quizID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid quiz ID", err.Error())
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Missing excel file", err.Error())
	}

	if err := ctrl.quizService.ImportQuestionsFromExcel(ctx, user, quizID, fileHeader); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to import questions", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Questions imported successfully", nil)
}

// GetQuizByMeetingIDFiltered retrieves a list of purchases with pagination and filtering options
func (ctrl *QuizController) GetQuizByMeetingIDFiltered(c *fiber.Ctx) error {
	ctx := c.UserContext()
	opts := utils.ParseQueryOptions(c)
	user := c.Locals("user").(*utils.Claims)

	meetingID, err := uuid.Parse(c.Params("meetingID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid meeting ID", err.Error())
	}

	quizzes, total, err := ctrl.quizService.GetQuizByMeetingIDFiltered(ctx, meetingID, opts, user)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch quizzes", err.Error())
	}

	var quizzesResponse []dto.QuizResponse
	if copyErr := copier.Copy(&quizzesResponse, quizzes); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map quiz data", copyErr.Error())
	}

	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)
	return utils.SuccessWithMeta(c, fiber.StatusOK, "Quizzes fetched", quizzesResponse, meta)
}

// StartQuiz start
func (ctrl *QuizController) StartQuiz(c *fiber.Ctx) error {
	ctx := c.UserContext()
	user := c.Locals("user").(*utils.Claims)

	quizID, err := uuid.Parse(c.Params("quizID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid quiz ID", err.Error())
	}

	attempt, err := ctrl.quizService.StartQuiz(ctx, user, quizID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed start quiz", err.Error())
	}

	var quizAttemptResponse dto.QuizAttemptResponse
	if err := copier.CopyWithOption(&quizAttemptResponse, attempt, copier.Option{IgnoreEmpty: true}); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to map quiz attempt data", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Success Start Quiz", quizAttemptResponse)

}

// CreateQuizMetadata route post
func (ctrl *QuizController) CreateQuizMetadata(c *fiber.Ctx) error {
	ctx := c.UserContext()
	user := c.Locals("user").(*utils.Claims)
	body := c.Locals("body").(*dto.ImportQuizzesRequest)

	meetingID, err := uuid.Parse(c.Params("meetingID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid meeting ID", err.Error())
	}

	quiz, err := ctrl.quizService.CreateQuizMetadata(ctx, user, meetingID, body)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create quiz metadata", err.Error())
	}

	var quizResponse dto.QuizResponse
	if err := copier.Copy(&quizResponse, quiz); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to map quiz data", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Quiz metadata created", quizResponse)
}

// SubmitQuiz controller
func (ctrl *QuizController) SubmitQuiz(c *fiber.Ctx) error {
	ctx := c.UserContext()
	user := c.Locals("user").(*utils.Claims)

	attemptID, err := uuid.Parse(c.Params("attemptID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid quiz ID", err.Error())
	}

	if err := ctrl.quizService.SubmitQuiz(ctx, user, attemptID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "failed to submit quiz", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Quiz submitted successfully", nil)
}

// GetQuizByID for get quiz by id
func (ctrl *QuizController) GetQuizByID(c *fiber.Ctx) error {
	ctx := c.UserContext()
	user := c.Locals("user").(*utils.Claims)
	quizID, err := uuid.Parse(c.Params("quizID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid quiz ID", err.Error())
	}

	quiz, err := ctrl.quizService.GetQuizMetadata(ctx, user, quizID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "quiz not found", err.Error())
	}

	var quizResponse dto.QuizResponse
	if err := copier.Copy(&quizResponse, quiz); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to map quiz data", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Get Quiz metadata", quizResponse)
}

// GetQuizWithQuestions controller
func (ctrl *QuizController) GetQuizWithQuestions(c *fiber.Ctx) error {
	user := c.Locals("user").(*utils.Claims)

	quizID, err := uuid.Parse(c.Params("quizID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid quiz ID", err.Error())
	}

	quiz, err := ctrl.quizService.GetQuizWithQuestions(c.UserContext(), user, quizID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "cannot access quiz", err.Error())
	}
	var quizResponse dto.QuizResponse
	if err := copier.CopyWithOption(&quizResponse, quiz, copier.Option{IgnoreEmpty: true}); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to map quiz data", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Success Get Questions", quizResponse)
}

// GetActiveAttempt controller
func (ctrl *QuizController) GetActiveAttempt(c *fiber.Ctx) error {
	ctx := c.UserContext()
	user := c.Locals("user").(*utils.Claims)

	quizID, err := uuid.Parse(c.Params("quizID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid quiz ID", err.Error())
	}

	attempt, err := ctrl.quizService.GetActiveAttempt(ctx, quizID, user)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch attempt", err.Error())
	}

	if attempt == nil {
		return utils.SuccessResponse(c, fiber.StatusOK, "No attempt yet", nil)
	}

	var quizAttemptResponse dto.QuizAttemptResponse
	if err := copier.CopyWithOption(&quizAttemptResponse, attempt, copier.Option{IgnoreEmpty: true}); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to map quiz attempt data", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Success", quizAttemptResponse)
}

// GetAttemptDetail detail
func (ctrl *QuizController) GetAttemptDetail(c *fiber.Ctx) error {
	ctx := c.UserContext()
	user := c.Locals("user").(*utils.Claims)

	attemptID, err := uuid.Parse(c.Params("attemptID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid attempt ID", err.Error())
	}

	attemptDetail, err := ctrl.quizService.GetAttemptDetail(ctx, attemptID, user)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch attempt", err.Error())
	}

	fmt.Println(attemptDetail)

	var attemptDetailResponse dto.QuizAttemptFullResponse
	if err := copier.CopyWithOption(&attemptDetailResponse, attemptDetail, copier.Option{IgnoreEmpty: true}); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to map quiz attempt data", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Success", attemptDetailResponse)
}

// UpdateQuiz update
func (ctrl *QuizController) UpdateQuiz(c *fiber.Ctx) error {
	ctx := c.UserContext()
	body := c.Locals("body").(*dto.UpdateQuizRequest)
	quizID, err := uuid.Parse(c.Params("quizID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid quiz ID", err.Error())
	}

	user := c.Locals("user").(*utils.Claims)
	quiz, err := ctrl.quizService.UpdateQuiz(ctx, quizID, user, body)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update quiz", err.Error())
	}

	var quizResponse dto.QuizResponse
	if err := copier.Copy(&quizResponse, quiz); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to map quiz data", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Quiz updated", quizResponse)
}

// DeleteQuiz deletes a quiz by ID
func (ctrl *QuizController) DeleteQuiz(c *fiber.Ctx) error {
	ctx := c.UserContext()
	quizID, err := uuid.Parse(c.Params("quizID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid quiz ID", err.Error())
	}

	user := c.Locals("user").(*utils.Claims)
	if err := ctrl.quizService.DeleteQuiz(ctx, quizID, user); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete quiz", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Quiz deleted", nil)
}

// GetAttemptResult result
func (ctrl *QuizController) GetAttemptResult(c *fiber.Ctx) error {
	ctx := c.UserContext()
	user := c.Locals("user").(*utils.Claims)

	attemptID, err := uuid.Parse(c.Params("attemptID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid attempt ID", err.Error())
	}

	result, err := ctrl.quizService.GetAttemptResult(ctx, attemptID, user)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch result", err.Error())
	}

	var quizResultResponse dto.QuizResultResponse
	if err := copier.Copy(&quizResultResponse, result); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to map quiz data", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Success", quizResultResponse)
}
