package services

import (
	"brevet-api/dto"
	"brevet-api/repository"
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IDashboardService interface
type IDashboardService interface {
	GetAdminDashboard(ctx context.Context, period string) (*dto.DashboardAdminResponse, error)
	GetRevenueChart(ctx context.Context, period string) (*dto.RevenueChartResponse, error)
	GetPendingPayments(ctx context.Context, limit int) (*dto.PendingPaymentsResponse, error)
	GetBatchProgress(ctx context.Context, limit int) (*dto.BatchProgressResponse, error)
	GetTeacherWorkload(ctx context.Context, period string) (*dto.TeacherWorkloadResponse, error)
	GetCertificateStats(ctx context.Context, period string) (*dto.CertificateStatsResponse, error)
	GetRecentActivities(ctx context.Context, period string, limit int) (*dto.RecentActivitiesResponse, error)
	GetTeacherDashboard(ctx context.Context, teacherID uuid.UUID) (*dto.TeacherDashboardResponse, error)
	GetStudentScoreProgress(ctx context.Context, batchSlug string, studentID uuid.UUID) (*dto.StudentScoreProgressResponse, error)
	GetStudentDashboard(ctx context.Context, studentID uuid.UUID) (*dto.StudentDashboardResponse, error)
}

// DashboardService provides methods for dashboard statistics
type DashboardService struct {
	purchaseRepo    repository.IPurchaseRepository
	batchRepo       repository.IBatchRepository
	certificateRepo repository.ICertificateRepository
	db              *gorm.DB
}

// NewDashboardService creates a new instance of DashboardService
func NewDashboardService(
	purchaseRepo repository.IPurchaseRepository,
	batchRepo repository.IBatchRepository,
	certificateRepo repository.ICertificateRepository,
	db *gorm.DB,
) IDashboardService {
	return &DashboardService{
		purchaseRepo:    purchaseRepo,
		batchRepo:       batchRepo,
		certificateRepo: certificateRepo,
		db:              db,
	}
}

// GetTeacherDashboard returns dashboard statistics for a teacher
func (s *DashboardService) GetTeacherDashboard(ctx context.Context, teacherID uuid.UUID) (*dto.TeacherDashboardResponse, error) {
	now := time.Now()

	// Total courses (distinct batches taught by this teacher)
	var totalCourses int64
	if err := s.db.WithContext(ctx).
		Table("batches").
		Select("COUNT(DISTINCT batches.id)").
		Joins("JOIN meetings ON meetings.batch_id = batches.id").
		Joins("JOIN meeting_teachers mt ON mt.meeting_id = meetings.id").
		Where("mt.user_id = ?", teacherID).
		Scan(&totalCourses).Error; err != nil {
		return nil, fmt.Errorf("failed to count total courses: %w", err)
	}

	// Active courses (batches not yet ended)
	var activeCourses int64
	if err := s.db.WithContext(ctx).
		Table("batches").
		Select("COUNT(DISTINCT batches.id)").
		Joins("JOIN meetings ON meetings.batch_id = batches.id").
		Joins("JOIN meeting_teachers mt ON mt.meeting_id = meetings.id").
		Where("mt.user_id = ? AND batches.end_at >= ?", teacherID, now).
		Scan(&activeCourses).Error; err != nil {
		return nil, fmt.Errorf("failed to count active courses: %w", err)
	}

	// Active students (distinct paid users in teacher's batches)
	var activeStudents int64
	if err := s.db.WithContext(ctx).
		Table("purchases").
		Select("COUNT(DISTINCT purchases.user_id)").
		Joins("JOIN batches ON batches.id = purchases.batch_id").
		Joins("JOIN meetings ON meetings.batch_id = batches.id").
		Joins("JOIN meeting_teachers mt ON mt.meeting_id = meetings.id").
		Where("mt.user_id = ? AND purchases.payment_status = ?", teacherID, "paid").
		Scan(&activeStudents).Error; err != nil {
		return nil, fmt.Errorf("failed to count active students: %w", err)
	}

	// Ongoing tasks: assignments and quizzes that have not ended yet
	var ongoingAssignments int64
	if err := s.db.WithContext(ctx).
		Table("assignments").
		Joins("JOIN meetings ON meetings.id = assignments.meeting_id").
		Joins("JOIN meeting_teachers mt ON mt.meeting_id = meetings.id").
		Where("mt.user_id = ? AND assignments.end_at >= ?", teacherID, now).
		Count(&ongoingAssignments).Error; err != nil {
		return nil, fmt.Errorf("failed to count ongoing assignments: %w", err)
	}

	var ongoingQuizzes int64
	if err := s.db.WithContext(ctx).
		Table("quizzes").
		Joins("JOIN meetings ON meetings.id = quizzes.meeting_id").
		Joins("JOIN meeting_teachers mt ON mt.meeting_id = meetings.id").
		Where("mt.user_id = ? AND quizzes.end_time >= ?", teacherID, now).
		Count(&ongoingQuizzes).Error; err != nil {
		return nil, fmt.Errorf("failed to count ongoing quizzes: %w", err)
	}

	ongoingTasks := ongoingAssignments + ongoingQuizzes

	// Pending grading: assignment submissions without grades in teacher's batches
	var pendingGrading int64
	if err := s.db.WithContext(ctx).
		Table("assignment_submissions").
		Joins("JOIN assignments ON assignments.id = assignment_submissions.assignment_id").
		Joins("JOIN meetings ON meetings.id = assignments.meeting_id").
		Joins("JOIN meeting_teachers mt ON mt.meeting_id = meetings.id").
		Where("mt.user_id = ? AND assignment_submissions.id NOT IN (SELECT assignment_submission_id FROM assignment_grades)", teacherID).
		Count(&pendingGrading).Error; err != nil {
		return nil, fmt.Errorf("failed to count pending grading: %w", err)
	}

	// Upcoming schedules: next two meetings taught by the teacher
	type UpcomingMeeting struct {
		ID          string
		Title       string
		StartAt     time.Time
		EndAt       time.Time
		BatchSlug   string
		BatchTitle  string
		CourseTitle *string
	}

	var upcoming []UpcomingMeeting
	if err := s.db.WithContext(ctx).
		Table("meetings").
		Select(`
			meetings.id,
			meetings.title,
			meetings.start_at,
			meetings.end_at,
			batches.slug as batch_slug,
			batches.title as batch_title,
			courses.title as course_title
		`).
		Joins("JOIN meeting_teachers mt ON mt.meeting_id = meetings.id").
		Joins("JOIN batches ON batches.id = meetings.batch_id").
		Joins("LEFT JOIN courses ON courses.id = batches.course_id").
		Where("mt.user_id = ? AND meetings.start_at > ?", teacherID, now).
		Order("meetings.start_at ASC").
		Scan(&upcoming).Error; err != nil {
		return nil, fmt.Errorf("failed to get upcoming schedules: %w", err)
	}

	var upcomingItems []dto.UpcomingScheduleItem
	for _, m := range upcoming {
		item := dto.UpcomingScheduleItem{
			MeetingID:    m.ID,
			MeetingTitle: m.Title,
			BatchSlug:    m.BatchSlug,
			BatchTitle:   m.BatchTitle,
			StartAt:      m.StartAt,
			EndAt:        m.EndAt,
		}
		if m.CourseTitle != nil {
			item.CourseTitle = *m.CourseTitle
		}
		upcomingItems = append(upcomingItems, item)
	}

	if upcomingItems == nil {
		upcomingItems = []dto.UpcomingScheduleItem{}
	}

	return &dto.TeacherDashboardResponse{
		TotalCourses:      totalCourses,
		ActiveCourses:     activeCourses,
		ActiveStudents:    activeStudents,
		OngoingTasks:      ongoingTasks,
		PendingGrading:    pendingGrading,
		TotalUpcoming:     len(upcomingItems),
		UpcomingSchedules: upcomingItems,
	}, nil
}

// GetStudentDashboard returns dashboard metrics for a student
func (s *DashboardService) GetStudentDashboard(ctx context.Context, studentID uuid.UUID) (*dto.StudentDashboardResponse, error) {
	now := time.Now()
	sevenDays := now.AddDate(0, 0, 7)

	// Total courses (paid purchases)
	var totalCourses int64
	if err := s.db.WithContext(ctx).
		Table("purchases").
		Select("COUNT(DISTINCT batch_id)").
		Where("user_id = ? AND payment_status = ?", studentID, "paid").
		Scan(&totalCourses).Error; err != nil {
		return nil, fmt.Errorf("failed to count total courses: %w", err)
	}

	// Active courses (paid and batch not ended)
	var activeCourses int64
	if err := s.db.WithContext(ctx).
		Table("purchases").
		Select("COUNT(DISTINCT purchases.batch_id)").
		Joins("JOIN batches ON batches.id = purchases.batch_id").
		Where("purchases.user_id = ? AND purchases.payment_status = ? AND batches.end_at >= ?", studentID, "paid", now).
		Scan(&activeCourses).Error; err != nil {
		return nil, fmt.Errorf("failed to count active courses: %w", err)
	}

	// Batch IDs the student has paid
	var batchIDs []string
	if err := s.db.WithContext(ctx).
		Table("purchases").
		Select("DISTINCT batch_id").
		Where("user_id = ? AND payment_status = ?", studentID, "paid").
		Pluck("batch_id", &batchIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to get batch list: %w", err)
	}

	// Average progress across batches
	var totalProgress float64
	for _, batchID := range batchIDs {
		progress, err := s.calculateStudentProgress(ctx, batchID, studentID)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate progress: %w", err)
		}
		totalProgress += progress
	}
	var avgProgress float64
	if len(batchIDs) > 0 {
		avgProgress = totalProgress / float64(len(batchIDs))
	}

	// Certificates count
	var certificateCount int64
	if err := s.db.WithContext(ctx).
		Table("certificates").
		Where("user_id = ?", studentID).
		Count(&certificateCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count certificates: %w", err)
	}

	// Upcoming tasks/quizzes in next 7 days for student's batches
	var upcomingAssignments int64
	if len(batchIDs) > 0 {
		if err := s.db.WithContext(ctx).
			Table("assignments").
			Joins("JOIN meetings ON meetings.id = assignments.meeting_id").
			Where("meetings.batch_id IN ? AND assignments.end_at BETWEEN ? AND ?", batchIDs, now, sevenDays).
			Count(&upcomingAssignments).Error; err != nil {
			return nil, fmt.Errorf("failed to count upcoming assignments: %w", err)
		}
	}

	var upcomingQuizzes int64
	if len(batchIDs) > 0 {
		if err := s.db.WithContext(ctx).
			Table("quizzes").
			Joins("JOIN meetings ON meetings.id = quizzes.meeting_id").
			Where("meetings.batch_id IN ? AND quizzes.start_time BETWEEN ? AND ?", batchIDs, now, sevenDays).
			Count(&upcomingQuizzes).Error; err != nil {
			return nil, fmt.Errorf("failed to count upcoming quizzes: %w", err)
		}
	}

	return &dto.StudentDashboardResponse{
		TotalCourses:      totalCourses,
		ActiveCourses:     activeCourses,
		AvgProgress:       avgProgress,
		TotalCertificates: certificateCount,
		UpcomingTasks:     upcomingAssignments + upcomingQuizzes,
	}, nil
}

