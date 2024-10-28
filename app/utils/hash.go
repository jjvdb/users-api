package utils

import (
	"github.com/alexedwards/argon2id"
)

func HashPassword(s string) string {
	hash, _ := argon2id.CreateHash(s, argon2id.DefaultParams)
	return hash
}

func CheckPassword(password string, hash string) bool {
	ans, _ := argon2id.ComparePasswordAndHash(password, hash)
	return ans
}
