package utils

import (
	"brevet-api/config"
	"fmt"
	"net/mail"
	"strconv"
	"time"

	"gopkg.in/gomail.v2"
)

// SendVerificationEmail mengirim email verifikasi ke pengguna
func SendVerificationEmail(toEmail string, code string, token string) error {
	frontendURL := config.GetEnv("FRONTEND_URL", "http://localhost:3000")
	verificationURL := fmt.Sprintf("%s/auth/verify?token=%s", frontendURL, token)

	codeHTML := ""
	for _, digit := range code {
		codeHTML += fmt.Sprintf(`<div style="display:inline-block;width:40px;height:50px;line-height:50px;margin:0 5px;text-align:center;font-size:24px;border:2px solid #333;border-radius:8px;">%c</div>`, digit)
	}

	subject := "Verifikasi Akun Anda"
	message := fmt.Sprintf(`
		<div style="font-family:Arial,sans-serif;max-width:600px;margin:auto;padding:20px;border:1px solid #eee;border-radius:10px;">
			<h2 style="text-align:center;color:#4A90E2;">Verifikasi Akun Anda</h2>
			<p>Halo,</p>
			<p>Silakan masukkan kode verifikasi berikut di aplikasi/web kami untuk melanjutkan:</p>
			<div style="text-align:center;margin:30px 0;">%s</div>
			
			<p style="text-align:center;">atau klik tombol di bawah ini:</p>
			<div style="text-align:center;margin:20px 0;">
				<a href="%s" style="display:inline-block;padding:12px 24px;background-color:#4A90E2;color:#fff;text-decoration:none;border-radius:6px;font-size:16px;">Verifikasi Sekarang</a>
			</div>
			<p>Jika tombol tidak bisa diklik, salin link ini ke browser:</p>
			<p><a href="%s">%s</a></p>

			<p>Jika Anda tidak merasa melakukan permintaan ini, abaikan email ini.</p>
			<hr style="margin:40px 0;border:none;border-top:1px solid #ccc;" />
			<footer style="text-align:center;color:#888;font-size:12px;">
				Tax Center Gunadarma<br/>
				© %d All rights reserved.
			</footer>
		</div>
	`, codeHTML, verificationURL, verificationURL, verificationURL, time.Now().Year())

	return sendEmail(toEmail, subject, message)
}

func sendEmail(emailuser string, subject string, message string) error {
	// Validasi email tujuan
	if _, err := mail.ParseAddress(emailuser); err != nil {
		return fmt.Errorf("invalid email address: %w", err)
	}

	// Ambil konfigurasi dari environment variable dengan default
	smtpHost := config.GetEnv("SMTP_HOST", "smtp.gmail.com")
	smtpPortStr := config.GetEnv("SMTP_PORT", "587")
	smtpUser := config.GetEnv("SMTP_USER", "")
	smtpPass := config.GetEnv("SMTP_PASS", "")

	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		return fmt.Errorf("invalid SMTP_PORT: %w", err)
	}

	// Membuat pesan email
	m := gomail.NewMessage()
	m.SetHeader("From", smtpUser, "Tax Center Gunadarma")
	m.SetHeader("To", emailuser)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", message)

	// Mengirim email menggunakan SMTP
	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	d.TLSConfig = nil // Gunakan TLS

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