// GetStudentScoreProgress returns score progress per meeting for a student in a batch
func (s *DashboardService) GetStudentScoreProgress(ctx context.Context, batchSlug string, studentID uuid.UUID) (*dto.StudentScoreProgressResponse, error) {
	// Get batch id and title
	var batch struct {
		ID    string
		Title string
	}

	if err := s.db.WithContext(ctx).
		Table("batches").
		Select("id, title").
		Where("slug = ?", batchSlug).
		Scan(&batch).Error; err != nil {
		return nil, fmt.Errorf("failed to get batch: %w", err)
	}

	if batch.ID == "" {
		return nil, fmt.Errorf("batch not found")
	}

	// Get meetings for the batch ordered by start time
	var meetings []struct {
		ID      string
		Title   string
		StartAt time.Time
	}

	if err := s.db.WithContext(ctx).
		Table("meetings").
		Select("id, title, start_at").
		Where("batch_id = ?", batch.ID).
		Order("start_at ASC").
		Scan(&meetings).Error; err != nil {
		return nil, fmt.Errorf("failed to get meetings: %w", err)
	}

	var items []dto.StudentMeetingScoreItem
	for _, m := range meetings {
		var assignmentAvg sql.NullFloat64
		if err := s.db.WithContext(ctx).
			Table("assignment_grades").
			Select("AVG(assignment_grades.grade)").
			Joins("JOIN assignment_submissions ON assignment_submissions.id = assignment_grades.assignment_submission_id").
			Joins("JOIN assignments ON assignments.id = assignment_submissions.assignment_id").
			Where("assignments.meeting_id = ? AND assignment_submissions.user_id = ?", m.ID, studentID).
			Scan(&assignmentAvg).Error; err != nil {
			return nil, fmt.Errorf("failed to calculate assignment average: %w", err)
		}

		var quizAvg sql.NullFloat64
		if err := s.db.WithContext(ctx).
			Table("quiz_results").
			Select("AVG(quiz_results.score_percent)").
			Joins("JOIN quiz_attempts ON quiz_attempts.id = quiz_results.attempt_id").
			Joins("JOIN quizzes ON quizzes.id = quiz_attempts.quiz_id").
			Where("quizzes.meeting_id = ? AND quiz_attempts.user_id = ?", m.ID, studentID).
			Scan(&quizAvg).Error; err != nil {
			return nil, fmt.Errorf("failed to calculate quiz average: %w", err)
		}

		var assignmentPtr *float64
		if assignmentAvg.Valid {
			assignmentPtr = &assignmentAvg.Float64
		}

		var quizPtr *float64
		if quizAvg.Valid {
			quizPtr = &quizAvg.Float64
		}

		items = append(items, dto.StudentMeetingScoreItem{
			MeetingID:     m.ID,
			MeetingTitle:  m.Title,
			MeetingDate:   m.StartAt,
			AssignmentAvg: assignmentPtr,
			QuizAvg:       quizPtr,
		})
	}

	if items == nil {
		items = []dto.StudentMeetingScoreItem{}
	}

	return &dto.StudentScoreProgressResponse{
		BatchSlug:  batchSlug,
		BatchTitle: batch.Title,
		Data:       items,
	}, nil
}

