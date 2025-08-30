package services

import (
	"brevet-api/utils"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

// IEmailService interface
type IEmailService interface {
	SendWithAttachment(to, subject, body, attachmentPath string) error
	SendVerificationEmail(email, code, token string) error
}

// EmailService for struct
type EmailService struct {
	SMTPHost string
	SMTPPort int
	Username string
	Password string
	From     string
}

// NewEmailServiceFromEnv is init
func NewEmailServiceFromEnv() (IEmailService, error) {
	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")

	if host == "" || portStr == "" || user == "" || pass == "" {
		return nil, fmt.Errorf("missing SMTP configuration in environment variables")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SMTP_PORT value: %w", err)
	}

	return &EmailService{
		SMTPHost: host,
		SMTPPort: port,
		Username: user,
		Password: pass,
		From:     user,
	}, nil
}

// SendWithAttachment mengirim email dengan attachment file
func (s *EmailService) SendWithAttachment(to, subject, body, attachmentPath string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)
	m.Attach(attachmentPath)

	d := gomail.NewDialer(s.SMTPHost, s.SMTPPort, s.Username, s.Password)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

// SendVerificationEmail send
func (s *EmailService) SendVerificationEmail(email, code, token string) error {
	go utils.SendVerificationEmail(email, code, token)
	return nil
}
