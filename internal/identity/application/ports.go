package application

import "time"

// TokenIssuer is the port for creating authentication tokens.
// Implemented by infrastructure/auth.JWTService.
type TokenIssuer interface {
	Issue(inspectorID, companyID, role string) (token string, expiresAt time.Time, err error)
}

// PasswordHasher is the port for password hashing and verification.
// Implemented by infrastructure/auth.BcryptHasher.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hash, password string) bool
}