// calculateStudentProgress calculates progress for a student in a batch
// Progress = (completed assignments + quizzes + attendances) / total items
func (s *DashboardService) calculateStudentProgress(ctx context.Context, batchID string, studentID uuid.UUID) (float64, error) {
	// Count total items
	var totalAssignments, totalQuizzes, totalMeetings int64
	s.db.WithContext(ctx).
		Table("assignments").
		Joins("JOIN meetings ON meetings.id = assignments.meeting_id").
		Where("meetings.batch_id = ?", batchID).
		Count(&totalAssignments)

	s.db.WithContext(ctx).
		Table("quizzes").
		Joins("JOIN meetings ON meetings.id = quizzes.meeting_id").
		Where("meetings.batch_id = ?", batchID).
		Count(&totalQuizzes)

	s.db.WithContext(ctx).
		Table("meetings").
		Where("batch_id = ?", batchID).
		Count(&totalMeetings)

	totalItems := totalAssignments + totalQuizzes + totalMeetings
	if totalItems == 0 {
		return 0, nil
	}

	// Count completed items
	var completedAssignments, completedQuizzes, completedAttendances int64
	s.db.WithContext(ctx).
		Table("assignment_submissions").
		Joins("JOIN assignments ON assignments.id = assignment_submissions.assignment_id").
		Joins("JOIN meetings ON meetings.id = assignments.meeting_id").
		Where("meetings.batch_id = ? AND assignment_submissions.user_id = ?", batchID, studentID).
		Count(&completedAssignments)

	s.db.WithContext(ctx).
		Table("quiz_results").
		Joins("JOIN quiz_attempts ON quiz_attempts.id = quiz_results.attempt_id").
		Joins("JOIN quizzes ON quizzes.id = quiz_attempts.quiz_id").
		Joins("JOIN meetings ON meetings.id = quizzes.meeting_id").
		Where("meetings.batch_id = ? AND quiz_attempts.user_id = ?", batchID, studentID).
		Count(&completedQuizzes)

	s.db.WithContext(ctx).
		Table("attendances").
		Joins("JOIN meetings ON meetings.id = attendances.meeting_id").
		Where("meetings.batch_id = ? AND attendances.user_id = ? AND attendances.is_present = ?", batchID, studentID, true).
		Count(&completedAttendances)

	completedItems := completedAssignments + completedQuizzes + completedAttendances
	progress := (float64(completedItems) / float64(totalItems)) * 100
	return progress, nil
}

