package validate

import (
	"fmt"
	"unicode"
)

// PasswordValidator Provides some base regexps for enforcing password policy
// https://stackoverflow.com/questions/19605150/regex-for-password-must-contain-at-least-eight-characters-at-least-one-number-a
type PasswordValidator struct {
	Failure  string
	Validate func(string, uint) bool
}

var (
	// PasswordUpperLowerNumber Ensures password contains at least one upper,
	// lower, and number
	PasswordUpperLowerNumber = PasswordValidator{
		Failure:  "Password must contain at least one upper case character, one lower, and one number, with a minium length of %d",
		Validate: passwordUpperLowerNumber,
	}
)

// Password Checks the password for the given validator
func Password(
	errors map[string][]string,
	password string,
	minLength uint,
	validator PasswordValidator,
	field string,
) bool {
	ok := validator.Validate(password, minLength)
	if !ok {
		return Fail(field, errors, fmt.Sprintf(validator.Failure, minLength))
	}
	return ok
}

func passwordUpperLowerNumber(password string, length uint) bool {
	number, upper, _ := passwordFingerprint(password)

	if uint(len(password)) < length {
		return false
	}

	if !number || !upper {
		return false
	}

	return true
}

func passwordFingerprint(password string) (number, upper, special bool) {
	letters := 0
	for _, c := range password {
		switch {
		case unicode.IsNumber(c):
			number = true
		case unicode.IsUpper(c):
			upper = true
			letters++
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			special = true
		case unicode.IsLetter(c) || c == ' ':
			letters++
		default:
			//return false, false, false, false
		}
	}

	return
}
