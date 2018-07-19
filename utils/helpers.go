package utils

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// IsValidUUID is a function that returns true if the input is a valid uuid.
// And false if the input isn't a valid uuid
// https://stackoverflow.com/a/46315070
func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

// HashPassword hashes the given password with bcrypt
func HashPassword(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes)
}

// CheckPasswordHash checks a plaintext password with a hashed password
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