// GetAdminDashboard returns admin dashboard statistics based on period
func (s *DashboardService) GetAdminDashboard(ctx context.Context, period string) (*dto.DashboardAdminResponse, error) {
	// Tentukan tanggal mulai berdasarkan period
	var startDate time.Time
	now := time.Now()

	switch period {
	case "7d":
		startDate = now.AddDate(0, 0, -7)
	case "30d":
		startDate = now.AddDate(0, 0, -30)
	case "90d":
		startDate = now.AddDate(0, 0, -90)
	default:
		return nil, fmt.Errorf("invalid period: must be 7d, 30d, or 90d")
	}

	var response dto.DashboardAdminResponse
	response.Period = period

	// 1. Total Pendapatan (dari purchase yang paid dalam periode)
	var totalRevenue float64
	err := s.db.WithContext(ctx).
		Table("purchases").
		Select("COALESCE(SUM(prices.price), 0)").
		Joins("JOIN prices ON prices.id = purchases.price_id").
		Where("purchases.payment_status = ? AND purchases.created_at >= ?", "paid", startDate).
		Scan(&totalRevenue).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate total revenue: %w", err)
	}
	response.TotalRevenue = int64(totalRevenue)

	// 2. Peserta Aktif (user dengan purchase paid, distinct user)
	var activeParticipants int64
	err = s.db.WithContext(ctx).
		Table("purchases").
		Select("COUNT(DISTINCT user_id)").
		Where("payment_status = ?", "paid").
		Scan(&activeParticipants).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count active participants: %w", err)
	}
	response.ActiveParticipants = activeParticipants

	// 3. Batch Aktif (batch yang masih berlangsung atau belum selesai)
	var activeBatches int64
	err = s.db.WithContext(ctx).
		Table("batches").
		Where("end_at >= ?", now).
		Count(&activeBatches).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count active batches: %w", err)
	}
	response.ActiveBatches = activeBatches

	// 4. Pembelian Baru dalam periode
	var newPurchases int64
	err = s.db.WithContext(ctx).
		Table("purchases").
		Where("created_at >= ?", startDate).
		Count(&newPurchases).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count new purchases: %w", err)
	}
	response.NewPurchases = newPurchases

	// 5. Total Sertifikat yang diterbitkan dalam periode
	var totalCertificates int64
	err = s.db.WithContext(ctx).
		Table("certificates").
		Where("created_at >= ?", startDate).
		Count(&totalCertificates).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count certificates: %w", err)
	}
	response.TotalCertificates = totalCertificates

	// 6. Tingkat Penyelesaian (completion rate)
	// Hitung dari jumlah certificate / total active participants
	if activeParticipants > 0 {
		// Hitung total certificates yang pernah diterbitkan (all time)
		var allTimeCertificates int64
		err = s.db.WithContext(ctx).
			Table("certificates").
			Count(&allTimeCertificates).Error
		if err != nil {
			return nil, fmt.Errorf("failed to count all certificates: %w", err)
		}
		response.CompletionRate = (float64(allTimeCertificates) / float64(activeParticipants)) * 100
	} else {
		response.CompletionRate = 0
	}

	return &response, nil
}

