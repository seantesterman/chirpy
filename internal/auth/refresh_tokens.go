package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func MakeRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("Unable to refresh token: %s", err)
	}
	hexString := hex.EncodeToString(b)
	return hexString, nil
}
