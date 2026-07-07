package utils

import (
	"fmt"
	"os"

	"github.com/resend/resend-go/v2"
)

// SendOTPEmail is SYNCHRONOUS. It blocks until Resend accepts the email.
// This is required for Vercel/Serverless deployments.
func SendOTPEmail(toEmail, otp string) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	fromEmail := os.Getenv("RESEND_FROM_EMAIL")

	if fromEmail == "" {
		fromEmail = "onboarding@resend.dev"
	}

	client := resend.NewClient(apiKey)

	params := &resend.SendEmailRequest{
		From:    fromEmail,
		To:      []string{toEmail},
		Subject: "Your URL Shortener Login Code",
		Text:    fmt.Sprintf("Your 6-digit verification code is: %s\n\nThis code will expire in 5 minutes.", otp),
	}

	// Wait for the Resend API to finish
	_, err := client.Emails.Send(params)

	// Return the error to the caller
	return err
}