// GetRevenueChart returns revenue chart data per day based on period
func (s *DashboardService) GetRevenueChart(ctx context.Context, period string) (*dto.RevenueChartResponse, error) {
	// Tentukan tanggal mulai berdasarkan period
	var startDate time.Time
	now := time.Now()

	switch period {
	case "7d":
		startDate = now.AddDate(0, 0, -7)
	case "30d":
		startDate = now.AddDate(0, 0, -30)
	case "90d":
		startDate = now.AddDate(0, 0, -90)
	default:
		return nil, fmt.Errorf("invalid period: must be 7d, 30d, or 90d")
	}

	// Query untuk mendapatkan revenue per hari
	type DailyRevenue struct {
		Date    string
		Revenue float64
	}

	var dailyRevenues []DailyRevenue
	err := s.db.WithContext(ctx).
		Table("purchases").
		Select("DATE(purchases.created_at) as date, COALESCE(SUM(prices.price), 0) as revenue").
		Joins("JOIN prices ON prices.id = purchases.price_id").
		Where("purchases.payment_status = ? AND purchases.created_at >= ?", "paid", startDate).
		Group("DATE(purchases.created_at)").
		Order("date ASC").
		Scan(&dailyRevenues).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get revenue chart data: %w", err)
	}

	// Convert ke format response
	var dataPoints []dto.RevenueChartDataPoint
	for _, dr := range dailyRevenues {
		dataPoints = append(dataPoints, dto.RevenueChartDataPoint{
			Date:    dr.Date,
			Revenue: dr.Revenue,
		})
	}

	// Jika tidak ada data, kembalikan array kosong
	if dataPoints == nil {
		dataPoints = []dto.RevenueChartDataPoint{}
	}

	return &dto.RevenueChartResponse{
		Period: period,
		Data:   dataPoints,
	}, nil
}

// GetPendingPayments returns list of purchases that need verification
func (s *DashboardService) GetPendingPayments(ctx context.Context, limit int) (*dto.PendingPaymentsResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	type PurchaseData struct {
		PurchaseID        string
		InvoiceNumber     int
		UserName          string
		UserEmail         string
		BatchSlug         string
		BatchTitle        string
		Amount            float64
		PaymentStatus     string
		PaymentProof      *string
		TransferAmount    float64
		BankAccountName   *string
		BankAccountNumber *string
		CreatedAt         time.Time
	}

	var purchases []PurchaseData
	err := s.db.WithContext(ctx).
		Table("purchases").
		Select(`
			purchases.id as purchase_id,
			purchases.invoice_number,
			users.name as user_name,
			users.email as user_email,
			batches.slug as batch_slug,
			batches.title as batch_title,
			prices.price as amount,
			purchases.payment_status,
			purchases.payment_proof,
			purchases.transfer_amount,
			purchases.buyer_bank_account_name as bank_account_name,
			purchases.buyer_bank_account_number as bank_account_number,
			purchases.created_at
		`).
		Joins("LEFT JOIN users ON users.id = purchases.user_id").
		Joins("LEFT JOIN batches ON batches.id = purchases.batch_id").
		Joins("LEFT JOIN prices ON prices.id = purchases.price_id").
		Where("purchases.payment_status IN ?", []string{"pending"}).
		Order("purchases.created_at DESC").
		Limit(limit).
		Scan(&purchases).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get pending payments: %w", err)
	}

	var items []dto.PendingPaymentItem
	for _, p := range purchases {
		items = append(items, dto.PendingPaymentItem{
			PurchaseID:        p.PurchaseID,
			InvoiceNumber:     p.InvoiceNumber,
			UserName:          p.UserName,
			UserEmail:         p.UserEmail,
			BatchSlug:         p.BatchSlug,
			BatchTitle:        p.BatchTitle,
			Amount:            p.Amount,
			PaymentStatus:     p.PaymentStatus,
			PaymentProof:      p.PaymentProof,
			TransferAmount:    p.TransferAmount,
			BankAccountName:   p.BankAccountName,
			BankAccountNumber: p.BankAccountNumber,
			CreatedAt:         p.CreatedAt,
		})
	}

	if items == nil {
		items = []dto.PendingPaymentItem{}
	}

	return &dto.PendingPaymentsResponse{
		Total: len(items),
		Data:  items,
	}, nil
}

