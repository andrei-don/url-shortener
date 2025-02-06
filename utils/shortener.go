package utils

import (
	"crypto/sha256"
	"encoding/base64"
)

func GenerateShortCode(url string) string {
	hash := sha256.Sum256([]byte(url))
	shortCode := base64.URLEncoding.EncodeToString(hash[:])[:6]
	return shortCode
}
