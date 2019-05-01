package security

import (
	"bytes"
	"context"
	"testing"

	zxcvbn "github.com/nbutton23/zxcvbn-go"

	"golang.org/x/crypto/bcrypt"
)

var testPasswords = []struct {
	Hash     []byte
	Password []byte
	Salt     []byte
	Success  bool
}{
	{
		Hash:     []byte("$2a$12$r2wy7p2eir7OqVW.eif.Cu/0P6Ysg1wzBEQQyWRilN0Kb9/QENGGy"),
		Password: []byte(""),
		Salt:     []byte(""),
		Success:  true,
	},
	{
		Hash:     []byte("$2a$12$r2wy8p2eir7OqVW.eif.Cu/0P6Ysg1wzBEQQyWRilN0Kb9/QENGGy"),
		Password: []byte(""),
		Salt:     []byte(""),
		Success:  false,
	},
	{
		Hash:     []byte("$2a$12$ekk6GLEiBgqYeG6AQji.5eD9lyVn5DVooN5EgFdk8I/7iC7AEsnaG"),
		Password: []byte("1234"),
		Salt:     []byte(""),
		Success:  true,
	},
	{
		Hash:     []byte("$2a$12$ekk6GLEiBgqYeG6AQji.5eD9lyVn5DVooN5EgFdk8I/7iC7AEsnaG"),
		Password: []byte("12345"),
		Salt:     []byte(""),
		Success:  false,
	},
	{
		Hash:     []byte("$2a$12$hoj7ixM7MHk/vtByOFX0/.4CbDPcmsIx2hMRfQZFLyCsTezmZWVlG"),
		Password: []byte("test"),
		Salt:     []byte(""),
		Success:  true,
	},
	{
		Hash:     []byte("invalid"),
		Password: []byte("test"),
		Salt:     []byte(""),
		Success:  false,
	},
	{
		Hash:     []byte("$2a$12$bf59wHyny.Wb1EAJ.P5wGeh4zvdSn4p/xAzsTUM6OlTtLzDNJlcG2"),
		Password: []byte("1234"),
		Salt:     []byte("2"),
		Success:  true,
	},
}

func TestCompareHashAndPassword(t *testing.T) {
	for i, password := range testPasswords {
		pwd := combinePasswordAndSalt(password.Password, password.Salt)
		err := bcrypt.CompareHashAndPassword(password.Hash, pwd)
		if (err == nil) != password.Success {
			t.Errorf("%d: Expected success: %t.  Was error empty: %t.  Password: %+v", i, password.Success, (err == nil), password)
		}
	}
}

func TestAuthenticateUser(t *testing.T) {
	for i, password := range testPasswords {
		err := AuthenticateUser(context.Background(), password.Salt, password.Hash, password.Password)
		if (err == nil) != password.Success {
			t.Errorf("%d: Expected success: %t.  Was error empty: %t.  Password: %+v", i, password.Success, (err == nil), password)
		}
	}
}

func TestCombinePasswordAndHash(t *testing.T) {
	salt := []byte("a")
	password := []byte("bc")

	combined := combinePasswordAndSalt(password, salt)

	if bytes.Compare(combined, []byte(append(salt, password...))) != 0 {
		t.Errorf("Unexpected combined salt and hash")
	}
}

func TestNewSalt(t *testing.T) {
	ns, err := NewSalt()

	if err != nil {
		t.Error(err)
		return
	}

	if len(ns) != saltBytes {
		t.Errorf("Expected salt of length %d, but was %d", saltBytes, len(ns))
	}

	// Crude measure of entropy of the string to ensure we have some decent random results:
	strength := zxcvbn.PasswordStrength(string(ns), []string{})

	if strength.Score < 4 {
		t.Errorf("Salt (%x) unexpectedly weak: %d", string(ns), strength.Score)
	}
}

func TestHashPassword(t *testing.T) {
	// Check we can go both ways with a password:

	password := []byte("1234")
	salt, err := NewSalt()

	if err != nil {
		t.Error(err)
		return
	}

	hashed, err := HashPassword(password, salt)

	if err != nil {
		t.Error(err)
		return
	}

	// Try and authenticate with password and hash:
	err = bcrypt.CompareHashAndPassword(hashed, combinePasswordAndSalt(password, salt))

	if err != nil {
		t.Error(err)
		return
	}

	// Check it fails for wrong password:
	err = bcrypt.CompareHashAndPassword(hashed, []byte("wrong"))

	if err == nil {
		t.Errorf("Expected failure, but password marked as valid")
		return
	}

}
