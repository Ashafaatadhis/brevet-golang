package services

import (
	"brevet-api/config"
	"brevet-api/helpers"
	"brevet-api/models"
	"brevet-api/repository"
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"

	"github.com/phpdave11/gofpdf/contrib/gofpdi"
	"github.com/skip2/go-qrcode"
	"gorm.io/gorm"
)

// ICertificateService interface
type ICertificateService interface {
	EnsureCertificate(ctx context.Context, batchID uuid.UUID, userID uuid.UUID) (*models.Certificate, error)
}

// CertificateService provides methods for managing courses
type CertificateService struct {
	certRepo        repository.ICertificateRepository
	userRepo        repository.IUserRepository
	batchRepo       repository.IBatchRepository
	attendanceRepo  repository.IAttendanceRepository
	meetingRepo     repository.IMeetingRepository
	purchaseService IPurchaseService
	batchService    IBatchService
	fileService     IFileService
}

// NewCertificateService creates a new instance of CertificateService
func NewCertificateService(
	certRepo repository.ICertificateRepository,
	userRepo repository.IUserRepository,
	batchRepo repository.IBatchRepository,
	attendanceRepo repository.IAttendanceRepository,
	meetingRepo repository.IMeetingRepository,
	purchaseService IPurchaseService,
	batchService IBatchService,
	fileService IFileService,
) ICertificateService {
	return &CertificateService{
		certRepo:        certRepo,
		userRepo:        userRepo,
		batchRepo:       batchRepo,
		attendanceRepo:  attendanceRepo,
		meetingRepo:     meetingRepo,
		purchaseService: purchaseService,
		batchService:    batchService,
		fileService:     fileService,
	}
}

// LearningMaterial Misal data materi sudah ada
type LearningMaterial struct {
	No       string
	Material string
}

