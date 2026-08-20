package hash

import "golang.org/x/crypto/bcrypt"

// Password wraps bcrypt hashing with the project cost factor (12).
const Cost = 12

// HashPassword returns a bcrypt hash of the plaintext password.
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), Cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ComparePassword verifies a plaintext against a stored hash.
func ComparePassword(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}

// TokenHash is a lightweight opaque hash for refresh tokens (SHA256-ish via bcrypt is overkill).
// We reuse bcrypt at lower cost for simplicity in this MVP.
func TokenHash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func CompareTokenHash(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
