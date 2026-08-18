package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// generateOpaqueToken returns a random, URL-safe refresh token. Only its
// hash (see hashToken) is ever stored, so a database read alone can't be
// used to authenticate as the user.
func generateOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
