package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendOTPEmail(toEmail, otp string) error {
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_APP_PASSWORD") // Use a Google App Password
	smtpHost := "smtp-relay.brevo.com"
	smtpPort := "587"

	// Format the email message
	message := []byte(fmt.Sprintf("Subject: Your URL Shortener Login Code\n\nYour 6-digit verification code is: %s\n\nThis code will expire in 5 minutes.", otp))

	// Authenticate and send
	auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{toEmail}, message)
	
	return err
}