// generatePDF membuat PDF sertifikat dari template gambar
// generatePDF membuat PDF sertifikat dari template gambar
func (s *CertificateService) generatePDF(userName string, batch *models.Batch, certNumber string, listOfMeeting []string, qrPath string) (string, error) {
	// pastikan folder certificates ada
	if _, err := os.Stat("./certificates"); os.IsNotExist(err) {
		if err := os.Mkdir("./certificates", os.ModePerm); err != nil {
			return "", fmt.Errorf("failed to create certificates directory: %w", err)
		}
	}

	templatePath := "./templates/sertifikat.pdf"
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return "", fmt.Errorf("template file not found: %s", templatePath)
	}

	// Validasi font files sebelum digunakan
	fonts := map[string]string{
		"./fonts/lucida.ttf":    "Lucida",
		"./fonts/cambriai.ttf":  "Cambria-I",
		"./fonts/cambria.ttf":   "Cambria",
		"./fonts/cambriaib.ttf": "Cambria-IB",
		"./fonts/cambriab.ttf":  "Cambria-B",
	}

	pdf := gofpdf.New("L", "mm", "A4", "")

	// Add fonts with error handling
	for fontPath, fontName := range fonts {
		if _, err := os.Stat(fontPath); os.IsNotExist(err) {
			fmt.Printf("[WARNING] Font file not found: %s, using default font\n", fontPath)
			continue
		}

		// Determine font style
		var style string
		if strings.Contains(fontName, "-IB") {
			style = "IB"
		} else if strings.Contains(fontName, "-I") {
			style = "I"
		} else if strings.Contains(fontName, "-B") {
			style = "B"
		} else {
			style = ""
		}

		// Clean font name
		cleanName := strings.Split(fontName, "-")[0]
		pdf.AddUTF8Font(cleanName, style, fontPath)
	}

	importer := gofpdi.NewImporter()

	// Validate template pages
	tpl1 := importer.ImportPage(pdf, templatePath, 1, "/MediaBox")

	pdf.AddPage()
	importer.UseImportedTemplate(pdf, tpl1, 0, 0, 297, 210)

	// Add QR code with error handling
	if _, err := os.Stat(qrPath); err == nil {
		pageW, pageH := pdf.GetPageSize()
		qrSize := 40.0
		margin := 20.0
		x := pageW - qrSize - margin
		y := pageH - qrSize - margin
		pdf.Image(qrPath, x, y, qrSize, 0, false, "", 0, "")
	} else {
		fmt.Printf("[WARNING] QR code file not found: %s\n", qrPath)
	}

	// Add text to page 1
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "", 24) // Use default Arial if Lucida fails
	pdf.SetXY(100, 95)
	pdf.CellFormat(100, 10, userName, "", 0, "C", false, 0, "")

	pdf.SetFont("Arial", "", 12)
	pdf.Ln(11)
	pdf.CellFormat(0, 10, "No. "+certNumber, "", 1, "C", false, 0, "")
	pdf.Ln(12.5)
	pdf.SetFont("Arial", "I", 12)
	pdf.CellFormat(0, 10, "Training Programs of The Applied "+batch.Course.Title, "", 1, "C", false, 0, "")
	pdf.Ln(0.2)

	// Format dates
	startDate := batch.StartAt.Format("January 02, 2006")
	endDate := batch.EndAt.Format("January 02, 2006")
	period := fmt.Sprintf("%s – %s, Gunadarma University", startDate, endDate)
	pdf.CellFormat(0, 10, period, "", 1, "C", false, 0, "")

	// Print date
	printDate := time.Now()
	day := printDate.Day()
	suffix := helpers.OrdinalSuffix(day)
	month := printDate.Format("January")
	year := printDate.Format("2006")
	formattedPrintDate := fmt.Sprintf("Jakarta, %s %d%s, %s", month, day, suffix, year)
	pdf.Ln(0.3)
	pdf.CellFormat(0, 10, formattedPrintDate, "", 1, "C", false, 0, "")

	// Try to import page 2 with error handling
	tpl2 := importer.ImportPage(pdf, templatePath, 2, "/MediaBox")
	if tpl2 == 0 {
		fmt.Println("[WARNING] Failed to import page 2 from template, creating simple page")
		// Create a simple second page instead
		pdf.AddPage()
	} else {
		// Build materials data
		materials := make([]LearningMaterial, 0, len(listOfMeeting))
		for i, m := range listOfMeeting {
			fmt.Printf("[DEBUG] Meeting %d raw: %q (len=%d)\n", i+1, m, len(m))
			materials = append(materials, LearningMaterial{
				No:       fmt.Sprintf("%d", i+1),
				Material: m,
			})
		}

		pdf.AddPage()
		importer.UseImportedTemplate(pdf, tpl2, 0, 0, 297, 210)

		// Table configuration
		startX := 25.0
		rowHeight := 8.0
		colWidthNo := 15.0
		colWidthMaterial := 150.0
		padding := 1.0
		pageW, pageH := pdf.GetPageSize()

		// Calculate header height with safety check
		headerNo := "No."
		headerMateri := "Learning Materials"

		// Safe SplitLines with validation
		var linesHeaderNo, linesHeaderMateri [][]byte

		// Use smaller width to avoid issues
		safeColWidthNo := colWidthNo - 2*padding
		safeColWidthMaterial := colWidthMaterial - 2*padding

		if safeColWidthNo > 0 {
			linesHeaderNo = pdf.SplitLines([]byte(headerNo), safeColWidthNo)
		}
		if safeColWidthMaterial > 0 {
			linesHeaderMateri = pdf.SplitLines([]byte(headerMateri), safeColWidthMaterial)
		}

		// Validate results
		if len(linesHeaderNo) == 0 {
			linesHeaderNo = [][]byte{[]byte(headerNo)}
		}
		if len(linesHeaderMateri) == 0 {
			linesHeaderMateri = [][]byte{[]byte(headerMateri)}
		}

		maxLinesHeader := len(linesHeaderNo)
		if len(linesHeaderMateri) > maxLinesHeader {
			maxLinesHeader = len(linesHeaderMateri)
		}
		cellHHeader := float64(maxLinesHeader)*rowHeight + 2*padding

		// Calculate total body height with safety
		tableBodyH := 0.0
		for i, m := range materials {
			fmt.Printf("[DEBUG] SplitLines material %d: %q (len=%d)\n", i+1, m.Material, len(m.Material))

			var linesMaterial [][]byte
			if safeColWidthMaterial > 0 && len(m.Material) > 0 {
				linesMaterial = pdf.SplitLines([]byte(m.Material), safeColWidthMaterial)
			}

			// Validate and ensure at least one line
			if len(linesMaterial) == 0 {
				linesMaterial = [][]byte{[]byte(m.Material)}
			}

			fmt.Printf("[DEBUG] lines=%d\n", len(linesMaterial))
			cellH := float64(len(linesMaterial))*rowHeight + 2*padding
			tableBodyH += cellH
		}

		// Calculate table position
		tableW := colWidthNo + colWidthMaterial
		tableH := cellHHeader + tableBodyH
		startX = (pageW - tableW) / 2
		startY := (pageH - tableH) / 2

		// Draw header
		pdf.SetFont("Arial", "B", 12)
		pdf.SetXY(startX+padding, startY+padding)
		pdf.MultiCell(colWidthNo-2*padding, rowHeight, headerNo, "1", "C", false)

		pdf.SetFont("Arial", "I", 12)
		pdf.SetXY(startX+colWidthNo+padding, startY+padding)
		pdf.MultiCell(colWidthMaterial-2*padding, rowHeight, headerMateri, "1", "C", false)

		// Draw data rows
		pdf.SetY(startY + cellHHeader)
		for i, m := range materials {
			yStart := pdf.GetY()
			fmt.Printf("[DEBUG] Render row %d: %q at Y=%.2f\n", i+1, m.Material, yStart)

			var linesMaterial [][]byte
			if safeColWidthMaterial > 0 && len(m.Material) > 0 {
				linesMaterial = pdf.SplitLines([]byte(m.Material), safeColWidthMaterial)
			}

			// Ensure at least one line
			if len(linesMaterial) == 0 {
				linesMaterial = [][]byte{[]byte(m.Material)}
			}

			fmt.Printf("[DEBUG] Row %d has %d lines\n", i+1, len(linesMaterial))
			cellH := float64(len(linesMaterial))*rowHeight + 2*padding

			// Draw cells
			pdf.SetFont("Arial", "", 12)
			pdf.SetXY(startX+padding, yStart+padding)
			pdf.CellFormat(colWidthNo-2*padding, cellH-2*padding, m.No+".", "1", 0, "C", false, 0, "")

			pdf.SetFont("Arial", "I", 12)
			pdf.SetXY(startX+colWidthNo+padding, yStart+padding)
			pdf.MultiCell(colWidthMaterial-2*padding, rowHeight, m.Material, "1", "C", false)

			pdf.SetY(yStart + cellH)
		}
	}

	// Generate PDF with error handling
	buf := bytes.NewBuffer(nil)
	if err := pdf.Output(buf); err != nil {
		return "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	fileName := fmt.Sprintf("certificate_%s_%s.pdf",
		uuid.New().String()[:8],
		time.Now().Format("20060102"))

	// Save using FileService
	publicURL, err := s.fileService.SaveGeneratedFile("certificates", fileName, buf.Bytes())
	if err != nil {
		return "", err
	}

	return publicURL, nil
}

