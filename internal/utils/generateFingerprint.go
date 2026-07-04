package utils

import (
	"crypto/sha256"
	"fmt"
)

// GenerateFingerprint creates a secure, anonymous hash unique to a specific browser/IP combo
func GenerateFingerprint(ip, userAgent string) string {
	input := fmt.Sprintf("%s-%s", ip, userAgent)
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash)
}