package utils

import (
	"math/rand"
	"strings"
)

func generateRandomString(charset string, n uint) string {
	var sb strings.Builder
	for i := uint(0); i < n; i++ {
		randomIndex := rand.Intn(len(charset))
		sb.WriteByte(charset[randomIndex])
	}
	return sb.String()
}

func GenerateLowercase(n uint) string {
	const charset = "abcdefghijklmnopqrstuvwxyz"
	return generateRandomString(charset, n)
}

func GenerateUppercase(n uint) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return generateRandomString(charset, n)
}

func GenerateAlphanumericUppercase(n uint) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	return generateRandomString(charset, n)
}

func GenerateAlphanumeric(n uint) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	return generateRandomString(charset, n)
}

func GenerateNumeric(n uint) string {
	const charset = "0123456789"
	return generateRandomString(charset, n)
}
