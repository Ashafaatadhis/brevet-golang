package dto

import "time"

// DashboardAdminResponse represents the admin dashboard statistics
type DashboardAdminResponse struct {
	TotalRevenue       int64   `json:"total_revenue"`        // Total pendapatan dalam periode
	ActiveParticipants int64   `json:"active_participants"`  // Peserta aktif (yang sudah bayar)
	ActiveBatches      int64   `json:"active_batches"`       // Jumlah batch aktif
	NewPurchases       int64   `json:"new_purchases"`        // Pembelian baru dalam periode
	CompletionRate     float64 `json:"completion_rate"`      // Tingkat penyelesaian (%)
	TotalCertificates  int64   `json:"total_certificates"`   // Total sertifikat yang diterbitkan
	Period             string  `json:"period"`               // Periode filter (7d, 30d, 90d)
}

// RevenueChartDataPoint represents a single data point in the revenue chart
type RevenueChartDataPoint struct {
	Date    string  `json:"date"`    // Format: "2024-10-26" atau "26 Okt"
	Revenue float64 `json:"revenue"` // Total pendapatan pada hari tersebut
}

// RevenueChartResponse represents the revenue chart data
type RevenueChartResponse struct {
	Period string                  `json:"period"` // Periode filter (7d, 30d, 90d)
	Data   []RevenueChartDataPoint `json:"data"`   // Array data per hari
}

// PendingPaymentItem represents a purchase that needs verification
type PendingPaymentItem struct {
	PurchaseID         string  `json:"purchase_id"`
	InvoiceNumber      int     `json:"invoice_number"`
	UserName           string  `json:"user_name"`
	UserEmail          string  `json:"user_email"`
	BatchSlug          string  `json:"batch_slug"`
	BatchTitle         string  `json:"batch_title"`
	Amount             float64 `json:"amount"`
	PaymentStatus      string  `json:"payment_status"` // pending, failed
	PaymentProof       *string `json:"payment_proof"`
	TransferAmount     float64 `json:"transfer_amount"`
	BankAccountName    *string `json:"bank_account_name"`
	BankAccountNumber  *string `json:"bank_account_number"`
	CreatedAt          time.Time `json:"created_at"`
}

// PendingPaymentsResponse represents list of pending payments
type PendingPaymentsResponse struct {
	Total int                  `json:"total"`
	Data  []PendingPaymentItem `json:"data"`
}

// BatchProgressItem represents batch with progress and next activity
type BatchProgressItem struct {
	BatchSlug       string  `json:"batch_slug"`
	BatchTitle      string  `json:"batch_title"`
	CourseTitle     string  `json:"course_title"`
	Quota           int     `json:"quota"`
	Enrolled        int     `json:"enrolled"`
	AvgProgress     float64 `json:"avg_progress"`      // Rata-rata progress siswa
	NextActivityType string `json:"next_activity_type"` // meeting, assignment, quiz
	NextActivityTitle string `json:"next_activity_title"`
	NextActivityDate *time.Time `json:"next_activity_date"`
}

// BatchProgressResponse represents list of batches with progress
type BatchProgressResponse struct {
	Total int                 `json:"total"`
	Data  []BatchProgressItem `json:"data"`
}

// TeacherWorkloadItem represents teacher workload statistics
type TeacherWorkloadItem struct {
	TeacherID           string `json:"teacher_id"`
	TeacherName         string `json:"teacher_name"`
	MeetingCount        int    `json:"meeting_count"`         // Jumlah pertemuan minggu ini
	TotalHours          float64 `json:"total_hours"`          // Total jam mengajar
	PendingGradingCount int    `json:"pending_grading_count"` // Tugas yang perlu dinilai
}

// TeacherWorkloadResponse represents list of teachers with workload
type TeacherWorkloadResponse struct {
	Period string                `json:"period"` // week, month
	Total  int                   `json:"total"`
	Data   []TeacherWorkloadItem `json:"data"`
}

// CertificateStatsResponse represents certificate statistics
type CertificateStatsResponse struct {
	Period            string `json:"period"` // 7d, 30d, 90d
	IssuedCount       int64  `json:"issued_count"`        // Sertifikat diterbitkan
}

// RecentActivityItem represents a recent activity in the system
type RecentActivityItem struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"` // payment_verified, submission, material_uploaded, registration, quiz_graded
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	ActorName    string    `json:"actor_name"`
	RelatedTo    string    `json:"related_to"`     // batch slug, meeting title, etc
	Amount       *float64  `json:"amount,omitempty"` // untuk payment
	Status       *string   `json:"status,omitempty"` // untuk payment
	CreatedAt    time.Time `json:"created_at"`
	RelativeTime string    `json:"relative_time"` // "58 menit lalu", "3 jam lalu"
}

// RecentActivitiesResponse represents list of recent activities
type RecentActivitiesResponse struct {
	Period string               `json:"period"` // 7d, 30d, 90d
	Total  int                  `json:"total"`
	Data   []RecentActivityItem `json:"data"`
}
