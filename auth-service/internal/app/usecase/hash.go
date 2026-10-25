package usecase

import (
	"crypto/sha256"
	"encoding/hex"
)

// hashRefreshToken — SHA-256 от refresh-токена
func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
