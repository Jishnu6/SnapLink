package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/resend/resend-go/v2"
)

// SendOTPEmail triggers an email send in the background using Resend and returns immediately.
func SendOTPEmail(toEmail, otp string) {
	// Spin up a background goroutine
	go func(recipient, code string) {
		apiKey := os.Getenv("RESEND_API_KEY")
		fromEmail := os.Getenv("RESEND_FROM_EMAIL")

		// Fallback to Resend's testing email if you haven't set your own domain yet
		if fromEmail == "" {
			fromEmail = "onboarding@resend.dev"
		}

		client := resend.NewClient(apiKey)

		// Format the email using the Resend struct
		params := &resend.SendEmailRequest{
			From:    fromEmail,
			To:      []string{recipient},
			Subject: "Your URL Shortener Login Code",
			Text:    fmt.Sprintf("Your 6-digit verification code is: %s\n\nThis code will expire in 5 minutes.", code),
		}

		// Send the email via the API
		_, err := client.Emails.Send(params)

		// Log the error since we are running in a background thread
		if err != nil {
			log.Printf("[Email Error] Failed to send OTP to %s: %v\n", recipient, err)
		}
	}(toEmail, otp)
}
