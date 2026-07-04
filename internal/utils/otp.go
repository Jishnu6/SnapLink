package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GenerateOTP creates a cryptographically secure 6-digit numeric string.
func GenerateOTP() (string, error) {
	// Define the maximum limit (999,999)
	max := big.NewInt(1000000)

	// Generate a cryptographically secure random number between 0 and 999,999
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	// Format the number as a 6-digit string, padding with leading zeros if necessary (e.g., "048291")
	return fmt.Sprintf("%06d", n.Int64()), nil
}
