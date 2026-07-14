package hash

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// Hashing (One-way): You run a password through a mathematical function that scrambles it into a fixed-length string of gibberish. There is no key to reverse this process.

const cost = 12
const maxLength = 72

func HashPassword(password string) (string, error) {
	pwbytes := []byte(password)
	if len(pwbytes) > maxLength {
		return "", errors.New("invalid password lenght")
	}

	hashed, err := bcrypt.GenerateFromPassword(pwbytes, cost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func ValidatePassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
