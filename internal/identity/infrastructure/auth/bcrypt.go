package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const hashCost = 12

// BcryptHasher hashes and verifies passwords using bcrypt.
type BcryptHasher struct{}

func (BcryptHasher) Hash(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), hashCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(b), nil
}

func (BcryptHasher) Verify(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
