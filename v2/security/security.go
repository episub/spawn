package security

import (
	"context"
	"crypto/rand"
	"errors"
	"io"

	opentracing "github.com/opentracing/opentracing-go"

	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost = 12
	saltBytes  = 32
)

// ErrMismatchedHashAndPassword Provided password does not match the hashed password
var ErrMismatchedHashAndPassword = errors.New("Invalid password")

// NewSalt Gets a new salt
func NewSalt() ([]byte, error) {
	salt := make([]byte, saltBytes)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		return salt, err
	}

	return salt, nil
}

// HashPassword Returns password hash given password and salt
func HashPassword(password []byte, salt []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(combinePasswordAndSalt(password, salt), bcryptCost)
}

func combinePasswordAndSalt(password []byte, salt []byte) []byte {
	return []byte(append(salt, password...))
}

// AuthenticateUser Returns nil if user has provided password, otherwise error
func AuthenticateUser(ctx context.Context, salt []byte, hashedPassword []byte, providedPassword []byte) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "AuthenticateUser")
	defer span.Finish()

	pwd := combinePasswordAndSalt(providedPassword, salt)

	err := bcrypt.CompareHashAndPassword(hashedPassword, pwd)

	// Use our own customised error so that we are not bound to bcrypt for statements about password validity
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return ErrMismatchedHashAndPassword
	}

	return err
}
