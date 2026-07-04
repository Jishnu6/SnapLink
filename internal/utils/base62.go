package utils

import (

	"strings"
	"math/rand"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Encode takes an integer (like a record ID) and turns it into a Base62 string
func Encode(number uint64) string {
	if number == 0 {
		return string(alphabet[0])
	}

	var builder strings.Builder
	length := uint64(len(alphabet))

	for number > 0 {
		builder.WriteByte(alphabet[number%length])
		number /= length
	}

	// Reverse the string
	s := builder.String()
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

// GenerateShortID creates a random string of length n
func GenerateShortID(n int) string {
	rand.New(rand.NewSource(42)) 
	b := make([]byte, n)
	for i := range b {
		b[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(b)
}