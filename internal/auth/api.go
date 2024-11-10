package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	header := headers.Get("Authorization")
	if header == "" {
		return "", errors.New("invalid header")
	}

	apiKey_check := strings.HasPrefix(header, "ApiKey ")
	if !apiKey_check {
		return "", errors.New("invalid header (label)")
	}

	apiKey := strings.TrimPrefix(header, "ApiKey ")
	return apiKey, nil
}
