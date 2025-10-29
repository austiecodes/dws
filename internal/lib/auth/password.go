package auth

import "golang.org/x/crypto/bcrypt"

const passwordCost = bcrypt.DefaultCost

// HashPassword generates a bcrypt hash for the provided password string.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), passwordCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ComparePassword checks whether the supplied plaintext password matches the stored hash.
func ComparePassword(hash, candidate string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(candidate)) == nil
}
