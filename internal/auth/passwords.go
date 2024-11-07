package auth

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	encrypted_bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	encrypted_string := string(encrypted_bytes)
	return encrypted_string, nil
}

func CheckPasswordHash(password, hash string) error {
	bytePassword := []byte(password)
	byteHash := []byte(hash)
	err := bcrypt.CompareHashAndPassword(byteHash, bytePassword)
	return err
}