// GetBatchProgress returns list of batches with progress and next activities
func (s *DashboardService) GetBatchProgress(ctx context.Context, limit int) (*dto.BatchProgressResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	type BatchData struct {
		BatchSlug   string
		BatchTitle  string
		CourseTitle string
		Quota       int
		Enrolled    int64
	}

	var batches []BatchData
	err := s.db.WithContext(ctx).
		Table("batches").
		Select(`
			batches.slug as batch_slug,
			batches.title as batch_title,
			courses.title as course_title,
			batches.quota,
			COUNT(DISTINCT purchases.user_id) as enrolled
		`).
		Joins("LEFT JOIN courses ON courses.id = batches.course_id").
		Joins("LEFT JOIN purchases ON purchases.batch_id = batches.id AND purchases.payment_status = 'paid'").
		Where("batches.end_at >= ?", time.Now()).
		Group("batches.id, batches.slug, batches.title, courses.title, batches.quota").
		Order("enrolled DESC").
		Limit(limit).
		Scan(&batches).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get batch progress: %w", err)
	}

	var items []dto.BatchProgressItem
	for _, b := range batches {
		// Calculate average progress untuk batch ini
		avgProgress := s.calculateBatchAvgProgress(ctx, b.BatchSlug)

		// Get next activity
		nextActivityType, nextActivityTitle, nextActivityDate := s.getNextActivity(ctx, b.BatchSlug)

		items = append(items, dto.BatchProgressItem{
			BatchSlug:         b.BatchSlug,
			BatchTitle:        b.BatchTitle,
			CourseTitle:       b.CourseTitle,
			Quota:             b.Quota,
			Enrolled:          int(b.Enrolled),
			AvgProgress:       avgProgress,
			NextActivityType:  nextActivityType,
			NextActivityTitle: nextActivityTitle,
			NextActivityDate:  nextActivityDate,
		})
	}

	if items == nil {
		items = []dto.BatchProgressItem{}
	}

	return &dto.BatchProgressResponse{
		Total: len(items),
		Data:  items,
	}, nil
}

// calculateBatchAvgProgress calculates average progress of all students in a batch
func (s *DashboardService) calculateBatchAvgProgress(ctx context.Context, batchSlug string) float64 {
	// Get batch ID from slug
	var batchID string
	err := s.db.WithContext(ctx).
		Table("batches").
		Select("id").
		Where("slug = ?", batchSlug).
		Scan(&batchID).Error
	if err != nil {
		return 0
	}

	// Get all students in this batch
	var studentIDs []string
	err = s.db.WithContext(ctx).
		Table("purchases").
		Select("DISTINCT user_id").
		Where("batch_id = ? AND payment_status = ?", batchID, "paid").
		Pluck("user_id", &studentIDs).Error
	if err != nil || len(studentIDs) == 0 {
		return 0
	}

	// Calculate progress for each student and average them
	var totalProgress float64
	for _, studentID := range studentIDs {
		// Get student progress (menggunakan logic yang sama seperti CalculateProgress di batch_service)
		var progress float64

		// Count total items (assignments + quizzes + meetings)
		var totalAssignments, totalQuizzes, totalMeetings int64
		s.db.WithContext(ctx).
			Table("assignments").
			Joins("JOIN meetings ON meetings.id = assignments.meeting_id").
			Where("meetings.batch_id = ?", batchID).
			Count(&totalAssignments)

		s.db.WithContext(ctx).
			Table("quizzes").
			Joins("JOIN meetings ON meetings.id = quizzes.meeting_id").
			Where("meetings.batch_id = ?", batchID).
			Count(&totalQuizzes)

		s.db.WithContext(ctx).
			Table("meetings").
			Where("batch_id = ?", batchID).
			Count(&totalMeetings)

		totalItems := totalAssignments + totalQuizzes + totalMeetings
		if totalItems == 0 {
			continue
		}

		// Count completed items
		var completedAssignments, completedQuizzes, completedAttendances int64
		s.db.WithContext(ctx).
			Table("assignment_submissions").
			Joins("JOIN assignments ON assignments.id = assignment_submissions.assignment_id").
			Joins("JOIN meetings ON meetings.id = assignments.meeting_id").
			Where("meetings.batch_id = ? AND assignment_submissions.user_id = ?", batchID, studentID).
			Count(&completedAssignments)

		s.db.WithContext(ctx).
			Table("quiz_results").
			Joins("JOIN quiz_attempts ON quiz_attempts.id = quiz_results.attempt_id").
			Joins("JOIN quizzes ON quizzes.id = quiz_attempts.quiz_id").
			Joins("JOIN meetings ON meetings.id = quizzes.meeting_id").
			Where("meetings.batch_id = ? AND quiz_attempts.user_id = ?", batchID, studentID).
			Count(&completedQuizzes)

		s.db.WithContext(ctx).
			Table("attendances").
			Joins("JOIN meetings ON meetings.id = attendances.meeting_id").
			Where("meetings.batch_id = ? AND attendances.user_id = ? AND attendances.is_present = ?", batchID, studentID, true).
			Count(&completedAttendances)

		completedItems := completedAssignments + completedQuizzes + completedAttendances
		progress = (float64(completedItems) / float64(totalItems)) * 100
		totalProgress += progress
	}

	// Return average
	if len(studentIDs) > 0 {
		return totalProgress / float64(len(studentIDs))
	}
	return 0
}