// generateQRCode Generate QR code base64 dari teks
func (s *CertificateService) generateQRCode(certID uuid.UUID) (string, error) {
	// pastikan folder ./qrcodes ada
	err := os.MkdirAll("./qrcodes", os.ModePerm)
	if err != nil {
		return "", err
	}

	filePath := fmt.Sprintf("./qrcodes/%s.png", uuid.New().String())

	baseURL := config.GetEnv("FRONTEND_URL", "http://localhost")
	url := fmt.Sprintf("%s/certificates/%s", baseURL, certID.String())

	err = qrcode.WriteFile(url, qrcode.Medium, 256, filePath)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func (s *CertificateService) generateCertificateNumber(ctx context.Context, batchID uuid.UUID) (string, error) {
	// prefix bisa statis, misalnya kode lembaga
	prefix := "20100112"

	// ambil nomor urut terakhir untuk batch ini
	lastSeq, err := s.certRepo.GetLastSequenceByBatch(ctx, batchID)
	if err != nil {
		return "", err
	}

	// increment
	nextSeq := lastSeq + 1

	// misal batch keberapa, bisa ambil dari batch table
	batch, err := s.batchRepo.FindByID(ctx, batchID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%s %04d", prefix, batch.ID.String()[:8], nextSeq), nil
}

// EnsureCertificate memastikan sertifikat dibuat untuk user
func (s *CertificateService) EnsureCertificate(ctx context.Context, batchID uuid.UUID, userID uuid.UUID) (*models.Certificate, error) {
	fmt.Println("[DEBUG] Mulai EnsureCertificate", "batchID:", batchID, "userID:", userID)

	// bayar dulu bos
	isPaid, err := s.purchaseService.HasPaid(ctx, userID, batchID)
	fmt.Println("[DEBUG] isPaid:", isPaid, "err:", err)
	if err != nil {
		return nil, err
	}
	if !isPaid {
		return nil, fmt.Errorf("forbidden")
	}

	// 1. cek progress assignment & quiz
	progress, err := s.batchService.CalculateProgress(ctx, batchID, userID)
	fmt.Println("[DEBUG] progress:", progress, "err:", err)
	if err != nil {
		return nil, err
	}
	if progress < 100 {
		return nil, fmt.Errorf("progress belum 100%%")
	}

	// 2. cek attendance
	totalMeetings, err := s.batchRepo.CountMeetings(ctx, batchID)
	fmt.Println("[DEBUG] totalMeetings:", totalMeetings, "err:", err)
	if err != nil {
		return nil, err
	}
	attendedMeetings, err := s.attendanceRepo.CountByBatchUser(ctx, batchID, userID)
	fmt.Println("[DEBUG] attendedMeetings:", attendedMeetings, "err:", err)
	if err != nil {
		return nil, err
	}
	if totalMeetings > 0 && attendedMeetings < totalMeetings {
		return nil, fmt.Errorf("murid belum hadir di semua pertemuan (%d/%d)", attendedMeetings, totalMeetings)
	}

	// cek sertifikat existing
	_, err = s.certRepo.GetByBatchUser(ctx, batchID, userID)
	fmt.Println("[DEBUG] GetByBatchUser err:", err)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if err == nil {
		return nil, fmt.Errorf("sertifikat sudah ada untuk user %s di batch %s", userID, batchID)
	}

	certNumber, err := s.generateCertificateNumber(ctx, batchID)
	fmt.Println("[DEBUG] certNumber:", certNumber, "err:", err)
	if err != nil {
		return nil, err
	}

	cert := &models.Certificate{
		BatchID: batchID,
		Number:  certNumber,
		UserID:  userID,
	}
	err = s.certRepo.Create(ctx, cert)
	fmt.Println("[DEBUG] Insert Certificate err:", err, "certID:", cert.ID)
	if err != nil {
		return nil, err
	}

	qrPath, err := s.generateQRCode(cert.ID)
	fmt.Println("[DEBUG] QR generated at:", qrPath, "err:", err)
	if err != nil {
		return nil, err
	}
	defer os.Remove(qrPath)

	user, err := s.userRepo.FindByID(ctx, userID)
	fmt.Println("[DEBUG] user:", user, "err:", err)
	if err != nil {
		return nil, err
	}

	batch, err := s.batchRepo.GetBatchWithCourse(ctx, batchID)
	fmt.Println("[DEBUG] batch:", batch, "err:", err)
	if err != nil || batch == nil {
		return nil, err
	}

	listOfMeeting, err := s.meetingRepo.GetMeetingNamesByBatchID(ctx, batch.ID)
	fmt.Println("[DEBUG] listOfMeeting:", listOfMeeting, "err:", err)
	if err != nil {
		return nil, err
	}

	pdfURL, err := s.generatePDF(user.Name, batch, certNumber, listOfMeeting, qrPath)
	fmt.Println("[DEBUG] pdfURL:", pdfURL, "err:", err)
	if err != nil {
		return nil, err
	}

	cert.URL = pdfURL
	cert.QRCode = fmt.Sprintf("/certificates/%s", cert.ID)

	err = s.certRepo.Update(ctx, cert)
	fmt.Println("[DEBUG] Update certificate err:", err)
	if err != nil {
		return nil, err
	}

	fmt.Println("[DEBUG] Sukses generate certificate", "certID:", cert.ID, "URL:", cert.URL)
	return cert, nil
}