// getNextActivity returns the next upcoming activity for a batch
func (s *DashboardService) getNextActivity(ctx context.Context, batchSlug string) (string, string, *time.Time) {
	now := time.Now()

	// Get batch ID
	var batchID string
	err := s.db.WithContext(ctx).
		Table("batches").
		Select("id").
		Where("slug = ?", batchSlug).
		Scan(&batchID).Error

	if err != nil {
		return "", "", nil
	}

	// Cari next meeting
	type NextActivity struct {
		Type  string
		Title string
		Date  time.Time
	}

	var activities []NextActivity

	// 1. Next meeting
	var meeting NextActivity
	err = s.db.WithContext(ctx).
		Table("meetings").
		Select("'meeting' as type, title, start_at as date").
		Where("batch_id = ? AND start_at > ?", batchID, now).
		Order("start_at ASC").
		Limit(1).
		Scan(&meeting).Error
	if err == nil && meeting.Title != "" {
		activities = append(activities, meeting)
	}

	// 2. Next assignment end date
	var assignment NextActivity
	err = s.db.WithContext(ctx).
		Table("assignments").
		Select("'assignment' as type, assignments.title, assignments.end_at as date").
		Joins("JOIN meetings ON meetings.id = assignments.meeting_id").
		Where("meetings.batch_id = ? AND assignments.end_at > ?", batchID, now).
		Order("assignments.end_at ASC").
		Limit(1).
		Scan(&assignment).Error
	if err == nil && assignment.Title != "" {
		activities = append(activities, assignment)
	}

	// 3. Next quiz
	var quiz NextActivity
	err = s.db.WithContext(ctx).
		Table("quizzes").
		Select("'quiz' as type, quizzes.title, quizzes.start_time as date").
		Joins("JOIN meetings ON meetings.id = quizzes.meeting_id").
		Where("meetings.batch_id = ? AND quizzes.start_time > ?", batchID, now).
		Order("quizzes.start_time ASC").
		Limit(1).
		Scan(&quiz).Error
	if err == nil && quiz.Title != "" {
		activities = append(activities, quiz)
	}

	// Sort by date and return the earliest one
	if len(activities) == 0 {
		return "", "", nil
	}

	// Find the earliest activity
	earliest := activities[0]
	for _, act := range activities {
		if act.Date.Before(earliest.Date) {
			earliest = act
		}
	}

	return earliest.Type, earliest.Title, &earliest.Date
}

// GetTeacherWorkload returns teacher workload statistics
func (s *DashboardService) GetTeacherWorkload(ctx context.Context, period string) (*dto.TeacherWorkloadResponse, error) {
	var startDate time.Time
	now := time.Now()

	switch period {
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	default:
		startDate = now.AddDate(0, 0, -7)
		period = "week"
	}

	type TeacherData struct {
		TeacherID    string
		TeacherName  string
		MeetingCount int64
		TotalHours   float64
	}

	var teachers []TeacherData
	err := s.db.WithContext(ctx).
		Table("users").
		Select(`
			users.id as teacher_id,
			users.name as teacher_name,
			COUNT(DISTINCT meetings.id) as meeting_count,
			COALESCE(SUM(EXTRACT(EPOCH FROM (meetings.end_at - meetings.start_at)) / 3600), 0) as total_hours
		`).
		Joins("JOIN meeting_teachers ON meeting_teachers.user_id = users.id").
		Joins("JOIN meetings ON meetings.id = meeting_teachers.meeting_id").
		Where("users.role_type = ? AND meetings.start_at >= ?", "guru", startDate).
		Group("users.id, users.name").
		Order("meeting_count DESC").
		Limit(10).
		Scan(&teachers).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get teacher workload: %w", err)
	}

	var items []dto.TeacherWorkloadItem
	for _, t := range teachers {
		var pendingCount int64
		s.db.WithContext(ctx).
			Table("assignment_submissions").
			Joins("JOIN assignments ON assignments.id = assignment_submissions.assignment_id").
			Joins("JOIN meetings ON meetings.id = assignments.meeting_id").
			Joins("JOIN meeting_teachers ON meeting_teachers.meeting_id = meetings.id").
			Where("meeting_teachers.user_id = ? AND assignment_submissions.id NOT IN (SELECT assignment_submission_id FROM assignment_grades)", t.TeacherID).
			Count(&pendingCount)

		items = append(items, dto.TeacherWorkloadItem{
			TeacherID:           t.TeacherID,
			TeacherName:         t.TeacherName,
			MeetingCount:        int(t.MeetingCount),
			TotalHours:          t.TotalHours,
			PendingGradingCount: int(pendingCount),
		})
	}

	if items == nil {
		items = []dto.TeacherWorkloadItem{}
	}

	return &dto.TeacherWorkloadResponse{
		Period: period,
		Total:  len(items),
		Data:   items,
	}, nil
}

// GetCertificateStats returns certificate statistics
func (s *DashboardService) GetCertificateStats(ctx context.Context, period string) (*dto.CertificateStatsResponse, error) {
	var startDate time.Time
	now := time.Now()

	switch period {
	case "7d":
		startDate = now.AddDate(0, 0, -7)
	case "30d":
		startDate = now.AddDate(0, 0, -30)
	case "90d":
		startDate = now.AddDate(0, 0, -90)
	default:
		return nil, fmt.Errorf("invalid period: must be 7d, 30d, or 90d")
	}

	var issuedCount int64
	s.db.WithContext(ctx).
		Table("certificates").
		Where("created_at >= ?", startDate).
		Count(&issuedCount)

	return &dto.CertificateStatsResponse{
		Period:      period,
		IssuedCount: issuedCount,
	}, nil
}

// GetRecentActivities returns recent activities in the system
func (s *DashboardService) GetRecentActivities(ctx context.Context, period string, limit int) (*dto.RecentActivitiesResponse, error) {
	if limit <= 0 {
		limit = 20
	}

	// Tentukan tanggal mulai berdasarkan period
	var startDate time.Time
	now := time.Now()

	switch period {
	case "7d":
		startDate = now.AddDate(0, 0, -7)
	case "30d":
		startDate = now.AddDate(0, 0, -30)
	case "90d":
		startDate = now.AddDate(0, 0, -90)
	default:
		return nil, fmt.Errorf("invalid period: must be 7d, 30d, or 90d")
	}

	var items []dto.RecentActivityItem

	// Recent payments
	type PaymentActivity struct {
		ID         string
		AdminName  string
		BatchTitle string
		Amount     float64
		CreatedAt  time.Time
	}

	var payments []PaymentActivity
	s.db.WithContext(ctx).
		Table("purchases").
		Select(`
			purchases.id,
			'Admin' as admin_name,
			batches.title as batch_title,
			prices.price as amount,
			purchases.created_at
		`).
		Joins("LEFT JOIN batches ON batches.id = purchases.batch_id").
		Joins("LEFT JOIN prices ON prices.id = purchases.price_id").
		Where("purchases.payment_status = ? AND purchases.created_at >= ?", "paid", startDate).
		Order("purchases.created_at DESC").
		Limit(5).
		Scan(&payments)

	for _, p := range payments {
		relTime := getRelativeTime(p.CreatedAt)
		items = append(items, dto.RecentActivityItem{
			ID:           p.ID,
			Type:         "payment_verified",
			Title:        "Pembayaran terverifikasi",
			Description:  fmt.Sprintf("%s • %s • Rp %.0f", p.AdminName, p.BatchTitle, p.Amount),
			ActorName:    p.AdminName,
			RelatedTo:    p.BatchTitle,
			Amount:       &p.Amount,
			CreatedAt:    p.CreatedAt,
			RelativeTime: relTime,
		})
	}

	// Recent submissions
	type SubmissionActivity struct {
		ID          string
		UserName    string
		Assignment  string
		BatchTitle  string
		SubmittedAt time.Time
	}

	var submissions []SubmissionActivity
	s.db.WithContext(ctx).
		Table("assignment_submissions").
		Select(`
			assignment_submissions.id,
			users.name as user_name,
			assignments.title as assignment,
			batches.title as batch_title,
			assignment_submissions.created_at as submitted_at
		`).
		Joins("LEFT JOIN users ON users.id = assignment_submissions.user_id").
		Joins("LEFT JOIN assignments ON assignments.id = assignment_submissions.assignment_id").
		Joins("LEFT JOIN meetings ON meetings.id = assignments.meeting_id").
		Joins("LEFT JOIN batches ON batches.id = meetings.batch_id").
		Where("assignment_submissions.created_at >= ?", startDate).
		Order("assignment_submissions.created_at DESC").
		Limit(5).
		Scan(&submissions)

	for _, sub := range submissions {
		relTime := getRelativeTime(sub.SubmittedAt)
		items = append(items, dto.RecentActivityItem{
			ID:           sub.ID,
			Type:         "submission",
			Title:        "Pengumpulan tugas",
			Description:  fmt.Sprintf("%s • %s — %s", sub.UserName, sub.Assignment, sub.BatchTitle),
			ActorName:    sub.UserName,
			RelatedTo:    sub.BatchTitle,
			CreatedAt:    sub.SubmittedAt,
			RelativeTime: relTime,
		})
	}

	if items == nil {
		items = []dto.RecentActivityItem{}
	}

	return &dto.RecentActivitiesResponse{
		Period: period,
		Total:  len(items),
		Data:   items,
	}, nil
}

// getRelativeTime returns relative time string
func getRelativeTime(t time.Time) string {
	duration := time.Since(t)

	if duration.Minutes() < 60 {
		return fmt.Sprintf("%.0f menit lalu", duration.Minutes())
	}
	if duration.Hours() < 24 {
		return fmt.Sprintf("%.0f jam lalu", duration.Hours())
	}
	days := int(duration.Hours() / 24)
	return fmt.Sprintf("%d hari lalu", days)
}
